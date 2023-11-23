package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/model"
)

type Scheduler struct {
	Database *database.Database
	// TODO improve queue by abstracting the tasks being processed (https://mrkaran.dev/posts/job-queue-golang/)
	PodQueue        chan model.Pod
	queueOnline     bool
	PodNameByWorker map[string]string
	NextWorker      int
	Client          *http.Client

	shutdown chan struct{}
}

func (s *Scheduler) Start() {
	s.queueOnline = true

	go s.schedulePods()
	s.updatePods(1000)
}

func (s *Scheduler) EnqueuePod(p model.Pod) error {
	if !s.queueOnline {
		return fmt.Errorf("queue offline")
	}

	select {
	case s.PodQueue <- p:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (s *Scheduler) SendPodForDeletion(podName string) (int, string) {
	worker := s.PodNameByWorker[podName]
	client := http.Client{}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/pods/%s", worker, podName), nil)
	if err != nil {
		msg := "error creating DELETE request"
		log.Print(msg, err)
		return http.StatusInternalServerError, ""
	}

	resp, err := client.Do(req)
	if err != nil {
		msg := "cannot connect to worker %s"
		log.Printf(msg, worker, err)
		return http.StatusServiceUnavailable, fmt.Sprintf(msg, worker)
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print("error reading request body")
		return http.StatusInternalServerError, ""
	}

	return statusCode, string(body)
}

func (s *Scheduler) schedulePods() {
	for {
		select {
		case p := <-s.PodQueue:
			go s.schedulePod(p)
		case <-s.shutdown:
			// CHECKME should channel be "flushed"?
			s.queueOnline = false
			return
		}
	}
}

func (s *Scheduler) schedulePod(p model.Pod) {
	w := s.selectWorker()
	pBytes, _ := json.Marshal(p)

	resp, err := s.Client.Post(fmt.Sprintf("%s/pods", w), "application/json", bytes.NewBuffer(pBytes))
	if err != nil {
		log.Printf("error connecting to worker %s", w)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("error sending pod to worker %s", w)
		// TODO add max retries so pod doesn't get requeued indefinitely
		err := s.EnqueuePod(p)
		if err != nil {
			log.Printf("error requeuing pod %s: %v", p.Name, err)
		}
		return
	}

	s.PodNameByWorker[p.Name] = w
}

// CHECKME should this be part of the pod controller?
func (s *Scheduler) updatePods(intervalMillis uint) {
	for {
		var wg sync.WaitGroup

		// fetch pods from all workers asynchronously
		for _, w := range s.Database.GetWorkers() {
			wg.Add(1)

			go func(worker string) {
				defer wg.Done()

				resp, err := http.Get(fmt.Sprintf("%s/pods", worker))
				if err != nil {
					log.Printf("error connecting to %s: %v", worker, err)
					return
				}

				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Printf("error getting pods from %s: %v", worker, err)
					return
				}

				pods := []model.Pod{}
				json.NewDecoder(resp.Body).Decode(&pods)

				for _, p := range pods {
					switch p.State {
					case model.Stopped:
						// TODO delete s.PodNameByWorker as well
						s.Database.DeletePod(p.Name)
					default:
						s.Database.AddPod(p)
					}
				}
			}(w)
		}

		wg.Wait()
		time.Sleep(time.Millisecond * time.Duration(intervalMillis))
	}
}

// TODO Need algorithm for selecting workers. This is a first pass round-robin approach.
func (s *Scheduler) selectWorker() string {
	if s.NextWorker == len(s.Database.GetWorkers())-1 {
		s.NextWorker = 0
	}

	return s.Database.GetWorkers()[s.NextWorker]
}

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
	"github.com/golang-collections/collections/queue"
)

type Scheduler struct {
	PodQueue        *queue.Queue
	PodNameByWorker map[string]string
	Workers         []string // addresses
	NextWorker      int
	Database        *database.Database
}

func (s *Scheduler) Start() {
	for {
		s.UpdatePods()
		s.SchedulePods()
		time.Sleep(time.Millisecond * 1000)
	}
}

func (s *Scheduler) EnqueuePod(p model.Pod) {
	s.PodQueue.Enqueue(p)
}

func (s *Scheduler) SendPodForDeletion(podName string) (int, string) {
	worker := s.PodNameByWorker[podName]
	client := http.Client{}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/pods/%s", worker, podName), nil)
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

// TODO Need algorithm for selecting workers. This is a first pass round-robin approach.
func (s *Scheduler) selectWorker() string {
	if s.NextWorker == len(s.Workers)-1 {
		s.NextWorker = 0
	}

	return s.Workers[s.NextWorker]
}

func (s *Scheduler) SchedulePods() {
	for s.PodQueue.Len() > 0 {
		pi := s.PodQueue.Dequeue()
		p := pi.(model.Pod)
		w := s.selectWorker()
		pBytes, _ := json.Marshal(p)

		resp, err := http.Post(fmt.Sprintf("%s/pods", w), "application/json", bytes.NewBuffer(pBytes))
		if err != nil {
			log.Printf("error connecting to worker %s", w)
			continue
		}

		s.PodNameByWorker[p.Name] = w
		resp.Body.Close()
	}
}

// CHECKME should this be part of the pod controller?
func (s *Scheduler) UpdatePods() {
	var wg sync.WaitGroup

	// fetch pods from all workers asynchronously
	for _, w := range s.Workers {
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

}

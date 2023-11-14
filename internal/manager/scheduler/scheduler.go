package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/internal/model"
	"github.com/golang-collections/collections/queue"
)

type Scheduler struct {
	PodQueue        *queue.Queue
	PodNameByWorker map[string]string
	Workers         []string // addresses
	NextWorker      int
}

func (s *Scheduler) Start() {
	for {
		s.UpdateDatabase()
		s.SchedulePods()
		time.Sleep(time.Millisecond * 500)
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

// TODO Fetch pods from workers to update database state
func (s *Scheduler) UpdateDatabase() {

}

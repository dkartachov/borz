package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/pkg/model"
)

type Scheduler struct {
	apiServerAddr string
	client        *http.Client
	nextWorker    int
}

func NewScheduler(APIServerAddr string) *Scheduler {
	return &Scheduler{
		apiServerAddr: APIServerAddr,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		nextWorker: 0,
	}
}

func (s *Scheduler) Start() {
	for {
		// TODO add filter param to API Server GET /pods endpoint to only get unscheduled pods
		url := fmt.Sprintf("%s/pods", s.apiServerAddr)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Print(err)
			continue
		}

		resp, err := s.client.Do(req)
		if err != nil {
			log.Print(err)
			continue
		}

		var pods []model.Pod

		err = json.NewDecoder(resp.Body).Decode(&pods)
		if err != nil {
			log.Print(err)
			continue
		}

		var unscheduledPods []model.Pod

		for _, p := range pods {
			if p.Worker == "" {
				unscheduledPods = append(unscheduledPods, p)
			}
		}

		if len(unscheduledPods) > 0 {
			log.Print("found unscheduled pods")
			s.schedulePods(unscheduledPods)
		}

		resp.Body.Close()
		time.Sleep(10 * time.Second)
	}
}

// TODO Need algorithm for selecting workers. This is a first pass round-robin approach.
func (s *Scheduler) selectWorker() (*model.Worker, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/workers", s.apiServerAddr), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var workers []model.Worker

	err = json.NewDecoder(resp.Body).Decode(&workers)
	if err != nil {
		return nil, err
	}

	nextWorker := s.nextWorker

	if s.nextWorker == len(workers)-1 {
		s.nextWorker = 0
	} else {
		s.nextWorker += 1
	}

	return &workers[nextWorker], nil
}

func (s *Scheduler) schedulePod(p model.Pod) error {
	worker, err := s.selectWorker()
	if err != nil {
		return fmt.Errorf("error selecting worker: %v", err)
	}

	payload, err := json.Marshal(map[string]string{
		"state":  string(model.Scheduled),
		"worker": worker.Addr,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/pods/%s", s.apiServerAddr, p.ID.String()), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// CHECKME add timeout to retry scheduling?
	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		} else {
			return fmt.Errorf("error assigning worker to pod %s: %s", p.Name, string(bytes))
		}
	}

	return nil
}

func (s *Scheduler) schedulePods(pods []model.Pod) {
	for _, p := range pods {
		err := s.schedulePod(p)
		if err != nil {
			log.Printf("error scheduling pod %s: %v", p.Name, err)
		}
	}
}

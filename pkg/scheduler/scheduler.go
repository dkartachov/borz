package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/pkg/model"
)

type Scheduler struct {
	APIServerAddr string
	client        *http.Client
}

func NewScheduler(APIServer string) *Scheduler {
	return &Scheduler{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Scheduler) Start() {
	for {
		// TODO add filter param to API Server GET /pods endpoint to only get unscheduled pods
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/pods", s.APIServerAddr), nil)
		if err != nil {
			log.Fatal(err)
			continue
		}

		resp, err := s.client.Do(req)
		if err != nil {
			log.Fatal(err)
			continue
		}

		var pods []model.Pod

		err = json.NewDecoder(resp.Body).Decode(&pods)
		if err != nil {
			log.Fatal(err)
			continue
		}

		var unscheduledPods []model.Pod

		for _, p := range pods {
			if p.Worker == "" {
				unscheduledPods = append(unscheduledPods, p)
			}
		}

		s.schedulePods(unscheduledPods)
		resp.Body.Close()
		time.Sleep(time.Second)
	}
}

func (s *Scheduler) schedulePods(pods []model.Pod) {
	for _, p := range pods {

	}
}

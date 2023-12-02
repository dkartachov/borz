package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/manager/scheduler"
	"github.com/dkartachov/borz/internal/model"
)

type DeploymentController struct {
	database  *database.Database
	scheduler *scheduler.Scheduler
}

func NewDeploymentController(db *database.Database, s *scheduler.Scheduler) *DeploymentController {
	return &DeploymentController{
		database:  db,
		scheduler: s,
	}
}

func (dc *DeploymentController) Start() {
	for {
		dc.updateReplicas()
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
}

func (dc *DeploymentController) updateReplicas() {
	workers := dc.database.GetWorkers()
	deployments := dc.database.GetDeployments()

	for _, deploy := range deployments {
		for _, w := range workers {
			client := &http.Client{}

			req, err := http.NewRequest("GET", fmt.Sprintf("%s/pods", w), nil)
			if err != nil {
				log.Printf("error creating request: %v", err)
				continue
			}

			q := req.URL.Query()
			q.Add("name", deploy.MatchPod)
			req.URL.RawQuery = q.Encode()

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("error connecting to worker %s", w)
				continue
			}

			var pods []model.Pod
			json.NewDecoder(resp.Body).Decode(&pods)
		}
	}
}

package controller

import (
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

type PodController struct {
	database *database.Database
}

func NewPodController(db *database.Database) *PodController {
	return &PodController{
		database: db,
	}
}

func (pc *PodController) Start() {
	for {
		pc.updatePods()
		time.Sleep(time.Millisecond * time.Duration(1000))
	}
}

func (pc *PodController) SendPodForDeletion(podName string) (int, string) {
	worker, _ := pc.database.GetWorkerFromPod(podName)
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

func (pc *PodController) updatePods() {
	var wg sync.WaitGroup

	// fetch pods from all workers asynchronously
	for _, w := range pc.database.GetWorkers() {
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
					pc.database.DeletePod(p.Name)
					pc.database.RemovePodFromWorker(p.Name)
				default:
					pc.database.AddPod(p)
				}
			}
		}(w)
	}

	wg.Wait()
}

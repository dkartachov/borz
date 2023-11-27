package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dkartachov/borz/internal/manager/database"
	"github.com/dkartachov/borz/internal/model"
)

type Scheduler struct {
	Database *database.Database
	// TODO improve queue by abstracting the tasks being processed (https://mrkaran.dev/posts/job-queue-golang/)
	PodQueue   chan model.Pod
	NextWorker int
	Client     *http.Client

	queueOnline bool
	shutdown    chan struct{}
}

func New(db *database.Database) *Scheduler {
	return &Scheduler{
		// TODO make channel size configurable
		PodQueue:   make(chan model.Pod, 10),
		NextWorker: 0,
		Database:   db,
		Client:     &http.Client{},
	}
}

func (s *Scheduler) Start() {
	s.queueOnline = true

	go s.schedulePods()
	// s.updatePods(1000)
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

	s.Database.AddPodToWorker(p.Name, w)
}

// TODO Need algorithm for selecting workers. This is a first pass round-robin approach.
func (s *Scheduler) selectWorker() string {
	if s.NextWorker == len(s.Database.GetWorkers())-1 {
		s.NextWorker = 0
	}

	return s.Database.GetWorkers()[s.NextWorker]
}

package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dkartachov/borz/internal/task"
	"github.com/google/uuid"
)

type Signal struct {
	ShutdownAPI           chan struct{}
	ShutdownTaskScheduler chan struct{}
}

type Manager struct {
	name          string
	Workers       []string // worker addresses
	nextWorker    int
	Queue         chan task.Task
	tasks         map[uuid.UUID]task.Task
	taskWorkerMap map[uuid.UUID]string
	signal        Signal
}

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")
	m := Manager{
		name:          name,
		Workers:       workers, // TODO need a better way to instantiate this array
		nextWorker:    0,
		Queue:         make(chan task.Task),
		tasks:         make(map[uuid.UUID]task.Task),
		taskWorkerMap: make(map[uuid.UUID]string),
		signal:        Signal{ShutdownAPI: make(chan struct{}), ShutdownTaskScheduler: make(chan struct{})},
	}
	a := Api{
		Address: "localhost",
		Port:    port,
		Manager: m,
	}

	go m.runTaskScheduler()
	go m.updateTasks(1000)
	a.Start()

	log.Printf("[%s] exiting", m.name)
}

func (m *Manager) AddTask(t task.Task) {
	m.Queue <- t
}

func (m *Manager) GetTasks() []TaskResponse {
	var taskResp []TaskResponse = []TaskResponse{}

	for _, t := range m.tasks {
		taskResp = append(taskResp, TaskResponse{Task: t, Worker: m.taskWorkerMap[t.Id]})
	}

	return taskResp
}

// TODO Improve how workers are selected to send tasks to. This is a first pass round-robin approach
func (m *Manager) selectWorker(taskId uuid.UUID) string {
	if w, ok := m.taskWorkerMap[taskId]; ok {
		return w
	}

	if m.nextWorker == len(m.Workers)-1 {
		m.nextWorker = 0
	} else {
		m.nextWorker += 1
	}

	return m.Workers[m.nextWorker]
}

func (m *Manager) sendTask(t task.Task) {
	log.Printf("[%v] sending task %s", m.name, t.Name)
	// CHECKME Is there a better place to set this?
	if t.State == task.Pending {
		t.State = task.Scheduled
	}

	w := m.selectWorker(t.Id)
	log.Printf("[%v] selected worker %v", m.name, w)

	req, err := json.Marshal(t)
	if err != nil {
		log.Printf("[%v] error marshaling task request %v", m.name, t.Id)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/tasks", w), "application/json", bytes.NewBuffer(req))
	if err != nil {
		log.Printf("[%v] error connecting to %s", m.name, w)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[%v] error sending task to %s", m.name, w)
		resp.Body.Close()
		return
	}

	resp.Body.Close()
	m.taskWorkerMap[t.Id] = w
}

func (m *Manager) updateTasks(intervalMillis int) {
	for {
		select {
		case <-m.signal.ShutdownTaskScheduler:
			log.Printf("[%s] shutting down task updater", m.name)
			return
		default:
			log.Printf("[%v] updating tasks", m.name)
			for _, w := range m.Workers {
				resp, err := http.Get(fmt.Sprintf("%s/tasks", w))
				if err != nil {
					log.Printf("[%v] error connecting to %s", m.name, w)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					log.Printf("[%v] error getting tasks from %s", m.name, w)
					resp.Body.Close()
					continue
				}

				var tasks []task.Task
				json.NewDecoder(resp.Body).Decode(&tasks)

				for _, t := range tasks {
					m.tasks[t.Id] = t
					m.taskWorkerMap[t.Id] = w
				}
			}
			time.Sleep(time.Millisecond * time.Duration(intervalMillis))
		}
	}
}

func (m *Manager) runTaskScheduler() {
	for {
		select {
		case <-m.signal.ShutdownTaskScheduler:
			log.Printf("[%s] shutting down task scheduler", m.name)
			return
		case t := <-m.Queue:
			go m.sendTask(t)
		}
	}
}

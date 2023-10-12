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
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	name          string
	Workers       []string // worker addresses
	nextWorker    int
	TaskQueue     *queue.Queue
	tasks         map[uuid.UUID]task.Task
	taskWorkerMap map[uuid.UUID]string
}

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	workers := strings.Split(args[2], ",")
	m := Manager{
		name:          name,
		Workers:       workers, // TODO need a better way to instantiate this array
		nextWorker:    0,
		TaskQueue:     queue.New(),
		tasks:         make(map[uuid.UUID]task.Task),
		taskWorkerMap: make(map[uuid.UUID]string),
	}
	a := Api{
		Address: "localhost",
		Port:    port,
		Manager: m,
	}

	go m.run(1000)
	a.Init()
}

func (m *Manager) AddTask(t task.Task) {
	m.TaskQueue.Enqueue(t)
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
	}

	return m.Workers[m.nextWorker]
}

func (m *Manager) sendTasks() {
	log.Printf("[%v] sending tasks", m.name)
	if m.TaskQueue.Len() > 0 {
		ti := m.TaskQueue.Dequeue()
		t := ti.(task.Task)
		// CHECKME Is there a better place to set this?
		if t.State == task.Pending {
			t.State = task.Scheduled
		}

		w := m.selectWorker(t.Id)
		log.Printf("[%v] sending task to %v", m.name, w)

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
}

func (m *Manager) updateTasks() {
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
}

func (m *Manager) run(millis int) {
	for {
		// CHECKME should updateTasks and sendTasks run in parallel?
		m.updateTasks()
		m.sendTasks()

		log.Printf("[%v] sleeping for %d ms", m.name, millis)
		time.Sleep(time.Millisecond * time.Duration(millis))
	}
}

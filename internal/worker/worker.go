package worker

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dkartachov/borz/internal/docker"
	"github.com/dkartachov/borz/internal/task"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	name      string
	TaskQueue *queue.Queue
	tasks     map[uuid.UUID]task.Task
	Signal    Signal
}

type Signal struct {
	ShutdownAPI        chan struct{}
	ShutdownTaskRunner chan struct{}
}

func Run(args []string) {
	name := args[0]
	port, _ := strconv.Atoi(args[1])
	w := Worker{
		name:      name,
		TaskQueue: queue.New(),
		tasks:     make(map[uuid.UUID]task.Task), // CHECKME connect to database instead of storing tasks in memory?
		Signal:    Signal{ShutdownAPI: make(chan struct{}), ShutdownTaskRunner: make(chan struct{})},
	}
	a := Api{
		Address: "localhost",
		Port:    port,
		Worker:  w,
	}
	go w.runTasks(1000)
	a.Start()

	log.Printf("[%v] exiting", w.name)
}

func (w *Worker) Name() string {
	return w.name
}

func (w *Worker) AddTask(t task.Task) {
	w.TaskQueue.Enqueue(t)
}

func (w *Worker) StartTask(t task.Task) error {
	log.Printf("[%v] starting task %v", w.name, t.Id)

	d := docker.Docker{Image: t.Image}
	containerID, err := d.Start()
	if err != nil {
		log.Printf("[%s] error starting task %v: %v", w.name, t.Id, err)
	}

	t.ContainerID = containerID
	t.State = task.Running
	t.StartTime = time.Now().UTC()
	w.tasks[t.Id] = t

	return nil
}

func (w *Worker) StopTask(t task.Task) error {
	log.Printf("[%v] stopping task %v", w.name, t.Id)

	_, ok := w.tasks[t.Id]
	if !ok {
		return fmt.Errorf("[%v] task %v not found", w.name, t.Id)
	}

	d := docker.Docker{ContainerID: t.ContainerID}
	if err := d.Stop(); err != nil {
		return err
	}

	t.State = task.Completed
	t.FinishTime = time.Now().UTC()
	w.tasks[t.Id] = t

	return nil
}

func (w *Worker) Task(id uuid.UUID) (task.Task, bool) {
	t, exists := w.tasks[id]
	return t, exists
}

func (w *Worker) GetTasks() []task.Task {
	var tasks []task.Task = []task.Task{}

	for _, t := range w.tasks {
		tasks = append(tasks, t)
	}

	return tasks
}

// CHECKME Dequeue ALL tasks and run them concurrently using goroutines?
func (w *Worker) runTasks(intervalMillis int) {
	for {
		select {
		case <-w.Signal.ShutdownTaskRunner:
			log.Printf("[%v] shutting down task runner", w.name)
			log.Printf("[%v] stopping tasks", w.name)
			for _, t := range w.tasks {
				w.StopTask(t)
			}
			// CHECKME anything else?
			close(w.Signal.ShutdownAPI)
			return
		default:
			if w.TaskQueue.Len() > 0 {
				log.Printf("[%v] updating tasks", w.name)
				if err := w.runTask(); err != nil {
					log.Printf("[%v] error running task: %v", w.name, err)
				}
			} else {
				log.Printf("[%v] no tasks found in queue", w.name)
			}
		}
		log.Printf("[%v] sleeping for %d ms", w.name, intervalMillis)
		time.Sleep(time.Millisecond * time.Duration(intervalMillis))
	}
}

func (w *Worker) runTask() error {
	ti := w.TaskQueue.Dequeue()
	taskQueued := ti.(task.Task)
	taskPersisted, ok := w.tasks[taskQueued.Id]
	if !ok {
		taskPersisted = taskQueued
		newTask := taskQueued
		w.tasks[taskQueued.Id] = newTask
	}

	var err error = nil
	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			err = w.StartTask(taskQueued)
		case task.Stopping:
			err = w.StopTask(taskPersisted)
		default:
			err = fmt.Errorf("state transition to %v not supported", taskQueued.State)
		}
	} else {
		err = fmt.Errorf("invalid state transition %v to %v", taskPersisted.State, taskQueued.State)
	}

	return err
}

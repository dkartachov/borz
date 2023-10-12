package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id          uuid.UUID
	Name        string
	State       State
	Image       string
	ContainerID string
	StartTime   time.Time
	FinishTime  time.Time
}

type TaskEvent struct {
	Id    uuid.UUID
	State State // Desired state, the Task struct contains the current state
	Task  Task
}

package model

import "github.com/google/uuid"

type State string

const (
	Pending   State = "Pending"
	Scheduled State = "Scheduled"
	Running   State = "Running"
	Stopping  State = "Stopping"
	Stopped   State = "Stopped"
	Error     State = "Error"
)

type Container struct {
	Name  string
	Image string
	Port  int
	ID    string
}

type Pod struct {
	ID         uuid.UUID
	Name       string
	Containers []Container
	State      State
	Worker     string
}

func ValidStateTransition(source State, dest State) bool {
	stateTransitionMap := map[State][]State{
		Pending:   {Pending, Scheduled, Error},
		Scheduled: {Scheduled, Running, Error},
		Running:   {Running, Stopping, Error},
		Stopping:  {Stopping, Stopped, Error},
		Stopped:   {Stopped, Error},
		Error:     {Error},
	}

	validDestStates := stateTransitionMap[source]
	for _, s := range validDestStates {
		if s == dest {
			return true
		}
	}
	return false
}

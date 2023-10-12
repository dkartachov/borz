package task

type State int

const (
	Pending   State = iota // Received by Manager, waiting to be sent to a worker
	Scheduled              // Received by a worker, task is starting
	Running
	Stopping
	Completed
	Failed
)

func ValidStateTransition(source State, dest State) bool {
	stateTransitionMap := map[State][]State{
		Pending:   {Pending, Scheduled},
		Scheduled: {Scheduled, Running, Failed},
		Running:   {Running, Stopping, Completed, Failed},
		Stopping:  {Stopping, Completed},
		Completed: {Completed},
		Failed:    {Failed},
	}

	validDestStates := stateTransitionMap[source]
	for _, s := range validDestStates {
		if s == dest {
			return true
		}
	}
	return false
}

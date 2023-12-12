package borzlet

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Stopping
	Error
)

func ValidStateTransition(source State, dest State) bool {
	stateTransitionMap := map[State][]State{
		Pending:   {Pending, Scheduled, Error},
		Scheduled: {Scheduled, Running, Error},
		Running:   {Running, Stopping, Error},
		Stopping:  {Stopping, Error},
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

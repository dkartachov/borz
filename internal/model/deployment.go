package model

type Deployment struct {
	Name            string // unique
	MatchPod        string // match pod name
	ActualReplicas  int
	DesiredReplicas int
	Pod             Pod
}

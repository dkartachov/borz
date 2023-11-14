package model

type Deployment struct {
	Name       string // unique
	Containers []Container
	Replicas   int
}

func NewDeployment(name string, containers []Container, replicas int) Deployment {
	return Deployment{
		Name:       name,
		Containers: containers,
		Replicas:   replicas,
	}
}

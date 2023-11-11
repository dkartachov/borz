package deployment

import "github.com/dkartachov/borz/internal/task"

type Deployment struct {
	Name     string // unique
	Tasks    []task.Task
	Replicas int
}

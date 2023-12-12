package store

import (
	"github.com/dkartachov/borz/pkg/model"
	"github.com/google/uuid"
)

type Store interface {
	AddPod(p model.Pod) error
	GetPods() ([]model.Pod, error)
	GetPod(id uuid.UUID) (*model.Pod, error)
	DeletePod(id uuid.UUID) error

	AddWorker(w model.Worker) error
	GetWorkers() ([]model.Worker, error)
}

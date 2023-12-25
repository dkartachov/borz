package store

import (
	"github.com/dkartachov/borz/pkg/model"
	"github.com/google/uuid"
)

type Store interface {
	AddPod(model.Pod) error
	UpdatePod(model.Pod) error
	GetPods() ([]model.Pod, error)
	GetPod(uuid.UUID) (*model.Pod, error)
	DeletePod(uuid.UUID) error

	AddWorker(model.Worker) error
	GetWorkers() ([]model.Worker, error)
}

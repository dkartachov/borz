package store

import (
	"sync"

	"github.com/dkartachov/borz/pkg/model"
	"github.com/google/uuid"
)

type MemoryStore struct {
	workers         map[string]model.Worker
	deployments     map[string]model.Deployment
	pods            map[string]model.Pod
	podNameByWorker map[string]string
	podByDeployment map[string]string

	mu sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		workers:         make(map[string]model.Worker),
		deployments:     make(map[string]model.Deployment),
		pods:            make(map[string]model.Pod),
		podNameByWorker: make(map[string]string),
		podByDeployment: make(map[string]string),
	}
}

func (s *MemoryStore) AddPod(p model.Pod) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pods[p.ID.String()] = p
	return nil
}

func (s *MemoryStore) UpdatePod(p model.Pod) error {
	return s.AddPod(p)
}

func (s *MemoryStore) GetPod(id uuid.UUID) (*model.Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.pods[id.String()]
	if !ok {
		return nil, nil
	}
	return &p, nil
}

func (s *MemoryStore) GetPods() ([]model.Pod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pods := []model.Pod{}

	for _, p := range s.pods {
		pods = append(pods, p)
	}

	return pods, nil
}

func (s *MemoryStore) DeletePod(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pods, id.String())

	return nil
}

func (s *MemoryStore) AddWorker(w model.Worker) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.workers[w.Addr] = w
	return nil
}

func (s *MemoryStore) GetWorkers() ([]model.Worker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workers := []model.Worker{}

	for _, w := range s.workers {
		workers = append(workers, w)
	}

	return workers, nil
}

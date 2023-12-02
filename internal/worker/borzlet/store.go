package borzlet

import (
	"sync"

	"github.com/dkartachov/borz/internal/model"
)

type Store struct {
	pods map[string]model.Pod
	mu   sync.RWMutex
}

func (s *Store) Init() {
	s.pods = make(map[string]model.Pod)
}

func (s *Store) AddPod(p model.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pods[p.Name] = p
}

func (s *Store) AddContainerID(podName string, containerName string, containerID string) {
	p := s.GetPod(podName)

	for i, c := range p.Containers {
		if c.Name == containerName {
			c.ID = containerID
			p.Containers[i] = c
			s.AddPod(p)
			return
		}
	}
}

func (s *Store) GetPods() []model.Pod {
	pods := []model.Pod{}
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.pods {
		pods = append(pods, p)
	}

	return pods
}

func (s *Store) GetPodsByName(name string) []model.Pod {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pods []model.Pod

	for _, p := range s.pods {
		if p.Name == name {
			pods = append(pods, p)
		}
	}

	return pods
}

func (s *Store) GetPod(name string) model.Pod {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pod := s.pods[name]
	return pod
}

func (s *Store) DeletePod(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pods, name)
}

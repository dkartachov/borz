package borzlet

import (
	"github.com/dkartachov/borz/internal/model"
	"github.com/google/uuid"
)

func (pm *Borzlet) addPod(p model.Pod) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.pods[p.ID] = p
}

func (pm *Borzlet) getPod(id uuid.UUID) model.Pod {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.pods[id]
}

func (pm *Borzlet) getPods() []model.Pod {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var pods []model.Pod

	for _, p := range pm.pods {
		pods = append(pods, p)
	}

	return pods
}

func (pm *Borzlet) getRunningPods() []model.Pod {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var runningPods []model.Pod

	for _, pod := range pm.getPods() {
		if pod.State == model.Running {
			runningPods = append(runningPods, pod)
		}
	}

	return runningPods
}

func (pm *Borzlet) updateContainerInPod(podId uuid.UUID, container model.Container) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pod := pm.getPod(podId)

	for i := 0; i < len(pod.Containers); i++ {
		if pod.Containers[i].Name == container.Name {
			pod.Containers[i] = container
			break
		}
	}
}

// TODO add some kind of garbage collector that runs every few minutes to purge any stopped pods from memory

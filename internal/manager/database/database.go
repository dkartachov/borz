package database

import (
	"sync"

	"github.com/dkartachov/borz/internal/model"
)

type Database struct {
	Workers     []string
	Deployments map[string]model.Deployment
	Pods        map[string]model.Pod
	mu          sync.RWMutex
}

func (d *Database) AddDeployment(dep model.Deployment) {
	d.Deployments[dep.Name] = dep
}

func (d *Database) GetDeployments() []model.Deployment {
	deployments := []model.Deployment{}

	for _, d := range d.Deployments {
		deployments = append(deployments, d)
	}

	return deployments
}

func (d *Database) DeleteDeployment(name string) {
	delete(d.Deployments, name)
}

func (d *Database) PodNameExists(name string) bool {
	_, ok := d.Pods[name]
	return ok
}

func (d *Database) AddPod(p model.Pod) {
	d.mu.Lock()
	d.Pods[p.Name] = p
	d.mu.Unlock()
}

// func (d *Database) OverwritePods(pods []model.Pod) {
// 	newPods := make(map[string]model.Pod)

// 	for _, pod := range pods {
// 		newPods[pod.Name] = pod
// 	}

// 	d.mu.Lock()
// 	defer d.mu.Unlock()
// 	d.Pods = newPods
// }

func (d *Database) DeletePod(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.Pods, name)
}

func (d *Database) GetPods() []model.Pod {
	pods := []model.Pod{}

	for _, p := range d.Pods {
		pods = append(pods, p)
	}

	return pods
}

func (d *Database) GetPod(name string) model.Pod {
	return d.Pods[name]
}

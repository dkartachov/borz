package database

import (
	"sync"

	"github.com/dkartachov/borz/internal/model"
)

type Database struct {
	workers         []string
	deployments     map[string]model.Deployment
	pods            map[string]model.Pod
	podNameByWorker map[string]string
	podByDeployment map[string]string

	mu sync.RWMutex
}

func (d *Database) Init() {
	d.workers = []string{}
	d.deployments = make(map[string]model.Deployment)
	d.pods = make(map[string]model.Pod)
	d.podNameByWorker = make(map[string]string)
	d.podByDeployment = make(map[string]string)
}

func (d *Database) GetWorkers() []string {
	return d.workers
}

func (d *Database) AddWorkers(workers []string) {
	d.workers = workers
}

func (d *Database) GetWorkerFromPod(podName string) (string, bool) {
	w, ok := d.podNameByWorker[podName]
	return w, ok
}

func (d *Database) GetPodsByWorkers() map[string]string {
	return d.podNameByWorker
}

func (d *Database) AddPodToWorker(podName string, worker string) {
	d.podNameByWorker[podName] = worker
}

func (d *Database) RemovePodFromWorker(podName string) {
	delete(d.podNameByWorker, podName)
}

func (d *Database) AddPodToDeployment(podName string, deploymentName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.podByDeployment[podName] = deploymentName
}

func (d *Database) RemovePodFromDeployment(podName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.podByDeployment, podName)
}

func (d *Database) GetPodByDeployment() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.podByDeployment
}

func (d *Database) AddDeployment(dep model.Deployment) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deployments[dep.Name] = dep
}

func (d *Database) GetDeployments() []model.Deployment {
	deployments := []model.Deployment{}

	for _, d := range d.deployments {
		deployments = append(deployments, d)
	}

	return deployments
}

func (d *Database) DeleteDeployment(name string) {
	delete(d.deployments, name)
}

func (d *Database) PodNameExists(name string) bool {
	_, ok := d.pods[name]
	return ok
}

func (d *Database) AddPod(p model.Pod) {
	d.mu.Lock()
	d.pods[p.Name] = p
	d.mu.Unlock()
}

func (d *Database) DeletePod(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.pods, name)
}

func (d *Database) GetPods() []model.Pod {
	pods := []model.Pod{}

	for _, p := range d.pods {
		pods = append(pods, p)
	}

	return pods
}

func (d *Database) GetPod(name string) model.Pod {
	return d.pods[name]
}

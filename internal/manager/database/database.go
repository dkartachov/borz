package database

import (
	"github.com/dkartachov/borz/internal/model"
)

type Database struct {
	Workers     []string
	Deployments map[string]model.Deployment
	Pods        map[string]model.Pod
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
	d.Pods[p.Name] = p
}

func (d *Database) DeletePod(name string) {
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

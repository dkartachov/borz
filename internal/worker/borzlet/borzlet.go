package borzlet

import (
	"log"
	"time"

	"github.com/dkartachov/borz/internal/docker"
	"github.com/dkartachov/borz/internal/model"
	"github.com/golang-collections/collections/queue"
)

type Borzlet struct {
	JobQueue *queue.Queue
	Store    *Store
}

func (b *Borzlet) Start() {
	for {
		b.RunPods()
		time.Sleep(time.Millisecond * 1000)
	}
}

func (b *Borzlet) EnqueuePod(p model.Pod) {
	b.JobQueue.Enqueue(p)
}

func (b *Borzlet) RunPods() {
	for b.JobQueue.Len() > 0 {
		pi := b.JobQueue.Dequeue()
		p := pi.(model.Pod)

		switch p.State {
		case model.Scheduled:
			b.startPod(p)
		case model.Stopping:
			b.stopPod(p)
		}
	}
}

func (b *Borzlet) startPod(p model.Pod) {
	log.Printf("starting pod %s", p.Name)

	for _, c := range p.Containers {
		log.Printf("starting container %v", c.Name)
		d := docker.Docker{Image: c.Image}
		containerID, err := d.Start()
		if err != nil {
			p.State = model.Error
			b.Store.AddPod(p)
			log.Printf("error starting container: %v", err)
			return
		}

		b.Store.AddContainerID(p.Name, c.Name, containerID)
	}

	p.State = model.Running
	b.Store.AddPod(p)
}

func (b *Borzlet) stopPod(pod model.Pod) {
	log.Printf("stopping pod %s", pod.Name)

	for _, c := range pod.Containers {
		log.Printf("stopping container %v", c.Name)
		d := docker.Docker{ContainerID: c.ID}
		if err := d.Stop(); err != nil {
			pod.State = model.Error
			b.Store.AddPod(pod)
			log.Printf("error stopping container: %v", err)
			return
		}
	}

	pod.State = model.Stopped
	b.Store.AddPod(pod)
}

// TODO add some kind of purge function that runs every few minutes to purge any stopped pods from memory

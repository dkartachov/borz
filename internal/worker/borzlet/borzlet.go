package borzlet

import (
	"context"
	"fmt"
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
		b.RunPods(context.Background())
		time.Sleep(time.Millisecond * 1000)
	}
}

func (b *Borzlet) EnqueuePod(p model.Pod) {
	b.JobQueue.Enqueue(p)
}

func (b *Borzlet) RunPods(ctx context.Context) {
	// CHECKME Should probably run jobs concurrently.
	// Right now a problematic pod will prevent other pods from being scheduled.
	// Use a channel instead of queue?
	for b.JobQueue.Len() > 0 {
		pi := b.JobQueue.Dequeue()
		p := pi.(model.Pod)
		// CHECKME Should startPod and stopPod have different timeouts instead?
		ctx, cancelCtx := context.WithTimeout(ctx, 5*time.Minute) // default max 5 minutes to start/stop pods
		defer cancelCtx()

		switch p.State {
		case model.Scheduled:
			err := b.startPod(ctx, p)
			if err != nil {
				log.Printf("error starting pod %s: %v", p.Name, err)
			}
		case model.Stopping:
			err := b.stopPod(ctx, p)
			if err != nil {
				log.Printf("error stopping pod %s: %v", p.Name, err)
			}
		}
	}
}

func (b *Borzlet) startPod(ctx context.Context, p model.Pod) error {
	log.Printf("starting pod %s", p.Name)

	for _, c := range p.Containers {
		log.Printf("starting container %v", c.Name)
		d := docker.Docker{Image: c.Image}
		containerID, err := d.Start(ctx)
		if err != nil {
			p.State = model.Error
			b.Store.AddPod(p)
			return fmt.Errorf("error starting container: %v", err)
		}

		b.Store.AddContainerID(p.Name, c.Name, containerID)
	}

	p.State = model.Running
	b.Store.AddPod(p)
	return nil
}

func (b *Borzlet) stopPod(ctx context.Context, pod model.Pod) error {
	log.Printf("stopping pod %s", pod.Name)

	for _, c := range pod.Containers {
		log.Printf("stopping container %v", c.Name)

		d := docker.Docker{ContainerID: c.ID}
		err := d.Stop(ctx)
		if err != nil {
			pod.State = model.Error
			b.Store.AddPod(pod)
			return fmt.Errorf("error stopping container: %v", err)
		}
	}

	pod.State = model.Stopped
	b.Store.AddPod(pod)
	return nil
}

func (b *Borzlet) StopPods(ctx context.Context) error {
	// CHECKME Should timeout be variable and depend on number of pods?
	ctx, cancelCtx := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelCtx()

	stoppedPods := make(chan model.Pod, len(b.Store.GetPods()))

	for _, p := range b.Store.GetPods() {
		go func(pod model.Pod) {
			err := b.stopPod(context.Background(), pod)
			if err != nil {
				log.Printf("error stopping pod %s: %v", pod.Name, err)
				return
			}

			stoppedPods <- pod
		}(p)
	}

	for i := 0; i < len(b.Store.GetPods()); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case p := <-stoppedPods:
			log.Printf("successfully stopped pod %s", p.Name)
		}
	}

	return nil
}

// TODO add some kind of purge function that runs every few minutes to purge any stopped pods from memory

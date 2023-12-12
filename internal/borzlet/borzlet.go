package borzlet

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dkartachov/borz/internal/docker"
	"github.com/dkartachov/borz/internal/model"
	"github.com/google/uuid"
)

type Borzlet struct {
	client *http.Client
	pods   map[uuid.UUID]model.Pod
	mu     sync.RWMutex
}

func newBorzlet() *Borzlet {
	return &Borzlet{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		pods: make(map[uuid.UUID]model.Pod),
	}
}

func Run(args []string) {
	address := args[0]
	port, _ := strconv.Atoi(args[1])

	borzlet := newBorzlet()
	server := NewBorzletServer(address, port, borzlet)
	server.Start()

	log.Print("exiting")
}

func (b *Borzlet) runPod(p model.Pod) {
	// CHECKME Should startPod and stopPod have different timeouts instead?
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Minute) // default max 5 minutes to start/stop pods
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

func (b *Borzlet) startPod(ctx context.Context, p model.Pod) error {
	log.Printf("starting pod %s", p.Name)

	// CHECKME start containers concurrently?
	for _, c := range p.Containers {
		log.Printf("starting container %v", c.Name)
		d := docker.Docker{Image: c.Image}
		containerID, err := d.Start(ctx)
		if err != nil {
			p.State = model.Error
			b.addPod(p)
			return fmt.Errorf("error starting container: %v", err)
		}

		c.ID = containerID
		b.updateContainerInPod(p.ID, c)
	}

	p.State = model.Running
	b.addPod(p)
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
			b.addPod(pod)
			return fmt.Errorf("error stopping container: %v", err)
		}
	}

	pod.State = model.Stopped
	b.addPod(pod)
	return nil
}

func (b *Borzlet) StopPods(ctx context.Context) error {
	// CHECKME Should timeout be variable and depend on number of pods?
	ctx, cancelCtx := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelCtx()

	runningPods := b.getRunningPods()
	stoppedPods := make(chan model.Pod, len(runningPods))
	failedPods := make(chan model.Pod, len(runningPods))

	for _, p := range runningPods {
		go func(pod model.Pod) {
			err := b.stopPod(context.Background(), pod)
			if err != nil {
				log.Printf("error stopping pod %s: %v", pod.Name, err)
				failedPods <- pod
				return
			}

			stoppedPods <- pod
		}(p)
	}

	for i := 0; i < len(runningPods); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case p := <-stoppedPods:
			log.Printf("successfully stopped pod %s", p.Name)
		}
	}

	return nil
}

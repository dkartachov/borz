package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Docker struct {
	Image       string
	ContainerID string
}

func NewDockerClient() *client.Client {
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return c
}

func (d *Docker) Start(ctx context.Context) (string, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelCtx()

	c := NewDockerClient()
	reader, err := c.ImagePull(ctx, d.Image, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("error pulling image %s: %v", d.Image, err)
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

	// TODO add more config options
	cc := container.Config{
		Image: d.Image,
	}
	hc := container.HostConfig{
		PublishAllPorts: true,
	}

	createResp, err := c.ContainerCreate(ctx, &cc, &hc, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("error creating container: %v", err)
	}

	if err = c.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("error starting container: %v", err)
	}

	out, err := c.ContainerLogs(ctx, createResp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("error getting container logs: %v", err)
	}

	defer out.Close()

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	d.ContainerID = createResp.ID

	return createResp.ID, nil
}

func (d *Docker) Stop(ctx context.Context) error {
	c := NewDockerClient()

	err := c.ContainerStop(ctx, d.ContainerID, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("error stopping container %s: %v", d.ContainerID, err)
	}

	err = c.ContainerRemove(ctx, d.ContainerID, types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("error removing container %s: %v", d.ContainerID, err)
	}

	return nil
}

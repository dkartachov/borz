package docker

import (
	"context"
	"io"
	"os"

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

func (d *Docker) Start() (string, error) {
	ctx := context.Background()
	c := NewDockerClient()

	reader, err := c.ImagePull(ctx, d.Image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	io.Copy(os.Stdout, reader)

	cc := container.Config{
		Image: d.Image,
	}
	hc := container.HostConfig{
		PublishAllPorts: true,
	}
	createResp, err := c.ContainerCreate(ctx, &cc, &hc, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err = c.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	out, err := c.ContainerLogs(ctx, createResp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	d.ContainerID = createResp.ID

	return createResp.ID, nil
}

func (d *Docker) Stop() error {
	ctx := context.Background()
	c := NewDockerClient()

	if err := c.ContainerStop(ctx, d.ContainerID, container.StopOptions{}); err != nil {
		return err
	}

	return c.ContainerRemove(ctx, d.ContainerID, types.ContainerRemoveOptions{})
}

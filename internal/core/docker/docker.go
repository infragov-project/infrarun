package docker

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/infragov-project/infrarun/internal/core/utils"
)

type DockerEngine struct {
	Client *client.Client
}

func NewDockerEngine() (*DockerEngine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return &DockerEngine{
		Client: cli,
	}, err
}

func (engine *DockerEngine) EnsureImageExists(ctx context.Context, imageName string) error {
	reader, err := engine.Client.ImagePull(ctx, imageName, image.PullOptions{})

	if err != nil {
		// Initial pull failled, check if image exists locally

		_, inspectErr := engine.Client.ImageInspect(ctx, imageName)

		if inspectErr == nil {
			return nil // Image exists locally
		}

		return err
	}

	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		scanner.Text()
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	return nil
}

type ContainerInfo struct {
	Image       string
	Cmd         []string
	VolumeBinds []VolumeBind
}

type VolumeBind struct {
	Host  string
	Guest string
}

func (engine *DockerEngine) RunContainer(ctx context.Context, info ContainerInfo) (string, error) {

	resp, err := engine.Client.ContainerCreate(ctx, &container.Config{
		Image: info.Image,
		Cmd:   info.Cmd,
		Tty:   false,
	}, &container.HostConfig{
		Binds: utils.Map(info.VolumeBinds, func(x VolumeBind) string { return x.Host + ":" + x.Guest }),
	}, nil, nil, "")

	if err != nil {
		return "", err
	}

	containerID := resp.ID

	err = engine.Client.ContainerStart(ctx, containerID, container.StartOptions{})

	if err != nil {
		return "", err
	}

	statusCh, errCh := engine.Client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case error := <-errCh: // Got error from ContainerWait
		return "", error
	case <-statusCh:
		return containerID, nil
	}

}

func (engine *DockerEngine) CaptureStdOut(ctx context.Context, containerID string) ([]byte, error) {
	readCloser, err := engine.Client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Timestamps: false,
		Follow:     false,
		Details:    false,
	})

	if err != nil {
		return nil, err
	}

	defer readCloser.Close()

	return io.ReadAll(readCloser)
}

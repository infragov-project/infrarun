package engine

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/infragov-project/infrarun/internal/core/docker"
	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

type ToolExecution struct {
	Path   string
	Tool   *tools.Tool
	Report *sarif.Report
	Err error
}

func NewToolExecution(tool *tools.Tool, path string) (*ToolExecution, error) {
	absPath, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	return &ToolExecution{
		Tool: tool,
		Path: absPath,
	}, nil
}

type InfrarunEngine struct {
	Backend *docker.DockerEngine
}

func NewInfrarunEngine() (*InfrarunEngine, error) {
	backend, err := docker.NewDockerEngine()

	if err != nil {
		return nil, err
	}

	return &InfrarunEngine{
		Backend: backend,
	}, nil
}

func (engine *InfrarunEngine) Execute(ctx context.Context, toolExecution *ToolExecution) ([]byte, error) {
	if err := engine.Backend.EnsureImageExists(ctx, toolExecution.Tool.Image); err != nil {
		return nil, err
	}

	volumeBinds := []docker.VolumeBind{
		{Host: toolExecution.Path, Guest: toolExecution.Tool.InputPath},
	}

	outputDir := ""

	if !toolExecution.Tool.CaptureStdout {
		// Tempdir for output of container
		var err error
		outputDir, err = os.MkdirTemp("", "infrarun-")

		if err != nil {
			return nil, err
		}

		defer os.RemoveAll(outputDir)

		volumeBinds = append(volumeBinds, docker.VolumeBind{
			Host:  outputDir,
			Guest: toolExecution.Tool.OutputPath,
		})
	}

	containerID, err := engine.Backend.RunContainer(ctx, docker.ContainerInfo{
		Image:       toolExecution.Tool.Image,
		Cmd:         toolExecution.Tool.Cmd,
		VolumeBinds: volumeBinds,
	})

	if err != nil {
		return nil, err
	}

	if toolExecution.Tool.CaptureStdout {
		data, err := engine.Backend.CaptureStdOut(ctx, containerID)

		if err != nil {
			return nil, err
		}

		return extractStdoutFrames(data)
	} else {
		outputFilePath := filepath.Clean(outputDir + "/" + toolExecution.Tool.OutputFile)
		return os.ReadFile(outputFilePath)
	}

}

// This function processes docker's log format and extracts all stdout stream content.
// This allows us to run the containers in non-TTY mode and still get the clean stdout
// content.
func extractStdoutFrames(data []byte) ([]byte, error) {
	var out []byte
	i := 0

	for i < len(data) {
		if len(data[i:]) < 8 {
			return nil, fmt.Errorf("unexpected EOF in Docker log header")
		}

		streamType := data[i]
		length := binary.BigEndian.Uint32(data[i+4 : i+8])
		i += 8

		if len(data[i:]) < int(length) {
			return nil, fmt.Errorf("unexpected EOF in Docker log payload")
		}

		if streamType == 1 { // stdout
			out = append(out, data[i:i+int(length)]...)
		}

		i += int(length)
	}

	return out, nil
}

package engine

import (
	"context"
	"os"
	"path/filepath"

	"github.com/infragov-project/infrarun/internal/core/docker"
	"github.com/infragov-project/infrarun/internal/core/tools"
)

type ToolExecution struct {
	Path string
	Tool *tools.ToolDefinition
}

func NewToolExecution(tool *tools.ToolDefinition, path string) (*ToolExecution, error) {
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

	// Tempdir for output of container
	outputDir, err := os.MkdirTemp("", "infrarun-")

	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(outputDir)

	err = engine.Backend.RunContainer(ctx, docker.ContainerInfo{
		Image: toolExecution.Tool.Image,
		Cmd:   toolExecution.Tool.Cmd,
		VolumeBinds: []docker.VolumeBind{
			{Host: outputDir, Guest: toolExecution.Tool.OutputPath},
			{Host: toolExecution.Path, Guest: toolExecution.Tool.InputPath},
		},
	})

	if err != nil {
		return nil, err
	}

	// TODO: allow post processing since some tools might not export to SARIF
	outputFilePath := filepath.Clean(outputDir + "/" + toolExecution.Tool.OutputFile)

	return os.ReadFile(outputFilePath)
}

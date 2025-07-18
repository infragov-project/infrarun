package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/infragov-project/infrarun/internal/core/docker"
	"github.com/infragov-project/infrarun/internal/core/tools"
)

func main() {
	definitions := tools.GetEmbedToolDefinitions()

	for _, def := range definitions {
		fmt.Println(def)
	}

	eng, err := docker.NewDockerEngine()

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	err = eng.EnsureImageExists(ctx, "checkmarx/kics:latest")

	if err != nil {
		panic(err)
	}

	cwdAbs, err := filepath.Abs(".")

	if err != nil {
		panic(err)
	}

	err = eng.RunContainer(ctx, docker.ContainerInfo{
		Image: "checkmarx/kics:latest",
		Cmd:   []string{"scan", "-p", "/input", "-o", "/output"},
		VolumeBinds: []docker.VolumeBind{docker.VolumeBind{
			Host:  cwdAbs,
			Guest: "/input",
		}},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")

}

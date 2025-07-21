package main

import (
	"context"
	"fmt"
	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/internal/core/tools"
)

func main() {
	definitions := tools.GetEmbedToolDefinitions()

	for _, def := range definitions {
		fmt.Println(def)
	}

	eng, err := engine.NewInfrarunEngine()

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	tool := tools.GetEmbedToolDefinitions()["KICS"]

	execution, err := engine.NewToolExecution(&tool, ".")

	if err != nil {
		panic(err)
	}

	content, err := eng.Execute(ctx, execution)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}

package main

import (
	"fmt"

	"github.com/infragov-project/infrarun/internal/core/definition"
)

func main() {
	definitions := definition.GetEmbedToolDefinitions()

	for _, def := range definitions {
		fmt.Println(def)
	}
}

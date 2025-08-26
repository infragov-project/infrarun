package tools

import (
	"embed"
	"io/fs"
)

//go:embed definitions/*.yaml
var embedTools embed.FS

func GetEmbedToolDefinitions() map[string]Tool {
	tools := make(map[string]Tool)

	fileNames, err := fs.Glob(embedTools, "definitions/*.yaml")

	if err != nil {
		return tools
	}

	for _, fileName := range fileNames {
		yamlContent, err := embedTools.ReadFile(fileName)

		if err != nil {
			continue
		}

		tl, err := ToolFromYaml(yamlContent)

		if err != nil {
			continue
		}

		tools[tl.Name] = *tl
	}

	return tools
}

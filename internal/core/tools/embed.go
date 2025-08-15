package tools

import (
	"embed"
	"io/fs"

	"gopkg.in/yaml.v3"
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

		var definition toolDefinition

		err = yaml.Unmarshal(yamlContent, &definition)

		if err != nil {
			continue
		}

		tool, err := toolFromDefinition(definition)

		if err != nil {
			continue
		}

		tools[tool.Name] = *tool
	}

	return tools
}

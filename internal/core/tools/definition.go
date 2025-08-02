package tools

import (
	"embed"
	"io/fs"

	"gopkg.in/yaml.v3"
)

type ToolDefinition struct {
	Name       string   `yaml:"name"`
	Image      string   `yaml:"image"`
	Cmd        []string `yaml:"cmd"`
	InputPath  string   `yaml:"input_path"`
	OutputPath string   `yaml:"output_path"`
	OutputFile string   `yaml:"output_file"`
	Parser     string   `yaml:"parser"`
}

//go:embed definitions/*.yaml
var embedTools embed.FS

func GetEmbedToolDefinitions() map[string]ToolDefinition {
	definitions := make(map[string]ToolDefinition)

	fileNames, err := fs.Glob(embedTools, "definitions/*.yaml")

	if err != nil {
		return definitions
	}

	for _, fileName := range fileNames {
		yamlContent, err := embedTools.ReadFile(fileName)

		if err != nil {
			continue
		}

		var definition ToolDefinition

		err = yaml.Unmarshal(yamlContent, &definition)

		if err != nil {
			continue
		}

		definitions[definition.Name] = definition
	}

	return definitions
}

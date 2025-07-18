package tools

import (
	"embed"
	"gopkg.in/yaml.v3"
	"io/fs"
)

type toolDefinition struct {
	Name       string   `yaml:"name"`
	Image      string   `yaml:"image"`
	Cmd        []string `yaml:"cmd"`
	InputPath  string   `yaml:"input_path"`
	OutputPath string   `yaml:"output_path"`
	OutputFile string   `yaml:"output_file"`
}

//go:embed definitions/*.yaml
var embedTools embed.FS

func GetEmbedToolDefinitions() []toolDefinition {
	fileNames, err := fs.Glob(embedTools, "definitions/*.yaml")

	var definitions []toolDefinition

	if err != nil {
		return []toolDefinition{}
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

		definitions = append(definitions, definition)
	}

	return definitions
}

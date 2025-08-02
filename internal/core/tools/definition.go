package tools

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/infragov-project/infrarun/internal/core/results"
	"gopkg.in/yaml.v3"
)

type outputWrapper struct {
	Output outputSystem
}

type outputSystem interface {
	isOutputSystem()
}

type stdoutOutput struct {
	Type string `yaml:"type"`
}

func (stdoutOutput) isOutputSystem() {}

type fileOutput struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
	File string `yaml:"file"`
}

func (fileOutput) isOutputSystem() {}

// Custom parser for tagged union style object
func (w *outputWrapper) UnmarshalYAML(value *yaml.Node) error {
	var typeDetector struct {
		Type string `yaml:"type"`
	}

	if err := value.Decode(&typeDetector); err != nil {
		return err
	}

	switch typeDetector.Type {
	case "stdout":
		var stdout stdoutOutput
		if err := value.Decode(&stdout); err != nil {
			return err
		}
		w.Output = stdout

	case "file":
		var file fileOutput
		if err := value.Decode(&file); err != nil {
			return err
		}
		w.Output = file

	default:
		return fmt.Errorf("unknown output type: %s", typeDetector.Type)
	}

	return nil
}

type toolDefinition struct {
	Name      string        `yaml:"name"`
	Image     string        `yaml:"image"`
	Cmd       []string      `yaml:"cmd"`
	InputPath string        `yaml:"input_path"`
	Output    outputWrapper `yaml:"output"`
	Parser    string        `yaml:"parser"`
}

type Tool struct {
	Name          string
	Image         string
	Cmd           []string
	InputPath     string
	OutputPath    string
	OutputFile    string
	CaptureStdout bool // Will ignore OutputPath and OutputFile if true, since it uses stdout
	Parser        results.ResultParser
}

func toolFromDefinition(definition toolDefinition) (*Tool, error) {
	t := &Tool{
		Name:      definition.Name,
		Image:     definition.Image,
		Cmd:       definition.Cmd,
		InputPath: definition.InputPath,
	}

	parser, err := results.GetParser(definition.Parser)

	if err != nil {
		return nil, err
	}

	t.Parser = parser

	switch out := definition.Output.Output.(type) {
	case stdoutOutput:
		t.CaptureStdout = true
	case fileOutput:
		t.CaptureStdout = false
		t.OutputPath = out.Path
		t.OutputFile = out.File
	default:
		return nil, fmt.Errorf("unknown output type")
	}

	return t, nil

}

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

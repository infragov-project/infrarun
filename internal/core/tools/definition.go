package tools

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

func ToolFromYaml(content []byte) (*Tool, error) {
	var def toolDefinition

	err := yaml.Unmarshal(content, &def)

	if err != nil {
		return nil, err
	}

	return toolFromDefinition(def)
}

type toolDefinition struct {
	Name           string                     `yaml:"name"`
	Image          string                     `yaml:"image"`
	Cmd            []string                   `yaml:"cmd"`
	InputPath      string                     `yaml:"input_path"`
	Output         outputWrapper              `yaml:"output"`
	Parser         string                     `yaml:"parser"`
	OutputMappings []outputMappingDeffinition `yaml:"output_mapping"`
}

func toolFromDefinition(definition toolDefinition) (*Tool, error) {
	t := &Tool{
		Name:      definition.Name,
		Image:     definition.Image,
		Cmd:       definition.Cmd,
		InputPath: definition.InputPath,
	}

	for _, omDef := range definition.OutputMappings {
		om, err := outputMappingFromDefinition(omDef)

		if err != nil {
			return nil, err
		}

		t.outputMappings = append(t.outputMappings, *om)
	}

	parser, err := GetParser(definition.Parser)

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

type outputMappingDeffinition struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
}

func outputMappingFromDefinition(definition outputMappingDeffinition) (*OutputMapping, error) {
	re, err := regexp.Compile(definition.Pattern)

	if err != nil {
		return nil, err
	}

	return &OutputMapping{
		Pattern:     *re,
		Replacement: definition.Replacement,
	}, nil
}

// Wrapper struct to allow custom yaml parsing for tagged union style objects
type outputWrapper struct {
	Output outputSystem
}

type outputSystem interface {
	isOutputSystem() // Dummy method just to restrict the interface to a specific set of types
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

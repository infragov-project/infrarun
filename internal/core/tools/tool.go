package tools

import "regexp"

type Tool struct {
	Name           string
	Image          string
	Cmd            []string
	InputPath      string
	OutputPath     string
	OutputFile     string
	CaptureStdout  bool // Will ignore OutputPath and OutputFile if true, since it uses stdout
	Parser         ResultParser
	outputMappings []OutputMapping
}

type OutputMapping struct {
	Pattern     regexp.Regexp
	Replacement string
}

func (t Tool) ApplyOutputMappings(path string) string {
	for _, mapping := range t.outputMappings {
		if mapping.Pattern.MatchString(path) {
			return mapping.Pattern.ReplaceAllString(path, mapping.Replacement)
		}
	}

	return path
}

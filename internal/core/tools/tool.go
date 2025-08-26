package tools

import "regexp"

type Tool struct {
	Name                string
	Image               string
	Cmd                 []string
	InputPath           string
	OutputPath          string
	OutputFile          string
	CaptureStdout       bool // Will ignore OutputPath and OutputFile if true, since it uses stdout
	Parser              ResultParser
	pathTransformations []PathTransformation
}

type PathTransformation struct {
	Pattern     regexp.Regexp
	Replacement string
}

func (pt PathTransformation) Apply(path string) (string, bool) {
	if pt.Pattern.MatchString(path) {
		return pt.Pattern.ReplaceAllString(path, pt.Replacement), true
	}

	return path, false
}

func (t Tool) ApplyPathTransformations(path string) string {
	for _, transformation := range t.pathTransformations {
		if new, matched := transformation.Apply(path); matched {
			return new
		}
	}

	return path
}

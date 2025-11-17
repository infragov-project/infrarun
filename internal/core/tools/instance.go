package tools

import (
	"fmt"
	"regexp"
	"strings"
)

type ToolInstance struct {
	Name                string
	Image               string
	Cmd                 []string
	InputPath           string
	OutputPath          string
	OutputFile          string
	CaptureStdout       bool // Will ignore OutputPath and OutputFile if true, since it uses stdout
	Parser              ResultParser
	PathTransformations []PathTransformation
}

func (t *Tool) DefaultInstance() (*ToolInstance, error) {
	newCmd, err := patternFill(t.Cmd, t.defaultValues)

	if err != nil {
		return nil, err
	}

	return &ToolInstance{
		Name:                t.Name,
		Image:               t.Image,
		Cmd:                 newCmd,
		InputPath:           t.InputPath,
		OutputPath:          t.OutputPath,
		OutputFile:          t.OutputFile,
		CaptureStdout:       t.CaptureStdout,
		Parser:              t.Parser,
		PathTransformations: t.pathTransformations,
	}, nil
}

func (t *Tool) ToInstance(params map[string]any) (*ToolInstance, error) {
	fullParams := addDefaults(params, t.defaultValues)

	// TODO: allow other parts of tool definition to be parameterizable
	newCmd, err := patternFill(t.Cmd, fullParams)

	if err != nil {
		return nil, err
	}

	return &ToolInstance{
		Name:                t.Name,
		Image:               t.Image,
		Cmd:                 newCmd,
		InputPath:           t.InputPath,
		OutputPath:          t.OutputPath,
		OutputFile:          t.OutputFile,
		CaptureStdout:       t.CaptureStdout,
		Parser:              t.Parser,
		PathTransformations: t.pathTransformations,
	}, nil
}

var placeholderPattern = regexp.MustCompile("%{([a-zA-Z0-9_]+)}")

func patternFill(template []string, values map[string]any) ([]string, error) {
	var result []string

	for _, elem := range template {
		if name, ok := isIsolatedPlaceholder(elem); ok {
			val, exists := values[name]

			if !exists {
				return nil, fmt.Errorf("undefined placeholder value %q", name)
			}

			switch v := val.(type) {
			case string:
				result = append(result, v)
			case []string:
				result = append(result, v...)
			default:
				return nil, fmt.Errorf("unsupported type for placeholder value %q", name)
			}

			continue
		}

		// Inline case

		out := elem
		matches := placeholderPattern.FindAllStringSubmatch(out, -1)
		// Single pass replacement (no way to recursively fill placeholders, so it guarantees termination)
		for _, m := range matches {
			full := m[0]
			name := m[1]

			val, exists := values[name]

			if !exists {
				return nil, fmt.Errorf("undefined placeholder value %q", name)
			}

			switch v := val.(type) {
			case string:
				out = strings.ReplaceAll(out, full, v)
			case []string:
				return nil, fmt.Errorf("array substitution not allowed in non-isolated line: %q (placeholder: %q)", elem, name)
			default:
				return nil, fmt.Errorf("unsupported type for placeholder value %q", name)
			}

		}

		result = append(result, out)
	}

	return result, nil
}

func isIsolatedPlaceholder(line string) (string, bool) {
	// For line to be an isolated placeholder, then the following must hold:
	// 1. FindStringSubmatch with the placeholder pattern must find a
	// 	  match (checked by asserting that m is not nil).
	// 2. The first element of m (whole first match) must be equal to the whole line
	// 3. The placeholder pattern contains atleast one capture group, that will be interpreted as
	//    representing the placeholder name
	//
	// This guarantees (for the given placeholder pattern) the following:
	// 1. m has a second element that corresponds to the result of the first capture group
	// 2. this element is the name of the placeholder

	m := placeholderPattern.FindStringSubmatch(line)

	if m != nil && m[0] == line {
		return m[1], true
	}

	return "", false
}

// addDefaults takes two map[string]any: values and def (short for default) and returns
// a new map[string]any that is contains all key-value pairs in values aswell as
// all pairs contained in def, appart from the ones which the key is present in values.
// In short, addDefaults returns a map with the same elements as values, but with the added
// (if needed) default values present in def.
func addDefaults(values map[string]any, def map[string]any) map[string]any {
	result := values

	for k, v := range def {
		_, ok := result[k]

		if !ok {
			result[k] = v
		}
	}

	return result
}

package tools

import (
	"fmt"

	"github.com/infragov-project/infrarun/internal/core/tools/parsers/glitch"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

type ResultParser func([]byte) (*sarif.Report, error)

var parsers = map[string]ResultParser{
	"sarif":  parseJsonSARIF,
	"glitch": glitch.ParseGlitch,
	"kics":   kicsParser,
}

func GetParser(name string) (ResultParser, error) {
	parser, ok := parsers[name]

	if !ok {
		return nil, fmt.Errorf("parser not found")
	}

	return parser, nil
}

func parseJsonSARIF(data []byte) (*sarif.Report, error) {
	parsed, err := sarif.FromBytes(data)

	if err != nil {
		return nil, err
	}

	return parsed, nil
}

// Temporary parser for KICS, since it doesn't fill the level of each result
// and instead puts and empty string. Some SARIF consumers consider that
// invalid SARIF.
func kicsParser(data []byte) (*sarif.Report, error) {
	parsed, err := parseJsonSARIF(data)

	if err != nil {
		return nil, err
	}

	for _, run := range parsed.Runs {
		for _, res := range run.Results {
			res.Level = "warning"
		}
	}

	return parsed, nil
}

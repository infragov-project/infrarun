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

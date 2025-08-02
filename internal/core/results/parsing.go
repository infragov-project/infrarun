package results

import (
	"fmt"

	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

type ResultParser func([]byte) (*sarif.Report, error)

var parsers = map[string]ResultParser{
	"sarif": parseJsonSARIF,
}

func ParseResults(data []byte, parserName string) (*sarif.Report, error) {
	parser, ok := parsers[parserName]

	if !ok {
		return nil, fmt.Errorf("result parsing error: parser not found")
	}

	report, err := parser(data)

	if err != nil {
		return nil, fmt.Errorf("result parsing error: %w", err)
	}

	return report, nil
}

func parseJsonSARIF(data []byte) (*sarif.Report, error) {
	parsed, err := sarif.FromBytes(data)

	if err != nil {
		return nil, err
	}

	return parsed, nil
}

package glitch

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/owenrumney/go-sarif/v3/pkg/report"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

type result struct {
	Description string
	FilePath    string
	Line        int
	Name        string
	Context     string
	Note        string
}

func parseResult(line string) (*result, error) {

	split := strings.SplitN(line, ",", 7)

	if len(split) < 6 {
		return nil, fmt.Errorf("invalid line: %s", line)
	}

	lineNo, err := strconv.Atoi(split[2])

	if err != nil {
		return nil, fmt.Errorf("invalid line: %w", err)
	}

	return &result{
		Description: split[0],
		FilePath:    split[1],
		Line:        lineNo,
		Name:        split[3],
		Context:     split[4],
		Note:        split[5],
	}, nil

}

func parseReport(rep string) ([]result, error) {
	lines := strings.Split(rep, "\n")

	res := make([]result, 0)

	for _, l := range lines {
		r, err := parseResult(l)

		if err != nil {
			continue
		}

		res = append(res, *r)
	}

	return res, nil
}

func toSarif(results []result) (*sarif.Report, error) {
	rep := report.NewV210Report()

	run := sarif.NewRunWithInformationURI("GLITCH", "https://github.com/sr-lab/GLITCH")

	for _, r := range results {

		u := url.URL{
			Scheme: "file",
			Path:   r.FilePath,
		}

		run.AddDistinctArtifact(u.String())

		run.AddRule(r.Name).WithDescription(r.Description)

		run.CreateResultForRule(r.Name).
			WithLevel("warning").
			WithMessage(sarif.NewTextMessage(r.Context)).
			AddLocation(
				sarif.NewLocationWithPhysicalLocation(
					sarif.NewPhysicalLocation().
						WithArtifactLocation(sarif.NewSimpleArtifactLocation("file://" + r.FilePath)).
						WithRegion(sarif.NewSimpleRegion(r.Line, r.Line)),
				),
			)

	}

	rep.AddRun(run)

	return rep, rep.Validate()
}

func ParseGlitch(data []byte) (*sarif.Report, error) {
	res, err := parseReport(string(data))

	if err != nil {
		return nil, err
	}

	return toSarif(res)
}

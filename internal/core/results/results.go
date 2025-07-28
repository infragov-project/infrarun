package results

import (
	"net/url"
	"strings"

	"github.com/owenrumney/go-sarif/v3/pkg/report"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

func MergeReports(reports []*sarif.Report) *sarif.Report {
	merged := report.NewV210Report()

	for _, rep := range reports {
		for _, run := range rep.Runs {
			rep.AddRun(run)
		}
	}

	return merged
}

func ParseReport(content []byte) (*sarif.Report, error) {
	parsed, err := sarif.FromBytes(content)

	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func ReplaceFilePaths(report *sarif.Report, prefixMap map[string]string) {
	for _, run := range report.Runs {

		for _, base := range run.OriginalUriBaseIds {
			if base.URI != nil {
				processPath(base.URI, prefixMap)
			}
		}

		for _, artifact := range run.Artifacts {
			if artifact.Location != nil && artifact.Location.URI != nil {
				processPath(artifact.Location.URI, prefixMap)
			}
		}

		for _, res := range run.Results {
			for _, loc := range res.Locations {
				if loc.PhysicalLocation != nil && loc.PhysicalLocation.ArtifactLocation != nil && loc.PhysicalLocation.ArtifactLocation.URI != nil {
					processPath(loc.PhysicalLocation.ArtifactLocation.URI, prefixMap)
				}
			}

			for _, relLoc := range res.RelatedLocations {
				if relLoc.PhysicalLocation != nil && relLoc.PhysicalLocation.ArtifactLocation != nil && relLoc.PhysicalLocation.ArtifactLocation.URI != nil {
					processPath(relLoc.PhysicalLocation.ArtifactLocation.URI, prefixMap)
				}
			}

			for _, fix := range res.Fixes {
				for _, change := range fix.ArtifactChanges {
					if change.ArtifactLocation != nil && change.ArtifactLocation.URI != nil {
						processPath(change.ArtifactLocation.URI, prefixMap)
					}
				}
			}
		}

	}
}

func processPath(path *string, prefixMap map[string]string) {
	if path == nil {
		return
	}

	if strings.HasPrefix(*path, "file://") {
		u, err := url.Parse(*path)

		if err != nil {
			return
		}

		decodedPath, err := url.PathUnescape(u.Path)

		if err != nil {
			return
		}

		replacePrefix(&decodedPath, prefixMap)

		u.Path = url.PathEscape(decodedPath)

		*path = u.String()
	} else {
		replacePrefix(path, prefixMap)
	}
}

func replacePrefix(uri *string, prefixMap map[string]string) {

	for prefix, replacement := range prefixMap {
		if strings.HasPrefix(*uri, prefix) {
			*uri = replacement + strings.TrimPrefix(*uri, prefix)
			return
		}
	}
}

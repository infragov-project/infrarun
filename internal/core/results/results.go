package results

import (
	"net/url"
	"strings"

	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/owenrumney/go-sarif/v3/pkg/report"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

func MergeReports(reports []*sarif.Report) *sarif.Report {
	merged := report.NewV210Report()

	for _, rep := range reports {
		for _, run := range rep.Runs {
			merged.AddRun(run)
		}
	}

	return merged
}

func ReplaceFilePaths(report *sarif.Report, tool *tools.Tool) {
	for _, run := range report.Runs {

		for _, base := range run.OriginalUriBaseIds {
			if base.URI != nil {
				processPath(base.URI, tool)
			}
		}

		for _, artifact := range run.Artifacts {
			if artifact.Location != nil && artifact.Location.URI != nil {
				processPath(artifact.Location.URI, tool)
			}
		}

		for _, res := range run.Results {
			for _, loc := range res.Locations {
				if loc.PhysicalLocation != nil && loc.PhysicalLocation.ArtifactLocation != nil && loc.PhysicalLocation.ArtifactLocation.URI != nil {
					processPath(loc.PhysicalLocation.ArtifactLocation.URI, tool)
				}
			}

			for _, relLoc := range res.RelatedLocations {
				if relLoc.PhysicalLocation != nil && relLoc.PhysicalLocation.ArtifactLocation != nil && relLoc.PhysicalLocation.ArtifactLocation.URI != nil {
					processPath(relLoc.PhysicalLocation.ArtifactLocation.URI, tool)
				}
			}

			for _, fix := range res.Fixes {
				for _, change := range fix.ArtifactChanges {
					if change.ArtifactLocation != nil && change.ArtifactLocation.URI != nil {
						processPath(change.ArtifactLocation.URI, tool)
					}
				}
			}
		}

	}
}

func processPath(path *string, tool *tools.Tool) {
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

		u.Path = tool.ApplyPathTransformations(decodedPath)

		*path = u.String()
	} else {
		*path = tool.ApplyPathTransformations(*path)
	}
}

func GenerateFinalReport(reports map[*tools.Tool]sarif.Report) *sarif.Report {

	newReps := make([]*sarif.Report, 0)

	for t, r := range reports {
		ReplaceFilePaths(&r, t)
		newReps = append(newReps, &r)
	}

	return MergeReports(newReps)

}

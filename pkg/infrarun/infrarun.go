package infrarun

import (
	"context"
	"sync"

	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/internal/core/results"
	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

// A Tool represents the concept of a code analysis tool running inside a [Docker] container,
// together with a parser that converts its output into [SARIF].
//
// [Docker]: https://www.docker.com/
// [SARIF]: https://sarifweb.azurewebsites.net/
type Tool struct {
	impl *tools.Tool
}

// Name returns the display name of the given tool.
func (t Tool) Name() string {
	return t.impl.Name
}

// Image returns the [Docker image reference] of the image used by the given tool.
//
// [Docker image reference]: https://docs.docker.com/reference/cli/docker/image/tag/#description
func (t Tool) Image() string {
	return t.impl.Image
}

func toolFromImpl(impl *tools.Tool) Tool {
	return Tool{impl}
}

type ToolInstance struct {
	impl *tools.ToolInstance
}

func toolInstanceFromImpl(impl *tools.ToolInstance) ToolInstance {
	return ToolInstance{impl}
}

// GetAvailableTools returns a map with all the infrarun [Tool] available in the current process.
// This map has the [Tool]'s display name as the keys.
func GetAvailableTools() map[string]Tool {
	impls := tools.GetEmbedToolDefinitions()

	t := make(map[string]Tool)

	for k, v := range impls {
		t[k] = toolFromImpl(&v)
	}

	return t
}

// RunTools returns a [SARIF] report with the outputs of the execution of all tools in toolList when running inside path.
// In case something fails, it will return a nil report with a non-nil error.
//
// RunTools requires a currently running [Docker engine]. These tools will be called in parallel, with the paralelization left to the engine.
//
// [SARIF]: https://sarifweb.azurewebsites.net/
//
// [Docker Engine]: https://docs.docker.com/engine/
func RunTools(toolList []*ToolInstance, path string) (*sarif.Report, error) {
	ctx := context.Background()

	eng, err := engine.NewInfrarunEngine()

	if err != nil {
		return nil, err
	}

	execs := make([]*engine.ToolExecution, 0)

	for _, t := range toolList {
		exec, err := engine.NewToolExecution(t.impl, path, "**/*")

		if err != nil {
			return nil, err
		}

		execs = append(execs, exec)

	}

	var wg sync.WaitGroup

	for _, exec := range execs {
		wg.Add(1)

		go func() {
			defer wg.Done()

			content, err := eng.Execute(ctx, exec)

			if err != nil {
				exec.Err = err
				return
			}

			report, err := exec.Tool.Parser(content)

			if err != nil {
				exec.Err = err
				return
			}

			exec.Report = report

		}()
	}

	wg.Wait()

	reports := make(map[*tools.ToolInstance]sarif.Report)

	for _, exec := range execs {
		if exec.Err != nil {
			return nil, err
		}

		reports[exec.Tool] = *exec.Report
	}

	finalReport := results.GenerateFinalReport(reports)

	return finalReport, nil
}

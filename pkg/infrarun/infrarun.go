package infrarun

import (
	"context"
	"sync"

	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/internal/core/results"
	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

type Tool struct {
	impl *tools.Tool
}

func (t Tool) Name() string {
	return t.impl.Name
}

func (t Tool) Image() string {
	return t.impl.Image
}

func toolFromImpl(impl *tools.Tool) Tool {
	return Tool{impl}
}

func GetAvailableTools() map[string]Tool {
	impls := tools.GetEmbedToolDefinitions()

	t := make(map[string]Tool)

	for k, v := range impls {
		t[k] = toolFromImpl(&v)
	}

	return t
}

func RunTools(toolList []*Tool, path string) (*sarif.Report, error) {
	ctx := context.Background()

	eng, err := engine.NewInfrarunEngine()

	if err != nil {
		return nil, err
	}

	execs := make([]*engine.ToolExecution, 0)

	for _, t := range toolList {
		exec, err := engine.NewToolExecution(t.impl, path)

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

	reports := make(map[*tools.Tool]sarif.Report)

	for _, exec := range execs {
		if exec.Err != nil {
			return nil, err
		}

		reports[exec.Tool] = *exec.Report
	}

	finalReport := results.GenerateFinalReport(reports)

	return finalReport, nil
}

package run

import (
	"context"
	"sync"

	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/internal/core/results"
	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/infragov-project/infrarun/pkg/infrarun/plan"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

// Warning: Observer methods might be called in concurrent environment, watch for race conditions
type RunObserver interface {
	OnPreparation()
	OnRunStart(run *plan.Run)
	OnRunParse(run *plan.Run)
	OnRunCompletion(run *plan.Run, report *sarif.Report, err error)
}

type Option func(*runConfig)

type runConfig struct {
	observer RunObserver
}

func WithObserver(obs *RunObserver) Option {
	return func(opt *runConfig) {
		opt.observer = *obs
	}
}

func defaultRunConfig() runConfig {
	return runConfig{
		observer: emptyRunObserver{},
	}
}

type emptyRunObserver struct{}

func (o emptyRunObserver) OnPreparation() {}

func (o emptyRunObserver) OnRunStart(run *plan.Run) {}

func (o emptyRunObserver) OnRunCompletion(run *plan.Run, report *sarif.Report, err error) {}

func (o emptyRunObserver) OnRunParse(run *plan.Run) {}

// Run returns a [SARIF] report with the outputs of the execution of all tools in toolList when running inside path.
// In case something fails, it will return a nil report with a non-nil error.
//
// RunTools requires a currently running [Docker engine]. These tools will be called in parallel, with the paralelization left to the engine.
//
// [SARIF]: https://sarifweb.azurewebsites.net/
//
// [Docker Engine]: https://docs.docker.com/engine/
func Run(ctx context.Context, plan plan.Plan, opts ...Option) (*sarif.Report, error) {
	config := defaultRunConfig()

	for _, opt := range opts {
		opt(&config)
	}

	config.observer.OnPreparation()

	eng, err := engine.NewInfrarunEngine()

	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	for _, run := range plan.Runs {
		exec := run.Impl
		wg.Add(1)

		go func() {
			defer wg.Done()

			config.observer.OnRunStart(&run)
			content, err := eng.Execute(ctx, exec)

			if err != nil {
				exec.Err = err
				config.observer.OnRunCompletion(&run, nil, err)
				return
			}

			config.observer.OnRunParse(&run)
			report, err := exec.Tool.Parser(content)

			if err != nil {
				exec.Err = err
				config.observer.OnRunCompletion(&run, nil, err)
				return
			}

			exec.Report = report
			config.observer.OnRunCompletion(&run, report, nil)
		}()
	}

	wg.Wait()

	// TODO: see if this still works now that you can have more than one run per tool

	reports := make(map[*tools.ToolInstance]sarif.Report)

	for _, run := range plan.Runs {
		if run.Impl.Err != nil {
			continue
		}

		reports[run.Impl.Tool] = *run.Impl.Report
	}

	finalReport := results.GenerateFinalReport(reports)

	return finalReport, nil
}

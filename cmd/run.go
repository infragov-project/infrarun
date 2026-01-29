package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/infragov-project/infrarun/pkg/infrarun/plan"
	"github.com/infragov-project/infrarun/pkg/infrarun/run"
	"github.com/infragov-project/infrarun/pkg/infrarun/tool"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type filler struct{}

func (f filler) Fill(w io.Writer, width int, stat *decor.Statistics) {
	completed := int(float64(width) * float64(stat.Current) / float64(stat.Total))
	for range completed {
		_, err := fmt.Fprint(w, "\033[32m─\033[0m") // green

		if err != nil {
			return
		}
	}
	// Fill the rest with empty blocks
	for i := completed; i < width; i++ {
		_, err := fmt.Fprint(w, "─")

		if err != nil {
			return
		}
	}
}

func runRun(cmd *cobra.Command, args []string) {
	t := tool.GetAvailableTools()

	path, err := cmd.Flags().GetString("path")

	if err != nil {
		panic(err)
	}

	var p plan.Plan

	for _, toolName := range args {
		tool, ok := t[toolName]

		if !ok {
			panic("tool not found: " + toolName)
		}

		run, err := plan.NewSimpleRun(path, &tool)

		if err != nil {
			panic(err)
		}

		p.AddRun(run)
	}

	ctx := context.Background()

	obs := newObserver(&p)

	rep, err := run.Run(ctx, p, run.WithObserver(obs))

	if err != nil {
		panic(err)
	}

	obs.progress.Wait()

	err = printRaw(rep, os.Stdout)

	if err != nil {
		panic(err)
	}
}

type progressBarObserver struct {
	mutex    *sync.Mutex
	progress *mpb.Progress
	plan     *plan.Plan
	bars     map[*plan.Run]*mpb.Bar
}

func newObserver(p *plan.Plan) *progressBarObserver {
	return &progressBarObserver{
		mutex:    &sync.Mutex{},
		progress: mpb.New(mpb.WithOutput(os.Stderr)),
		plan:     p,
		bars:     make(map[*plan.Run]*mpb.Bar),
	}
}

func (o *progressBarObserver) OnEnginePreparation() {
	o.mutex.Lock()
	for _, r := range o.plan.Runs {
		o.bars[r] = o.progress.Add(
			100,
			filler{},
			mpb.PrependDecorators(decor.Name(r.ToolName()+"\t")),
			mpb.AppendDecorators(decor.Percentage()),
		)
		//o.bars[r] = o.progress.AddBar(100, mpb.PrependDecorators(decor.Name(r.ToolName())), mpb.AppendDecorators(decor.Percentage()))
	}
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnEngineFailure(err error) {
	o.mutex.Lock()
	for _, r := range o.plan.Runs {
		bar, ok := o.bars[r]

		if !ok {
			continue
		}

		bar.SetTotal(100, true)
	}

	panic(err)
}

func (o *progressBarObserver) OnRunStart(run *plan.Run) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.IncrBy(5)
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnRunCompletion(run *plan.Run, report *sarif.Report) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.SetTotal(100, true)
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnRunFail(run *plan.Run, err error) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.SetTotal(100, true)
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnRunParseFail(run *plan.Run, err error) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.SetTotal(100, true)
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnRunParse(run *plan.Run) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.IncrBy(70)
	o.mutex.Unlock()
}

type ReportPrinter func(rep *sarif.Report, writer io.Writer) error

func printRaw(rep *sarif.Report, writer io.Writer) error {
	return rep.PrettyWrite(writer)
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run one or more tools on a given directory",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Runs the tools passed as arguments in the current directory, then merges and presents the results as SARIF.`,
	Run:   runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("path", "p", ".", "path to run the tools at")
}

package cmd

import (
	"context"
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

	err = rep.PrettyWrite(os.Stdout)

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

func (o *progressBarObserver) OnPreparation() {
	o.mutex.Lock()
	for _, r := range o.plan.Runs {
		o.bars[r] = o.progress.AddBar(100, mpb.PrependDecorators(decor.Name(r.ToolName())), mpb.AppendDecorators(decor.Percentage()))
	}
	o.mutex.Unlock()
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

func (o *progressBarObserver) OnRunCompletion(run *plan.Run, report *sarif.Report, err error) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.IncrBy(80)
	o.mutex.Unlock()
}

func (o *progressBarObserver) OnRunParse(run *plan.Run) {
	o.mutex.Lock()
	bar, ok := o.bars[run]

	if !ok {
		return
	}

	bar.IncrBy(15)
	o.mutex.Unlock()
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

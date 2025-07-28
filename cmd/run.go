package cmd

import (
	"context"
	"os"
	"sync"

	"github.com/infragov-project/infrarun/internal/core/engine"
	"github.com/infragov-project/infrarun/internal/core/results"
	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func runRun(cmd *cobra.Command, args []string) {
	t := tools.GetEmbedToolDefinitions()

	eng, err := engine.NewInfrarunEngine()

	ctx := context.Background()

	path, _ := cmd.Flags().GetString("path")

	if err != nil {
		panic(err)
	}

	execs := make([]*engine.ToolExecution, 0)

	for _, toolName := range args {
		tool, ok := t[toolName]

		if !ok {
			panic("tool not found")
		}

		exec, err := engine.NewToolExecution(&tool, path)

		if err != nil {
			panic(err)
		}

		execs = append(execs, exec)
	}

	resultCh := make(chan *sarif.Report, len(execs))
	errCh := make(chan error, len(execs))
	progressBars := mpb.New()
	var wg sync.WaitGroup

	for _, exec := range execs {
		wg.Add(1)
		pBar := progressBars.AddBar(100, mpb.PrependDecorators(decor.Name(exec.Tool.Name)))

		go func() {
			defer wg.Done()
			defer pBar.SetTotal(100, true)

			pBar.IncrBy(5)

			content, err := eng.Execute(ctx, exec)

			if err != nil {
				errCh <- err
				return
			}

			pBar.IncrBy(80)

			rep, err := results.ParseReport(content)

			if err != nil {
				errCh <- err
				return
			}

			pBar.IncrBy(15)

			resultCh <- rep
		}()
	}

	wg.Wait()
	progressBars.Wait()

	close(resultCh)

	resultList := make([]*sarif.Report, 0)

	for result := range resultCh {
		resultList = append(resultList, result)
	}

	finalReport := results.MergeReports(resultList)

	finalReport.PrettyWrite(os.Stdout)

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

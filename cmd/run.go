package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/infragov-project/infrarun/pkg/infrarun/plan"
	"github.com/infragov-project/infrarun/pkg/infrarun/run"
	"github.com/infragov-project/infrarun/pkg/infrarun/tool"
	"github.com/olekukonko/tablewriter"
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

	err = prettyPrint(rep, os.Stdout)

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

func prettyPrint(rep *sarif.Report, writer io.Writer) error {
	for _, run := range rep.Runs {
		if len(run.Results) == 0 {
			continue
		}

		tool := run.Tool.Driver.FullName
		if tool == nil {
			tool = run.Tool.Driver.Name
		}
		fmt.Fprintf(writer, "\n\n\n%s\n", *tool)

		version := run.Tool.Driver.Version
		if version != nil {
			fmt.Fprintf(writer, "%s\n", *version)
		}

		uri := run.Tool.Driver.InformationURI
		if uri != nil {
			fmt.Fprintf(writer, "Information URI: %s\n", *uri)
		}

		fmt.Fprintf(writer, "\n")
		table := tablewriter.NewWriter(writer)
		table.Header("File", "Start Line", "End Line", "Rule ID", "Message")

		// A list of valid and sorted results
		processed := run.Results[:0]
		for _, result := range run.Results {
			// File location
			if len(result.Locations) == 0 || result.Locations[0].PhysicalLocation == nil || result.Locations[0].PhysicalLocation.ArtifactLocation == nil || result.Locations[0].PhysicalLocation.ArtifactLocation.URI == nil {
				continue
			}
			// Start line
			if result.Locations[0].PhysicalLocation.Region == nil || result.Locations[0].PhysicalLocation.Region.StartLine == nil {
				continue
			}
			// Rule ID
			if result.RuleID == nil {
				continue
			}
			processed = append(processed, result)
		}

		sorted := run.Results
		sort.Slice(sorted, func(i, j int) bool {
			// Results are sorted by the file name and then by the start line
			fileI := *sorted[i].Locations[0].PhysicalLocation.ArtifactLocation.URI
			fileJ := *sorted[j].Locations[0].PhysicalLocation.ArtifactLocation.URI

			if fileI != fileJ {
				return fileI < fileJ
			}

			startLineI := *sorted[i].Locations[0].PhysicalLocation.Region.StartLine
			startLineJ := *sorted[j].Locations[0].PhysicalLocation.Region.StartLine

			return startLineI < startLineJ
		})

		for _, result := range sorted {
			// Get file location
			if len(result.Locations) == 0 || result.Locations[0].PhysicalLocation == nil || result.Locations[0].PhysicalLocation.ArtifactLocation == nil || result.Locations[0].PhysicalLocation.ArtifactLocation.URI == nil {
				continue
			}
			file := *result.Locations[0].PhysicalLocation.ArtifactLocation.URI

			// Get start line
			if result.Locations[0].PhysicalLocation.Region == nil || result.Locations[0].PhysicalLocation.Region.StartLine == nil {
				continue
			}
			startLine := *result.Locations[0].PhysicalLocation.Region.StartLine

			// Get end line
			endLine := -1
			if result.Locations[0].PhysicalLocation.Region.EndLine != nil {
				endLine = *result.Locations[0].PhysicalLocation.Region.EndLine
			}

			// Get rule ID
			if result.RuleID == nil {
				continue
			}
			ruleId := *result.RuleID

			// Get message
			message := ""
			if result.Message != nil && result.Message.Text != nil {
				message = *result.Message.Text
				message = truncate(removeWhitespace(message), 50)
			}

			var endLineStr string
			if endLine == -1 {
				endLineStr = "-"
			} else {
				endLineStr = fmt.Sprintf("%d", endLine)
			}

			err := table.Append(file, startLine, endLineStr, ruleId, truncate(message, 50))
			if err != nil {
				return err
			}
		}

		err := table.Render()
		if err != nil {
			return err
		}
	}
	return nil
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Removes all whitespace characters except spaces from a string
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) && r != ' ' {
			return -1
		}
		return r
	}, s)
}

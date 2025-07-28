package cmd

import (
	"os"

	"github.com/infragov-project/infrarun/internal/core/tools"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func runList(cmd *cobra.Command, args []string) {
	detailed, err := cmd.Flags().GetBool("detailed")

	if err != nil {
		panic(err)
	}

	t := tools.GetEmbedToolDefinitions()

	table := tablewriter.NewWriter(os.Stdout)

	if detailed {
		table.Header("Name", "Image")

		for _, tool := range t {
			table.Append(tool.Name, tool.Image)
		}
	} else {
		table.Header("Name")

		for _, tool := range t {
			table.Append(tool.Name)
		}
	}

	table.Render()
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available tools",
	Long: `Gives a list with tools that are available to run with the tool.
	Provides a description for each tool provided.`,
	Args: cobra.ExactArgs(0),
	Run:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("detailed", "d", false, "Provide extra details of each tool.")
}

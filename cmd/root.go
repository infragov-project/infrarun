package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "infrarun",
	Short: "An execution framework for IaC code analysis tools.",
	Long:  `Infrarun allows the user to run multiple IaC code analysis tools in batch. It also combines all results into a single report.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}

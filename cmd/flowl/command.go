package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// root command
var rootCmd = &cobra.Command{
	Use:   "flowl",
	Short: "function flow language",
	Long: `A function flow tool, used to Parse, Create, Run
	and Manage flow based on funtion`,
}

func Execute() {
	initCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initCmd() {
	{
		var showAll bool

		parseCmd := &cobra.Command{
			Use:          "parse [path of flowl file]",
			Short:        "Parse a flowl file",
			Example:      "flowl parse [-a] ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return parseFlowl(args[0], showAll)
			},
		}
		parseCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show run queue and blocks, only show run queue by default")
		rootCmd.AddCommand(parseCmd)
	}

	{
		runCmd := &cobra.Command{
			Use:          "run [path of flowl file]",
			Short:        "run a flowl file",
			Example:      "flowl run ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runFlowl(args[0])
			},
		}
		rootCmd.AddCommand(runCmd)
	}
}

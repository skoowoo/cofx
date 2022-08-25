package main

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// root command
var rootCmd = &cobra.Command{
	Use:   "cofunc",
	Short: "function fabric",
	Long:  `A function fabric tool, used to Parse, Create, Run and Manage flow based on funtion fabric`,
}

func Execute() {
	initCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initCmd() {
	{
		var showAll bool

		parseCmd := &cobra.Command{
			Use:          "parse [path of flowl file]",
			Short:        "Parse a flowl source file",
			Example:      "cofunc parse [-a] ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return parseflowl(args[0], showAll)
			},
		}
		parseCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show run queue and blocks, only show run queue by default")
		rootCmd.AddCommand(parseCmd)
	}

	{
		runCmd := &cobra.Command{
			Use:          "run [path of flowl file]",
			Short:        "Run a flowl file",
			Example:      "cofunc run ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runflowl(args[0])
			},
		}
		rootCmd.AddCommand(runCmd)
	}

	{
		logCmd := &cobra.Command{
			Use:          "log [flow id] [function seq]",
			Short:        "View the execution log of the flow or function",
			Example:      "cofunc run b0804ec967f48520697662a204f5fe72 1",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				id := args[0]
				seq, err := strconv.ParseInt(args[1], 10, 64)
				if err != nil {
					return err
				}
				return viewLog(id, int(seq))
			},
		}
		rootCmd.AddCommand(logCmd)
	}

	{
		listCmd := &cobra.Command{
			Use:          "list",
			Short:        "List all flows that you coded in the flow source directory",
			Example:      "cofunc list",
			SilenceUsage: true,
			Args:         cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return listFlows()
			},
		}
		rootCmd.AddCommand(listCmd)
	}
}

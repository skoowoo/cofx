package main

import (
	"os"
	"strconv"

	"github.com/cofunclabs/cofunc/pkg/nameid"
	"github.com/spf13/cobra"
)

// root command
var rootCmd = &cobra.Command{
	Use: "cofunc",
	Long: `
An automation engine based on function fabric, can used to parse, create, run
and manage flow

Execute 'cofunc' command directly and no any args or sub-command, will list
all flows in interactive mode

Environment variables:
	CO_LOG_DIR=<path of a directory>           // Set the log directory
	CO_FLOW_SOURCE_DIR=<path of a directory>   // Set the flowl source directory

Examples:
	cofunc
	cofunc list
	cofunc parse ./helloworld.flowl
	cofunc run ./helloworld.flowl
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive := true
		return listFlows(interactive)
	},
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
		var stdout bool

		runCmd := &cobra.Command{
			Use:          "run [path of flowl file] or [flow name or id]",
			Short:        "Run a flowl file",
			Example:      "cofunc run ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runflowl(nameid.NameOrID(args[0]), stdout, false)
			},
		}
		runCmd.Flags().BoolVarP(&stdout, "stdout", "s", false, "Directly print the output of the flow to stdout")
		rootCmd.AddCommand(runCmd)
	}

	{
		logCmd := &cobra.Command{
			Use:          "log [flow name or id] [function seq]",
			Short:        "View the execution log of the flow or function",
			Example:      "cofunc run b0804ec967f48520697662a204f5fe72 1",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				nameorid := nameid.NameOrID(args[0])
				seq, err := strconv.ParseInt(args[1], 10, 64)
				if err != nil {
					return err
				}
				return viewLog(nameorid, int(seq))
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
				interactive := false
				return listFlows(interactive)
			},
		}
		rootCmd.AddCommand(listCmd)
	}
}

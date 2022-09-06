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
	COFUNC_HOME=<path of a directory>           // Default $HOME/.cofunc

Examples:
	cofunc
	cofunc list
	cofunc parse ./helloworld.flowl
	cofunc run ./helloworld.flowl
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mainList()
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
		runCmd := &cobra.Command{
			Use:          "run [path of flowl file] or [flow name or id]",
			Short:        "Run a flowl file",
			Example:      "cofunc run ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runflowl(nameid.NameOrID(args[0]))
			},
		}
		rootCmd.AddCommand(runCmd)
	}

	{
		prunCmd := &cobra.Command{
			Use:          "prun [path of flowl file] or [flow name or id]",
			Short:        "Prettily run a flowl file",
			Example:      "cofunc prun ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fullscreen := false
				return prunflowl(nameid.NameOrID(args[0]), fullscreen)
			},
		}
		rootCmd.AddCommand(prunCmd)
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
				return listFlows()
			},
		}
		rootCmd.AddCommand(listCmd)
	}

	{
		stdCmd := &cobra.Command{
			Use:          "std",
			Short:        "List all functions in the standard library",
			Example:      "cofunc std",
			SilenceUsage: true,
			Args:         cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return listStd()
				} else {
					return inspectStd(args[0])
				}
			},
		}
		rootCmd.AddCommand(stdCmd)
	}
}

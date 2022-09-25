package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/cofxlabs/cofx/pkg/nameid"
	"github.com/spf13/cobra"
)

// root command
var rootCmd = &cobra.Command{
	Use: "cofx",
	Long: `A powerful automation workflow engine based on low code programming language

Environment variables:
  COFX_HOME=<path of a directory>           // Default $HOME/.cofx

Examples:
  cofx
  cofx list
  cofx run  helloworld.flowl
  cofx run  helloworld
  cofx run  fc5e038d38a57032085441e7fe7010b0
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return indexEntry()
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
		var envs []string
		runCmd := &cobra.Command{
			Use:          "run [path to flowl file] or [flow name or id]",
			Short:        "Run a flowl",
			Example:      "cofx run ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				for _, env := range envs {
					kv := strings.Split(env, "=")
					if len(kv) == 2 {
						os.Setenv(kv[0], kv[1])
					}
				}
				return runEntry(nameid.NameOrID(args[0]))
			},
		}
		rootCmd.AddCommand(runCmd)
		runCmd.Flags().StringSliceVarP(&envs, "env", "e", nil, "Set environment variables, e.g. -e FOO=bar -e BAZ=qux")
	}

	{
		var envs []string
		prunCmd := &cobra.Command{
			Use:          "prun [path to flowl file] or [flow name or id]",
			Short:        "Prettily run a flowl",
			Example:      "cofx prun ./example.flowl",
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				for _, env := range envs {
					kv := strings.Split(env, "=")
					if len(kv) == 2 {
						os.Setenv(kv[0], kv[1])
					}
				}
				fullscreen := false
				return prunEntry(nameid.NameOrID(args[0]), fullscreen)
			},
		}
		rootCmd.AddCommand(prunCmd)
		prunCmd.Flags().StringSliceVarP(&envs, "env", "e", nil, "Set environment variables, e.g. -e FOO=bar -e BAZ=qux")
	}

	{
		logCmd := &cobra.Command{
			Use:          "log [flow name or id] [function seq]",
			Short:        "View the execution log of the function",
			Example:      "cofx run b0804ec967f48520697662a204f5fe72 1000",
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
			Example:      "cofx list",
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
			Use:          "std [function name]",
			Short:        "List all functions in the standard library or show the manifest of a function",
			Example:      "cofx std",
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

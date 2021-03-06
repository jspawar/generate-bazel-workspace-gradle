package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

var logger = zap.S()

const rootLongHelp = `This utility is intended to assist migration of Maven/Gradle projects to Bazel.

All of the subcommands output Bazel workspace files.
`

var rootCmd = &cobra.Command{
	Use:  "generate-bazel-workspace-gradle",
	Long: rootLongHelp,
}

func init() {
	rootCmd.AddCommand(artifactCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
)

const rootLongHelp =
`This utility is intended to assist migration of Maven/Gradle projects to Bazel.

All of the subcommands output Bazel workspace files.
`

var rootCmd = &cobra.Command{
	Use: "generate-bazel-workspace",
	Long: rootLongHelp,
}

func init() {
	rootCmd.AddCommand(artifactCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

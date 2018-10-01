package cmd

import (
	_ "github.com/jspawar/generate-bazel-workspace-gradle/logging"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"github.com/jspawar/generate-bazel-workspace-gradle/writer"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

const artifactLongHelp = ``

var (
	searchRepositories string
)

var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: `Generates Bazel workspace files from a single Maven artifact and its transitive dependencies`,
	Long:  artifactLongHelp,
	Run:   artifactRunner,
}

func init() {
	artifactCmd.Flags().StringVarP(&searchRepositories, "repos", "r",
		"https://repo.maven.apache.org/maven2",
		"Maven repositories to search through. First match is used.")
}

func artifactRunner(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		logger.Errorf("Invalid arg(s), see correct usage below:\n%s", cmd.UsageString())
		os.Exit(1)
	}

	artifactPom := maven.NewArtifact(args[0])
	searchRepositories = strings.Replace(searchRepositories, ", ", ",", -1)
	depWalker := &maven.DependencyWalker{
		Repositories:     strings.Split(searchRepositories, ","),
		RemoteRepository: maven.NewRemoteRepository(),
	}

	traversedPom, err := depWalker.TraversePOM(artifactPom)
	if err != nil {
		logger.Errorf("Failed to traverse artifact [%s] : %s", artifactPom.GetMavenCoords(), err)
		panic(err)
	}

	// TODO: write Bazel workspace files
	currentPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	outDir := filepath.Dir(currentPath)
	logger.Debugf("Writing Bazel workspace file to directory : %s", outDir)

	// write dependencies
	out, err := os.Create(outDir + "/generate_workspace.bzl")
	if err != nil {
		panic(err)
	}
	wr := writer.NewWorkspaceWriter(out)
	if err := wr.Write(traversedPom); err != nil {
		panic(err)
	}
	logger.Debug("Finished writing Bazel workspace files!")
}

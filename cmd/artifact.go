package cmd

import (
	"github.com/spf13/cobra"
	"strings"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	_ "github.com/jspawar/generate-bazel-workspace-gradle/logging"
	"os"
)

const artifactLongHelp =
``

var (
	searchRepositories string
)

var artifactCmd = &cobra.Command{
	Use: "artifact",
	Short: `Generates Bazel workspace files from a single Maven artifact and its transitive dependencies`,
	Long: artifactLongHelp,
	Run: artifactRunner,
}

func init() {
	artifactCmd.Flags().StringVarP(&searchRepositories, "repos", "r",
		"https://repo.maven.apache.org/maven2/",
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
		Repositories: strings.Split(searchRepositories, ","),
		RemoteRepository: maven.NewRemoteRepository(),
	}

	deps, err := depWalker.TraversePOM(artifactPom)
	if err != nil {
		logger.Errorf("Failed to traverse POM [%s] : %s", artifactPom.AsString(), err)
		panic(err)
	}
	for _, dep := range deps {
		logger.Infof("Found dep [%s] in repository [%s]", dep.AsString(), dep.Repository)
	}

	// TODO: write Bazel workspace files
}

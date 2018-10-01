package writer

import (
	"fmt"
	_ "github.com/jspawar/generate-bazel-workspace-gradle/logging"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"go.uber.org/zap"
	"io"
)

const indent = `  `
const mavenJarsBlockHeader = `def generated_maven_jars():`
const javaLibsBlockHeader = `def generated_java_libraries():`
const excludesDefinition = `excludes = native.existing_rules().keys()`
const artifactDefinitionHeader = `if "%s" not in excludes:`
const mavenJarRule = `native.maven_jar`
const javaLibRule = `native.java_library`

var logger = zap.S()

type WorkspaceWriter struct {
	out io.Writer
}

func NewWorkspaceWriter(w io.Writer) *WorkspaceWriter {
	return &WorkspaceWriter{out: w}
}

func (w *WorkspaceWriter) Write(artifact *maven.Artifact) error {
	w.out.Write([]byte(mavenJarsBlockHeader))
	w.out.Write([]byte("\n"))

	w.writeWithIndents(1, []byte(excludesDefinition))
	w.writeWithIndents(0, []byte("\n\n"))

	// write `maven_jar` rules
	if err := w.writeMavenJarRule(artifact); err != nil {
		return err
	}

	w.out.Write([]byte(javaLibsBlockHeader))
	w.out.Write([]byte("\n"))

	w.writeWithIndents(1, []byte(excludesDefinition))
	w.writeWithIndents(0, []byte("\n\n"))

	// write `java_library` rules
	if err := w.writeJavaLibraryRule(artifact); err != nil {
		return err
	}

	return nil
}

func (w *WorkspaceWriter) writeWithIndents(n int, bs []byte) {
	for i := 0; i < n; i++ {
		w.out.Write([]byte(indent))
	}
	w.out.Write(bs)
}

func (w *WorkspaceWriter) writeMavenJarRule(artifact *maven.Artifact) error {
	logger.Debugf("Writing Maven JAR rule for artifact: [%s]", artifact.GetMavenCoords())

	w.writeWithIndents(1, []byte(fmt.Sprintf(artifactDefinitionHeader, artifact.GetBazelRule())))
	w.writeWithIndents(0, []byte("\n"))

	w.writeWithIndents(2, []byte(mavenJarRule+`(`))

	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(fmt.Sprintf(`name = "%s",`, artifact.GetBazelRule())))
	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(fmt.Sprintf(`artifact = "%s",`, artifact.GetMavenCoords())))
	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(fmt.Sprintf(`repository = "%s",`, artifact.Repository)))
	w.writeWithIndents(0, []byte("\n"))

	w.writeWithIndents(2, []byte(`)`))

	w.writeWithIndents(0, []byte("\n\n"))

	// now add all dependencies as their own `maven_jar` rules
	if artifact.Dependencies != nil && len(artifact.Dependencies) > 0 {
		for _, dep := range artifact.Dependencies {
			w.writeMavenJarRule(dep)
		}
	}

	return nil
}

func (w *WorkspaceWriter) writeJavaLibraryRule(artifact *maven.Artifact) error {
	logger.Debugf("Writing Java library rule for artifact: [%s]", artifact.GetMavenCoords())

	w.writeWithIndents(1, []byte(fmt.Sprintf(artifactDefinitionHeader, artifact.GetBazelRule())))
	w.writeWithIndents(0, []byte("\n"))

	w.writeWithIndents(2, []byte(javaLibRule+`(`))

	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(fmt.Sprintf(`name = "%s",`, artifact.GetBazelRule())))
	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(`visibility = ["//visibility:public"],`))
	w.writeWithIndents(0, []byte("\n"))
	w.writeWithIndents(4, []byte(fmt.Sprintf(`exports = ["@%s//jar"],`, artifact.GetBazelRule())))
	w.writeWithIndents(0, []byte("\n"))

	// write `deps` property for input
	if artifact.Dependencies != nil && len(artifact.Dependencies) > 0 {
		w.writeWithIndents(4, []byte(`deps = [`))
		w.writeWithIndents(0, []byte("\n"))
		for _, dep := range artifact.Dependencies {
			w.writeWithIndents(6, []byte(fmt.Sprintf(`":%s",`, dep.GetBazelRule())))
			w.writeWithIndents(0, []byte("\n"))
		}
		w.writeWithIndents(4, []byte(`],`))
		w.writeWithIndents(0, []byte("\n"))
	}

	w.writeWithIndents(2, []byte(`)`))

	w.writeWithIndents(0, []byte("\n\n"))

	// now add all dependencies as their own `java_library` rules
	if artifact.Dependencies != nil && len(artifact.Dependencies) > 0 {
		for _, dep := range artifact.Dependencies {
			w.writeJavaLibraryRule(dep)
		}
	}

	return nil
}

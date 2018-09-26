package writer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/jspawar/generate-bazel-workspace-gradle/writer"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"github.com/onsi/gomega/gbytes"
	"os"
	"io/ioutil"
)

var _ = Describe("WorkspaceWriter", func() {
	var (
		err    error
		writer *WorkspaceWriter
		pom    *maven.Artifact
	)

	JustBeforeEach(func() {
		err = writer.Write(pom)
	})

	Context("writing to a bytes buffer", func() {
		var (
			out *gbytes.Buffer
		)

		BeforeEach(func() {
			out = gbytes.NewBuffer()
			writer = NewWorkspaceWriter(out)
		})

		AfterEach(func() {
			out.Close()
		})

		Context("given an artifact with no dependencies", func() {
			BeforeEach(func() {
				pom = &maven.Artifact{
					GroupID:    "org.fake",
					ArtifactID: "some-artifact",
					Version:    "0.0.1",
					Repository: "http://localhost/",
				}
			})

			It("should write Bazel rules for just that artifact", func() {

				Expect(err).ToNot(HaveOccurred())

				Expect(string(out.Contents())).To(Equal(
`def generated_maven_jars():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.maven_jar(
        name = "org_fake_some_artifact",
        artifact = "org.fake:some-artifact:0.0.1",
        repository = "http://localhost/",
    )

def generated_java_libraries():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.java_library(
        name = "org_fake_some_artifact",
        visibility = ["//visibility:public"],
        exports = ["@org_fake_some_artifact//jar"],
    )

`,
				))
			})
		})

		Context("given an artifact with dependencies", func() {
			BeforeEach(func() {
				pom = &maven.Artifact{
					GroupID:    "org.fake",
					ArtifactID: "some-artifact",
					Version:    "0.0.1",
					Repository: "http://localhost/",
					Dependencies: []*maven.Artifact{{
						GroupID:    "fake.org",
						ArtifactID: "another-artifact",
						Version:    "2.0.3",
						Repository: "http://localhost/",
					}},
				}
			})

			It("should write Bazel rules for that artifact AND all of its dependencies", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(string(out.Contents())).To(Equal(
`def generated_maven_jars():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.maven_jar(
        name = "org_fake_some_artifact",
        artifact = "org.fake:some-artifact:0.0.1",
        repository = "http://localhost/",
    )

  if "fake_org_another_artifact" not in excludes:
    native.maven_jar(
        name = "fake_org_another_artifact",
        artifact = "fake.org:another-artifact:2.0.3",
        repository = "http://localhost/",
    )

def generated_java_libraries():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.java_library(
        name = "org_fake_some_artifact",
        visibility = ["//visibility:public"],
        exports = ["@org_fake_some_artifact//jar"],
        deps = [
            ":fake_org_another_artifact",
        ],
    )

  if "fake_org_another_artifact" not in excludes:
    native.java_library(
        name = "fake_org_another_artifact",
        visibility = ["//visibility:public"],
        exports = ["@fake_org_another_artifact//jar"],
    )

`,
				))
			})
		})
	})

	Context("writing to a file", func() {
		var (
			out         *os.File
			outContents []byte
		)

		BeforeEach(func() {
			out, err = os.Create(os.TempDir() + "workspace_writer_test")
			Expect(err).ToNot(HaveOccurred())

			writer = NewWorkspaceWriter(out)
		})

		JustBeforeEach(func() {
			// have to open separate stream to inspect contents written to file
			outContents, err = ioutil.ReadFile(out.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			out.Close()
			err = os.Remove(os.TempDir() + "workspace_writer_test")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("given an artifact with no dependencies", func() {
			BeforeEach(func() {
				pom = &maven.Artifact{
					GroupID:    "org.fake",
					ArtifactID: "some-artifact",
					Version:    "0.0.1",
					Repository: "http://localhost/",
				}
			})

			It("should write Bazel rules for just that artifact", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(string(outContents)).To(Equal(
`def generated_maven_jars():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.maven_jar(
        name = "org_fake_some_artifact",
        artifact = "org.fake:some-artifact:0.0.1",
        repository = "http://localhost/",
    )

def generated_java_libraries():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.java_library(
        name = "org_fake_some_artifact",
        visibility = ["//visibility:public"],
        exports = ["@org_fake_some_artifact//jar"],
    )

`,
				))
			})
		})

		Context("given an artifact with dependencies", func() {
			BeforeEach(func() {
				pom = &maven.Artifact{
					GroupID:    "org.fake",
					ArtifactID: "some-artifact",
					Version:    "0.0.1",
					Repository: "http://localhost/",
					Dependencies: []*maven.Artifact{{
						GroupID:    "fake.org",
						ArtifactID: "another-artifact",
						Version:    "2.0.3",
						Repository: "http://localhost/",
					}},
				}
			})

			It("should write Bazel rules for that artifact AND all of its dependencies", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(string(outContents)).To(Equal(
`def generated_maven_jars():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.maven_jar(
        name = "org_fake_some_artifact",
        artifact = "org.fake:some-artifact:0.0.1",
        repository = "http://localhost/",
    )

  if "fake_org_another_artifact" not in excludes:
    native.maven_jar(
        name = "fake_org_another_artifact",
        artifact = "fake.org:another-artifact:2.0.3",
        repository = "http://localhost/",
    )

def generated_java_libraries():
  excludes = native.existing_rules().keys()

  if "org_fake_some_artifact" not in excludes:
    native.java_library(
        name = "org_fake_some_artifact",
        visibility = ["//visibility:public"],
        exports = ["@org_fake_some_artifact//jar"],
        deps = [
            ":fake_org_another_artifact",
        ],
    )

  if "fake_org_another_artifact" not in excludes:
    native.java_library(
        name = "fake_org_another_artifact",
        visibility = ["//visibility:public"],
        exports = ["@fake_org_another_artifact//jar"],
    )

`,
				))
			})
		})
	})
})

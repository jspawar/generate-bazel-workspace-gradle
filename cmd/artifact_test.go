package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"
	"github.com/onsi/gomega/gexec"
	"path/filepath"
	"os"
	"io/ioutil"
)

var _ = Describe("Artifact", func() {
	var (
		args []string
		command *exec.Cmd
		sess *gexec.Session
		out []byte
	)

	JustBeforeEach(func() {
		command = exec.Command(bin, args...)
		sess, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("run with no args", func() {
		BeforeEach(func() {
			args = []string{"artifact"}
		})

		It("returns the usage text", func() {
			Expect(sess.Wait().Err.Contents()).To(ContainSubstring(`Invalid arg(s), see correct usage below:`))

			Eventually(sess, "5s").Should(gexec.Exit(1))
		})
	})

	Context("run with no flags", func() {
		BeforeEach(func() {
			args = []string{"artifact", "junit:junit:4.9"}
		})

		Context("with a valid artifact definition as input", func() {
			It("should create Bazel workspace files", func() {
				Eventually(sess, "30s").Should(gexec.Exit(0))

				// generated file should be placed in same directory as binary
				Expect(func() error {
					_, err = os.Stat(filepath.Dir(bin) + "/generate_workspace.bzl")
					return err
				}()).To(Succeed())
				Expect(func() []byte {
					out, err = ioutil.ReadFile(filepath.Dir(bin) + "/generate_workspace.bzl")
					return out
				}()).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())

				// inspect contents of created file
				Expect(out).ToNot(BeEmpty())
				Expect(string(out)).To(Equal(
`def generated_maven_jars():
  excludes = native.existing_rules().keys()

  if "junit_junit" not in excludes:
    native.maven_jar(
        name = "junit_junit",
        artifact = "junit:junit:4.9",
        repository = "https://repo.maven.apache.org/maven2",
    )

  if "org_hamcrest_hamcrest_core" not in excludes:
    native.maven_jar(
        name = "org_hamcrest_hamcrest_core",
        artifact = "org.hamcrest:hamcrest-core:1.1",
        repository = "https://repo.maven.apache.org/maven2",
    )

def generated_java_libraries():
  excludes = native.existing_rules().keys()

  if "junit_junit" not in excludes:
    native.java_library(
        name = "junit_junit",
        visibility = ["//visibility:public"],
        exports = ["@junit_junit//jar"],
        deps = [
            ":org_hamcrest_hamcrest_core",
        ],
    )

  if "org_hamcrest_hamcrest_core" not in excludes:
    native.java_library(
        name = "org_hamcrest_hamcrest_core",
        visibility = ["//visibility:public"],
        exports = ["@org_hamcrest_hamcrest_core//jar"],
    )

`,
				))
			})
		})
	})
})

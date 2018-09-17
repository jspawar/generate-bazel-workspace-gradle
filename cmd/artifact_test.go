package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Artifact", func() {
	var (
		args []string
		command *exec.Cmd
		sess *gexec.Session
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

			Eventually(sess).Should(gexec.Exit(1))
		})
	})

	Context("run with no flags", func() {
		BeforeEach(func() {
			args = []string{"artifact", "junit:junit:4.9"}
		})

		Context("with a valid artifact definition as input", func() {
			It("should create Bazel workspace files", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})
	})
})

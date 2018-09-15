package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("DependencyWalker", func() {
	var (
		err          error
		repositories []string
		walker       *DependencyWalker
		pom          *Artifact
		deps         []Artifact
	)

	BeforeEach(func() {
		pom = &Artifact{
			GroupID:    "junit",
			ArtifactID: "junit",
			Version:    "4.9",
			Dependencies: []*Artifact{
				{
					GroupID:    "org.hamcrest",
					ArtifactID: "hamcrest-core",
					Version:    "1.1",
				},
			},
		}
	})

	JustBeforeEach(func() {
		walker = &DependencyWalker{Repositories: repositories}
		deps, err = walker.TraversePOM(pom)
	})

	Context("Given a single repository search", func() {
		BeforeEach(func() {
			repositories = []string{"https://repo.maven.apache.org/maven2/"}
		})

		Context("where all dependencies are available in the one repository", func() {
			It("should return all transitive dependencies without error", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(deps).ToNot(BeEmpty())
				Expect(deps).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.hamcrest"),
					"ArtifactID": Equal("hamcrest-core"),
					"Version":    Equal("1.1"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				}), MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("junit"),
					"ArtifactID": Equal("junit"),
					"Version":    Equal("4.9"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
			})
		})

		Context("where all dependencies are NOT available in the one repository", func() {
			var (
				badArtifact = &Artifact{
					GroupID:    "some.fake.org",
					ArtifactID: "some-fake-artifact",
					Version:    "0.0.0",
				}
			)

			BeforeEach(func() {
				pom.Dependencies = []*Artifact{badArtifact}
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("Failed to fetch dependency for POM [junit:junit:4.9] from configured search repositories"))
			})
		})

		Context("where a dependency is encountered twice", func() {
			BeforeEach(func() {
				pom.Dependencies = []*Artifact{
					{GroupID: pom.GroupID, ArtifactID: pom.ArtifactID, Version: pom.Version},
				}
			})

			It("should not duplicate in result", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(deps).ToNot(BeEmpty())
				Expect(deps).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("junit"),
					"ArtifactID": Equal("junit"),
					"Version":    Equal("4.9"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
			})
		})
	})

	Context("Given a multiple repository search", func() {
		Context("where all dependencies are available in any of the provided repositories", func() {})

		Context("where all dependencies are available in only one of the provided repositories", func() {})
	})
})

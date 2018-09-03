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
		pom          *MavenPom
		deps         []MavenArtifact
	)

	BeforeEach(func() {
		pom = &MavenPom{
			GroupID:    "org.apache.commons",
			ArtifactID: "commons-text",
			Version:    "1.4",
			Dependencies: []MavenArtifact{
				MavenArtifact{
					GroupID:    "org.apache.commons",
					ArtifactID: "commons-lang3",
					Version:    "3.7",
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
					"GroupID":    Equal("org.apache.commons"),
					"ArtifactID": Equal("commons-lang3"),
					"Version":    Equal("3.7"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				}), MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.apache.commons"),
					"ArtifactID": Equal("commons-text"),
					"Version":    Equal("1.4"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
			})
		})

		Context("where all dependencies are NOT available in the one repository", func() {
			var (
				badArtifact = MavenArtifact{
					GroupID:    "some.fake.org",
					ArtifactID: "some-fake-artifact",
					Version:    "0.0.0",
				}
			)

			BeforeEach(func() {
				pom.Dependencies = append(pom.Dependencies, badArtifact)
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("The following dependency/dependencies were not found in any of the search repositories : " + badArtifact.AsString()))
			})
		})
	})

	Context("Given a multiple repository search", func() {
		Context("where all dependencies are available in any of the provided repositories", func() {})

		Context("where all dependencies are available in only one of the provided repositories", func() {})
	})
})

package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("DependencyWalker", func() {
	var (
		err                 error
		repositories        []string
		walker              *DependencyWalker
		pom                 *Artifact
		deps                []Artifact
		mockMavenRepo       *MockMavenRepository
		mockServedArtifacts []Artifact
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
					Scope:      "compile",
				},
			},
		}
	})

	JustBeforeEach(func() {
		mockMavenRepo = &MockMavenRepository{Artifacts: mockServedArtifacts}
		mockMavenRepo.Start()

		walker = &DependencyWalker{Repositories: repositories}
		deps, err = walker.TraversePOM(pom)
	})

	AfterEach(func() {
		mockMavenRepo.Stop()
		mockServedArtifacts = nil
	})

	Context("Given a single repository search", func() {
		BeforeEach(func() {
			mockServedArtifacts = []Artifact{*pom}

			repositories = []string{"http://localhost:8080/"}
		})

		Context("where all dependencies are available in the one repository", func() {
			It("should return all transitive dependencies without error", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(deps).ToNot(BeEmpty())
				Expect(deps).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.hamcrest"),
					"ArtifactID": Equal("hamcrest-core"),
					"Version":    Equal("1.1"),
					"Scope":      Equal("compile"),
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
			BeforeEach(func() {
				// mock Maven repository serves no artifacts
				mockServedArtifacts = make([]Artifact, 0)

				pom = &Artifact{
					GroupID:    "some.fake.org",
					ArtifactID: "some-fake-artifact",
					Version:    "0.0.0",
				}
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix(
					"Failed to fetch POM [some.fake.org:some-fake-artifact:0.0.0] from configured search repositories"))
			})
		})

		Context("where a dependency is encountered twice", func() {
			BeforeEach(func() {
				mockServedArtifacts = []Artifact{*pom}

				pom.Dependencies = []*Artifact{
					{GroupID: pom.GroupID, ArtifactID: pom.ArtifactID, Version: pom.Version},
				}
			})

			It("should not duplicate in result", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(deps).ToNot(BeEmpty())
				Expect(deps).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.hamcrest"),
					"ArtifactID": Equal("hamcrest-core"),
					"Version":    Equal("1.1"),
					"Scope":      Equal("compile"),
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
	})

	Context("Given a multiple repository search", func() {
		Context("where all dependencies are available in any of the provided repositories", func() {})

		Context("where all dependencies are available in only one of the provided repositories", func() {})
	})
})

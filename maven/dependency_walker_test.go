package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven/mavenfakes"
	"github.com/pkg/errors"
)

var _ = Describe("DependencyWalker", func() {
	var (
		err              error
		repositories     []string
		remoteRepository *mavenfakes.FakeRemoteRepository
		walker           *DependencyWalker
		pom              *Artifact
		returnedPom      *Artifact
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
		remoteRepository = new(mavenfakes.FakeRemoteRepository)
	})

	JustBeforeEach(func() {
		walker = &DependencyWalker{Repositories: repositories, RemoteRepository: remoteRepository}
		returnedPom, err = walker.TraversePOM(pom)
	})

	Context("Given a single repository search", func() {
		BeforeEach(func() {
			remoteRepository.FetchRemoteModelReturnsOnCall(0, pom, nil)
			remoteRepository.FetchRemoteModelReturnsOnCall(1, pom.Dependencies[0], nil)

			repositories = []string{"http://localhost:8080/"}
		})

		Context("where all dependencies are available in the one repository", func() {
			It("should return all transitive dependencies without error", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(returnedPom).ToNot(BeNil())
				Expect(returnedPom).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("junit"),
					"ArtifactID": Equal("junit"),
					"Version":    Equal("4.9"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
				Expect(returnedPom.Dependencies).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.hamcrest"),
					"ArtifactID": Equal("hamcrest-core"),
					"Version":    Equal("1.1"),
					"Scope":      Equal("compile"),
					"Repository": Equal(repositories[0]),
				}))))
			})
		})

		Context("where all dependencies are NOT available in the one repository", func() {
			BeforeEach(func() {
				remoteRepository.FetchRemoteModelReturnsOnCall(0, nil, errors.New("oh no"))

				pom = &Artifact{
					GroupID:    "some.fake.org",
					ArtifactID: "some-fake-artifact",
					Version:    "0.0.0",
				}
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Failed to traverse POM [some.fake.org:some-fake-artifact:0.0.0] with configured search repositories: oh no"))
			})
		})

		Context("where a dependency is encountered twice", func() {
			BeforeEach(func() {
				remoteRepository.FetchRemoteModelReturnsOnCall(0, pom, nil)
				remoteRepository.FetchRemoteModelReturnsOnCall(1, pom, nil)

				pom.Dependencies = []*Artifact{
					{GroupID: pom.GroupID, ArtifactID: pom.ArtifactID, Version: pom.Version},
				}
			})

			It("should not duplicate in result", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(returnedPom).ToNot(BeNil())
				Expect(returnedPom).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("junit"),
					"ArtifactID": Equal("junit"),
					"Version":    Equal("4.9"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
				Expect(returnedPom.Dependencies).To(BeEmpty())
			})
		})

		Context("where one of the dependencies is marked as optional", func() {
			BeforeEach(func() {
				pom.Dependencies[0].Optional = true
			})

			It("should not include the optional dependencies in result", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(returnedPom).ToNot(BeNil())
				Expect(returnedPom).To(PointTo(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("junit"),
					"ArtifactID": Equal("junit"),
					"Version":    Equal("4.9"),
					"Scope":      BeEmpty(),
					"Repository": Equal(repositories[0]),
				})))
				Expect(returnedPom.Dependencies).To(BeEmpty())
			})
		})
	})

	PContext("Given a multiple repository search", func() {
		Context("where all dependencies are available in any of the provided repositories", func() {})

		Context("where all dependencies are available in only one of the provided repositories", func() {})
	})
})

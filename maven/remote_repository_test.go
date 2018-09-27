package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"net/http/httptest"
	"encoding/xml"
)

var _ = Describe("RemoteRepository", func() {
	var (
		err            error
		repo           RemoteRepository
		mockServer     *httptest.Server
		mockResponse   *Artifact
		toLookup       *Artifact
		remoteArtifact *Artifact
	)

	BeforeEach(func() {
		repo = NewRemoteRepository()
	})

	JustBeforeEach(func() {
		mockServer = initMockServer(mockResponse)

		remoteArtifact, err = repo.FetchRemoteArtifact(toLookup, mockServer.URL)
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Context("given an invalid artifact query", func() {
		BeforeEach(func() {
			mockResponse = &Artifact{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
				Version:    "1.0.1",
			}
			toLookup = &Artifact{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
			}
		})

		It("should return a meaningful error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid POM definition"))
		})
	})

	Context("given a valid artifact query", func() {
		BeforeEach(func() {
			mockResponse = &Artifact{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
				Version:    "1.0.1",
			}
			toLookup = &Artifact{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
				Version:    "1.0.1",
			}
		})

		Context("for an artifact that is NOT present in one of the desired repositories", func() {
			BeforeEach(func() {
				toLookup = &Artifact{
					GroupID:    "foo",
					ArtifactID: "bar",
					Version:    "1",
				}
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed to fetch POM [foo:bar:1] from configured search repositories"))
			})
		})

		Context("for an artifact that is present in one of the desired repositories", func() {
			Context("without a parent", func() {
				It("should return expected artifact without error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(remoteArtifact).ToNot(BeNil())

					Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
					Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
					Expect(remoteArtifact.Version).To(Equal(toLookup.Version))
				})
			})

			Context("WITH a parent", func() {
				BeforeEach(func() {
					mockResponse.Parent = &Artifact{
						GroupID: "foo",
						ArtifactID: "bar",
						Version: "2",
						Properties: Properties{
							Values: []Property{{XMLName: xml.Name{Local: "parent.prop"}, Value: "val"}},
						},
					}
					toLookup.Parent = &Artifact{
						GroupID: "foo",
						ArtifactID: "bar",
						Version: "2",
					}
				})

				It("should return expected artifact without error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(remoteArtifact).ToNot(BeNil())

					Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
					Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
					Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

					Expect(remoteArtifact.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent.GroupID).To(Equal("foo"))
					Expect(remoteArtifact.Parent.ArtifactID).To(Equal("bar"))
					Expect(remoteArtifact.Parent.Version).To(Equal("2"))

					Expect(remoteArtifact.Properties).To(MatchFields(IgnoreExtras, Fields{
						"Values": ContainElement(Property{
							XMLName: xml.Name{Local: "parent.prop"},
							Value: "val",
						}),
					}))
				})
			})
		})
	})
})

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
			Expect(err.Error()).To(Equal("failed to find POM [org.fake:some-artifact:] in configured search repositories"))
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
				Expect(err.Error()).To(Equal("failed to find POM [foo:bar:1] in configured search repositories"))
			})
		})

		Context("for an artifact that is present in one of the desired repositories", func() {
			Context("when remote POM is missing necessary tag(s)", func() {
				BeforeEach(func() {
					mockResponse.GroupID = ""
					toLookup.GroupID = mockResponse.GroupID
				})

				It("should return a meaningful error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error parsing POM [:some-artifact:1.0.1] : invalid POM definition"))
				})
			})

			Context("when remote POM depends on POM properties", func() {
				Context("from itself", func() {
					It("should return expected artifact, with POM properties, without error", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(remoteArtifact).ToNot(BeNil())

						Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
						Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
						Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

						Expect(remoteArtifact.Properties).To(MatchFields(IgnoreExtras, Fields{
							"Values": ContainElement(Property{
								XMLName: xml.Name{Local: "project.version"},
								Value:   "1.0.1",
							}),
						}))
					})
				})

				Context("from parent", func() {
					BeforeEach(func() {
						mockResponse.GroupID = ""
						mockResponse.Version = ""
						toLookup.GroupID = mockResponse.GroupID
						toLookup.Version = mockResponse.Version

						mockResponse.Parent = &Artifact{
							GroupID: "foo",
							ArtifactID: "parent",
							Version: "1.0.1",
						}
					})

					It("should return expected artifact, with POM properties, without error", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(remoteArtifact).ToNot(BeNil())

						Expect(remoteArtifact.GroupID).To(Equal("foo"))
						Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
						Expect(remoteArtifact.Version).To(Equal("1.0.1"))

						Expect(remoteArtifact.Properties).To(MatchFields(IgnoreExtras, Fields{
							"Values": ContainElement(Property{
								XMLName: xml.Name{Local: "project.version"},
								Value:   "1.0.1",
							}),
						}))
					})
				})
			})

			Context("when remote POM depends on Maven properties", func() {
				Context("from itself", func() {
					BeforeEach(func() {
						mockResponse.Dependencies = []*Artifact{{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${maven.property}",
						}}
						mockResponse.Properties = Properties{Values: []Property{
							{XMLName: xml.Name{Local: "maven.property"}, Value: "3.0.2"},
						}}
					})

					It("should return expected artifact, with Maven properties, without error", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(remoteArtifact).ToNot(BeNil())

						Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
						Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
						Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

						Expect(remoteArtifact.Properties).To(MatchFields(IgnoreExtras, Fields{
							"Values": ContainElement(Property{
								XMLName: xml.Name{Local: "maven.property"},
								Value:   "3.0.2",
							}),
						}))

						Expect(remoteArtifact.Dependencies).ToNot(BeEmpty())
						Expect(remoteArtifact.Dependencies).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
							"GroupID":    Equal("foo"),
							"ArtifactID": Equal("bar"),
							"Version":    Equal("3.0.2"),
						}))))
					})
				})

				Context("from parent", func() {
					BeforeEach(func() {
						mockResponse.Dependencies = []*Artifact{{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${maven.property}",
						}}

						mockResponse.Parent = &Artifact{
							GroupID: "foo",
							ArtifactID: "parent",
							Version: "1.0.1",
							Properties: Properties{Values: []Property{
								{XMLName: xml.Name{Local: "maven.property"}, Value: "4.0.4"},
							}},
						}
					})

					It("should return expected artifact, with Maven properties, without error", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(remoteArtifact).ToNot(BeNil())

						Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
						Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
						Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

						Expect(remoteArtifact.Properties).To(MatchFields(IgnoreExtras, Fields{
							"Values": ContainElement(Property{
								XMLName: xml.Name{Local: "maven.property"},
								Value:   "4.0.4",
							}),
						}))

						Expect(remoteArtifact.Dependencies).ToNot(BeEmpty())
						Expect(remoteArtifact.Dependencies).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
							"GroupID":    Equal("foo"),
							"ArtifactID": Equal("bar"),
							"Version":    Equal("4.0.4"),
						}))))
					})
				})
			})

			PContext("when remote POM has multiple levels of parents", func() {
				BeforeEach(func() {
					// TODO: need to refactor mock server to accept a list of artifacts to serve
					mockResponse.Parent = &Artifact{
						GroupID: "foo",
						ArtifactID: "bar",
						Version: "5.0",
						Parent: &Artifact{
							GroupID: "baz",
							ArtifactID: "thing",
							Version: "6.1",
						},
					}
				})

				It("should return expected artifact, with all levels of parents contained, without error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(remoteArtifact).ToNot(BeNil())

					Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
					Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
					Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

					Expect(remoteArtifact.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent).ToNot(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID": Equal("foo"),
						"ArtifactID": Equal("bar"),
						"Version": Equal("5.0"),
					})))

					Expect(remoteArtifact.Parent.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent.Parent).ToNot(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID": Equal("baz"),
						"ArtifactID": Equal("thing"),
						"Version": Equal("6.1"),
					})))
				})
			})

			Context("without any special context", func() {
				It("should return expected artifact without error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(remoteArtifact).ToNot(BeNil())

					Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
					Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
					Expect(remoteArtifact.Version).To(Equal(toLookup.Version))
				})
			})
		})
	})
})

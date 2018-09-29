package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"encoding/xml"
	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"net/http/httptest"
)

var _ = Describe("RemoteRepository", func() {
	var (
		err            error
		repo           RemoteRepository
		mockServer     *httptest.Server
		mockResponses  []Artifact
		toLookup       *Artifact
		remoteArtifact *Artifact
	)

	BeforeEach(func() {
		repo = NewRemoteRepository()
	})

	JustBeforeEach(func() {
		mockServer = initMockServer(mockResponses)

		remoteArtifact, err = repo.FetchRemoteArtifact(toLookup, mockServer.URL)
	})

	AfterEach(func() {
		mockServer.Close()
	})

	Context("given an invalid artifact query", func() {
		BeforeEach(func() {
			mockResponses = []Artifact{{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
				Version:    "1.0.1",
			}}
			toLookup = &Artifact{
				GroupID: "org.fake",
				Version: "1.0",
			}
		})

		It("should return a meaningful error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to find POM [org.fake::1.0] in configured search repositories"))
		})
	})

	Context("given a valid artifact query", func() {
		BeforeEach(func() {
			mockResponses = []Artifact{{
				GroupID:    "org.fake",
				ArtifactID: "some-artifact",
				Version:    "1.0.1",
			}}
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
					mockResponses[0].GroupID = ""
					toLookup.GroupID = mockResponses[0].GroupID
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
						mockResponses[0].GroupID = ""
						mockResponses[0].Version = ""
						toLookup.GroupID = mockResponses[0].GroupID
						toLookup.Version = mockResponses[0].Version

						mockResponses[0].Parent = &Artifact{
							GroupID:    "foo",
							ArtifactID: "parent",
							Version:    "1.0.1",
						}
						mockResponses = append(mockResponses, *mockResponses[0].Parent)
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
						mockResponses[0].Dependencies = []*Artifact{{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${maven.property}",
						}}
						mockResponses[0].Properties = Properties{Values: []Property{
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
						mockResponses[0].Dependencies = []*Artifact{{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${maven.property}",
						}}
						mockResponses[0].Parent = &Artifact{
							GroupID:    "foo",
							ArtifactID: "parent",
							Version:    "1.0.1",
							Properties: Properties{Values: []Property{
								{XMLName: xml.Name{Local: "maven.property"}, Value: "4.0.4"},
							}},
						}

						mockResponses = append(mockResponses, *mockResponses[0].Parent)
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

			Context("when remote POM has multiple levels of parents", func() {
				BeforeEach(func() {
					mockResponses[0].Parent = &Artifact{
						GroupID:    "foo",
						ArtifactID: "bar",
						Version:    "5.0",
						Parent: &Artifact{
							GroupID:    "baz",
							ArtifactID: "thing",
							Version:    "6.1",
							Parent: &Artifact{
								GroupID:    "third",
								ArtifactID: "level",
								Version:    "3.3.3",
							},
						},
					}
					mockResponses = append(mockResponses, *mockResponses[0].Parent)
					mockResponses = append(mockResponses, *mockResponses[0].Parent.Parent)
					mockResponses = append(mockResponses, *mockResponses[0].Parent.Parent.Parent)
				})

				It("should return expected artifact, with all levels of parents contained, without error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(remoteArtifact).ToNot(BeNil())

					Expect(remoteArtifact.GroupID).To(Equal(toLookup.GroupID))
					Expect(remoteArtifact.ArtifactID).To(Equal(toLookup.ArtifactID))
					Expect(remoteArtifact.Version).To(Equal(toLookup.Version))

					Expect(remoteArtifact.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("foo"),
						"ArtifactID": Equal("bar"),
						"Version":    Equal("5.0"),
					})))

					Expect(remoteArtifact.Parent.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent.Parent).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("baz"),
						"ArtifactID": Equal("thing"),
						"Version":    Equal("6.1"),
					})))

					Expect(remoteArtifact.Parent.Parent.Parent).ToNot(BeNil())
					Expect(remoteArtifact.Parent.Parent.Parent).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("third"),
						"ArtifactID": Equal("level"),
						"Version":    Equal("3.3.3"),
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

package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"encoding/xml"
	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("Models", func() {
	var (
		err            error
		artifactString string
		pomString      string
		pom            *Artifact
	)

	Context("Given an artifact definition as a string", func() {
		JustBeforeEach(func() {
			pom = NewArtifact(artifactString)
		})

		Context("that is valid", func() {
			BeforeEach(func() {
				artifactString = "some.group:with.some.artifact:1.0.0-SNAPSHOT"
			})

			It("should return an artifact object without error", func() {
				Expect(pom).ToNot(BeNil())

				Expect(pom.GroupID).To(Equal("some.group"))
				Expect(pom.ArtifactID).To(Equal("with.some.artifact"))
				Expect(pom.Version).To(Equal("1.0.0-SNAPSHOT"))
			})
		})

		Context("that is invalid", func() {
			BeforeEach(func() {
				artifactString = "some.group:with.some.artifact"
			})

			It("should return an empty object", func() {
				Expect(pom).ToNot(BeNil())

				Expect(pom.GroupID).To(BeEmpty())
				Expect(pom.ArtifactID).To(BeEmpty())
				Expect(pom.Version).To(BeEmpty())
			})
		})
	})

	Context("Given contents of a pom.xml file", func() {
		JustBeforeEach(func() {
			pom, err = UnmarshalPOM([]byte(pomString))
		})

		Context("with no Maven properties", func() {
			Context("and pom.xml has all necessary properties", func() {
				BeforeEach(func() {
					pomString = `
					<project xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
						<modelVersion>4.0.0</modelVersion>
						<parent>
							<groupId>org.apache.commons</groupId>
							<artifactId>commons-parent</artifactId>
							<version>46</version>
						</parent>
						<groupId>org.apache.commons</groupId>
						<artifactId>commons-text</artifactId>
						<version>1.4</version>
						<name>Apache Commons Text</name>
						<dependencies>
							<dependency>
								<groupId>org.apache.commons</groupId>
								<artifactId>commons-lang3</artifactId>
								<version>3.7</version>
							</dependency>
							<!-- testing -->
							<dependency>
								<groupId>org.junit.jupiter</groupId>
								<artifactId>junit-jupiter-engine</artifactId>
								<version>5.2.0</version>
								<scope>test</scope>
								<optional>true</optional>
							</dependency>
						</dependencies>
					</project>
					`
				})

				It("should deserialize without error", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(pom.ModelVersion).To(Equal("4.0.0"))
					Expect(pom.Parent).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.apache.commons"),
						"ArtifactID": Equal("commons-parent"),
						"Version":    Equal("46"),
						"Scope":      BeEmpty(),
					})))
					Expect(pom.GroupID).To(Equal("org.apache.commons"))
					Expect(pom.ArtifactID).To(Equal("commons-text"))
					Expect(pom.Version).To(Equal("1.4"))

					Expect(pom.Dependencies).ToNot(BeNil())
					Expect(pom.Dependencies).To(HaveLen(2))
					Expect(pom.Dependencies).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.apache.commons"),
						"ArtifactID": Equal("commons-lang3"),
						"Version":    Equal("3.7"),
						"Scope":      BeEmpty(),
						"Optional":   Equal(false),
					})), PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.junit.jupiter"),
						"ArtifactID": Equal("junit-jupiter-engine"),
						"Version":    Equal("5.2.0"),
						"Scope":      Equal("test"),
						"Optional":   Equal(true),
					}))))
				})
			})
		})
	})

	Context("Given a POM struct", func() {
		Context("to construct a POM path for", func() {
			var (
				searchPath string
			)

			JustBeforeEach(func() {
				searchPath = pom.PathToPOM()
			})

			Context("with valid artifact definition", func() {
				BeforeEach(func() {
					pom = &Artifact{
						GroupID:    "some.fake.org",
						ArtifactID: "some-fake-artifact",
						Version:    "0.0.0",
					}
				})

				It("should return a valid search path", func() {
					Expect(searchPath).To(Equal("some/fake/org/some-fake-artifact/0.0.0/some-fake-artifact-0.0.0.pom"))
				})
			})
		})

		Context("to construct a Bazel rule name for", func() {
			var (
				bazelRuleName string
			)

			JustBeforeEach(func() {
				bazelRuleName = pom.GetBazelRule()
			})

			Context("with valid artifact definition", func() {
				BeforeEach(func() {
					pom = &Artifact{
						GroupID:    "some.fake.org",
						ArtifactID: "some-fake-artifact",
						Version:    "0.0.0",
					}
				})

				It("should return a valid Bazel rule name", func() {
					Expect(bazelRuleName).To(Equal("some_fake_org_some_fake_artifact"))
				})
			})
		})

		Context("to interpolate properties for", func() {
			Context("properties coming from parent", func() {
				JustBeforeEach(func() {
					pom.InterpolateFromParent()
				})

				Context("with a parent and no group ID", func() {
					BeforeEach(func() {
						pom = &Artifact{
							ArtifactID: "bar",
							Version:    "12",
							Parent: &Artifact{
								GroupID: "foo",
							},
						}
					})

					It("should interpolate correctly", func() {
						Expect(pom.GroupID).To(Equal("foo"))
					})
				})

				Context("with a parent and no version", func() {
					BeforeEach(func() {
						pom = &Artifact{
							GroupID:    "foo",
							ArtifactID: "bar",
							Parent: &Artifact{
								Version: "4.3",
							},
						}
					})

					It("should interpolate correctly", func() {
						Expect(pom.Version).To(Equal("4.3"))
					})
				})
			})

			Context("properties coming from itself", func() {
				var (
					interpolated string
				)

				JustBeforeEach(func() {
					interpolated, err = pom.InterpolateFromProperties(pom.Version)
				})

				Context("and the expected property is present", func() {
					BeforeEach(func() {
						pom = &Artifact{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${expected.property}",
							Properties: Properties{Values: []Property{
								{XMLName: xml.Name{Local: "expected.property"}, Value: "6.7"},
							}},
						}
					})

					It("should interpolate correctly", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(interpolated).To(Equal("6.7"))
					})
				})

				Context("and the expected property depends on another property", func() {
					BeforeEach(func() {
						pom = &Artifact{
							GroupID:    "foo",
							ArtifactID: "bar",
							Version:    "${expected.property}",
							Properties: Properties{Values: []Property{
								{XMLName: xml.Name{Local: "expected.property"}, Value: "${another.property}"},
								{XMLName: xml.Name{Local: "another.property"}, Value: "6.7"},
							}},
						}

						pom.InterpolatePropertiesFromProperties()
					})

					It("should interpolate correctly", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(interpolated).To(Equal("6.7"))
					})
				})
			})
		})

		Context("to construct a metadata path for", func() {
			var metadataPath string

			BeforeEach(func() {
				pom = &Artifact{
					GroupID:    "foo.bar",
					ArtifactID: "thing",
				}
			})

			JustBeforeEach(func() {
				metadataPath = pom.MetadataPath()
			})

			It("should return an expected metadata path", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(metadataPath).To(Equal("foo/bar/thing/maven-metadata.xml"))
			})
		})

		Context("to construct a JAR SHA1 checksum path for", func() {
			var pathToSHA1 string

			JustBeforeEach(func() {
				pathToSHA1 = pom.PathToJarSHA1()
			})

			Context("with valid artifact definition", func() {
				BeforeEach(func() {
					pom = &Artifact{
						GroupID:    "some.fake.org",
						ArtifactID: "some-fake-artifact",
						Version:    "0.0.0",
						SHA:        "some-sha1",
					}
				})

				It("should return a valid path to the model's SHA1 for its JAR", func() {
					Expect(pathToSHA1).To(Equal("some/fake/org/some-fake-artifact/0.0.0/some-fake-artifact-0.0.0.jar.sha1"))
				})
			})
		})
	})
})

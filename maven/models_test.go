package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("Models", func() {
	var (
		err       error
		artifactString string
		pomString string
		pom       *Artifact
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
				Expect(err).ToNot(HaveOccurred())
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
				Expect(err).ToNot(HaveOccurred())
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

		// TODO: need to account for POM properties used to define dep versions
		Context("that is valid", func() {
			Context("with no Maven properties", func() {
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
					})), PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.junit.jupiter"),
						"ArtifactID": Equal("junit-jupiter-engine"),
						"Version":    Equal("5.2.0"),
						"Scope":      Equal("test"),
					}))))
				})
			})

			Context("with Maven properties specifying dependency versions", func() {
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
								<version>${apache.lang.version}</version>
							</dependency>
							<!-- testing -->
							<dependency>
								<groupId>org.junit.jupiter</groupId>
								<artifactId>junit-jupiter-engine</artifactId>
								<version>${junit.jupiter.version}</version>
								<scope>test</scope>
							</dependency>
						</dependencies>
						<properties>
							<junit.jupiter.version>5.2.0</junit.jupiter.version>
							<apache.lang.version>3.7</apache.lang.version>
						</properties>
					</project>
					`
				})

				It("should deserialize, with correctly interpolated dependency versions, without error", func() {
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

					Expect(pom.Properties.Values).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
						"XMLName": MatchFields(IgnoreExtras, Fields{"Local": Equal("junit.jupiter.version")}),
						"Value": Equal("5.2.0"),
					}), MatchFields(IgnoreExtras, Fields{
						"XMLName": MatchFields(IgnoreExtras, Fields{"Local": Equal("apache.lang.version")}),
						"Value": Equal("3.7"),
					})))

					Expect(pom.Dependencies).ToNot(BeNil())
					Expect(pom.Dependencies).To(HaveLen(2))
					Expect(pom.Dependencies).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.apache.commons"),
						"ArtifactID": Equal("commons-lang3"),
						"Version":    Equal("3.7"),
						"Scope":      BeEmpty(),
					})), PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.junit.jupiter"),
						"ArtifactID": Equal("junit-jupiter-engine"),
						"Version":    Equal("5.2.0"),
						"Scope":      Equal("test"),
					}))))
				})
			})

			Context("with a GroupID inherited from a parent POM", func() {
				BeforeEach(func() {
					pomString = `
					<project xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
						<modelVersion>4.0.0</modelVersion>
						<parent>
							<groupId>org.apache.commons</groupId>
							<artifactId>commons-parent</artifactId>
							<version>46</version>
						</parent>
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
					})), PointTo(MatchFields(IgnoreExtras, Fields{
						"GroupID":    Equal("org.junit.jupiter"),
						"ArtifactID": Equal("junit-jupiter-engine"),
						"Version":    Equal("5.2.0"),
						"Scope":      Equal("test"),
					}))))
				})
			})
		})
	})

	Context("Given a POM struct", func() {
		Context("to construct a search path for", func() {
			var (
				searchPath string
			)

			JustBeforeEach(func() {
				searchPath, err = pom.SearchPath()
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
					Expect(err).NotTo(HaveOccurred())
					Expect(searchPath).To(Equal("some/fake/org/some-fake-artifact/0.0.0/some-fake-artifact-0.0.0.pom"))
				})
			})

			Context("with invalid artifact definition", func() {
				BeforeEach(func() {
					pom = &Artifact{}
				})

				It("should return a meaningful error", func() {
					Expect(err).To(MatchError("invalid POM definition"))
				})
			})
		})
	})
})

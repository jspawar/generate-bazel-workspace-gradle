package maven_test

import (
	"encoding/xml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("Models", func() {
	var (
		err       error
		pomString string
		pom       *MavenPom
	)

	Context("Given a pom.xml file", func() {
		JustBeforeEach(func() {
			pom = &MavenPom{}
			err = xml.Unmarshal([]byte(pomString), pom)
		})

		Context("that is valid", func() {
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
				Expect(pom.Dependencies).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.apache.commons"),
					"ArtifactID": Equal("commons-lang3"),
					"Version":    Equal("3.7"),
					"Scope":      BeEmpty(),
				}), MatchFields(IgnoreExtras, Fields{
					"GroupID":    Equal("org.junit.jupiter"),
					"ArtifactID": Equal("junit-jupiter-engine"),
					"Version":    Equal("5.2.0"),
					"Scope":      Equal("test"),
				})))
			})
		})
	})

	Context("Given a POM struct", func() {
		Context("to construct a search path for", func() {
			var (
				searchPath string
			)

			JustBeforeEach(func() {
				searchPath, err = pom.AsArtifact().SearchPath()
			})

			Context("with valid artifact definition", func() {
				BeforeEach(func() {
					pom = &MavenPom{
						GroupID:    "some.fake.org",
						ArtifactID: "some-fake-artifact",
						Version:    "0.0.0",
					}
				})

				It("should return a valid search path", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(searchPath).To(Equal("some/fake/org/some-fake-artifact/0.0.0/"))
				})
			})

			Context("with invalid artifact definition", func() {
				BeforeEach(func() {
					pom = &MavenPom{}
				})

				It("should return a meaningful error", func() {
					Expect(err).To(MatchError("invalid POM definition"))
				})
			})
		})
	})
})

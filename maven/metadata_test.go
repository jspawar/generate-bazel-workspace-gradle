package maven_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/jspawar/generate-bazel-workspace-gradle/maven"
)

var _ = Describe("Metadata", func() {
	var (
		err            error
		metadataString string
		metadata       *Metadata
	)

	Context("given contents of a metadata file", func() {
		JustBeforeEach(func() {
			metadata, err = UnmarshalMetadata([]byte(metadataString))
		})

		Context("that is valid", func() {
			BeforeEach(func() {
				metadataString = `
<metadata xsi:schemaLocation="http://maven.apache.org/METADATA/1.1.0 http://maven.apache.org/xsd/metadata-1.1.0.xsd">
	<groupId>foo</groupId>
	<artifactId>bar</artifactId>
	<versioning>
		<latest>1.0.2-SNAPSHOT</latest>
		<release>1.0.1</release>
	</versioning>
</metadata>
`
			})

			It("should return a metadata object without error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(metadata).ToNot(BeNil())

				Expect(metadata.GroupID).To(Equal("foo"))
				Expect(metadata.ArtifactID).To(Equal("bar"))
				Expect(metadata.Latest).To(Equal("1.0.2-SNAPSHOT"))
				Expect(metadata.Release).To(Equal("1.0.1"))
			})
		})

		Context("that is invalid", func() {
			BeforeEach(func() {
				metadataString = ""
			})

			It("should return a meaningful error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("error parsing metadata"))
			})
		})
	})
})

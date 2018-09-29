package maven

import (
	"encoding/xml"
	"bytes"
	"golang.org/x/net/html/charset"
	"github.com/pkg/errors"
)

type Metadata struct {
	xml.Name
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Latest     string `xml:"versioning>latest"`
	Release    string `xml:"versioning>release"`
}

func UnmarshalMetadata(contents []byte) (*Metadata, error) {
	metadata := &Metadata{}
	reader := bytes.NewReader(contents)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(metadata); err != nil {
		return nil, errors.Wrapf(err, "error parsing metadata")
	}
	return metadata, nil
}

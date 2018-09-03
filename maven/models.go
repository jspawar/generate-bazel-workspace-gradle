package maven

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

// TODO: do any assertions about Maven model version?
type MavenPom struct {
	XMLName      xml.Name        `xml:"project"`
	Parent       *MavenArtifact  `xml:"parent"`
	ModelVersion string          `xml:"modelVersion"`
	GroupID      string          `xml:"groupId"`
	ArtifactID   string          `xml:"artifactId"`
	Version      string          `xml:"version"`
	Dependencies []MavenArtifact `xml:"dependencies>dependency"`
}

func (p *MavenPom) SearchPath() (string, error) {
	if p.GroupID == "" || p.ArtifactID == "" || p.Version == "" {
		return "", errors.New("Invalid POM definition")
	}
	return fmt.Sprintf("%s/%s/%s/", strings.Replace(p.GroupID, ".", "/", -1), p.ArtifactID, p.Version), nil
}

type MavenArtifact struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Repository string
	SHA        string
}

func (a *MavenArtifact) AsString() string {
	return fmt.Sprintf("%s:%s:%s", a.GroupID, a.ArtifactID, a.Version)
}

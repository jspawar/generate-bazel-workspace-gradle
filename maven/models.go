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

func (p *MavenPom) AsArtifact() *MavenArtifact {
	return &MavenArtifact{
		GroupID:    p.GroupID,
		ArtifactID: p.ArtifactID,
		Version:    p.Version,
	}
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

func (a *MavenArtifact) IsValid() bool {
	// TODO: use regex to check IDs and version syntax correctly
	return a.GroupID != "" && a.ArtifactID != "" && a.Version != ""
}

func (a *MavenArtifact) SearchPath() (string, error) {
	if a.GroupID == "" || a.ArtifactID == "" || a.Version == "" {
		return "", errors.New("invalid POM definition")
	}
	return fmt.Sprintf("%s/%s/%s/", strings.Replace(a.GroupID, ".", "/", -1), a.ArtifactID, a.Version), nil
}

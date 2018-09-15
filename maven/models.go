package maven

import (
	"errors"
	"fmt"
	"strings"
)

// TODO: do any assertions about Maven model version?
type Artifact struct {
	GroupID      string     `xml:"groupId"`
	ArtifactID   string     `xml:"artifactId"`
	Version      string     `xml:"version"`
	Scope        string     `xml:"scope"`
	Repository   string
	SHA          string
	Parent       *Artifact  `xml:"parent"`
	ModelVersion string     `xml:"modelVersion"`
	Dependencies []Artifact `xml:"dependencies>dependency"`
}

func (a *Artifact) AsString() string {
	return fmt.Sprintf("%s:%s:%s", a.GroupID, a.ArtifactID, a.Version)
}

func (a *Artifact) IsValid() bool {
	// TODO: use regex to check IDs and version syntax correctly
	return a.GroupID != "" && a.ArtifactID != "" && a.Version != ""
}

func (a *Artifact) SearchPath() (string, error) {
	if a.GroupID == "" || a.ArtifactID == "" || a.Version == "" {
		return "", errors.New("invalid POM definition")
	}
	return fmt.Sprintf("%s/%s/%s/", strings.Replace(a.GroupID, ".", "/", -1), a.ArtifactID, a.Version), nil
}

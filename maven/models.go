package maven

import (
	"fmt"
	"strings"
	"encoding/xml"
	"bytes"
	"golang.org/x/net/html/charset"
	"regexp"
	"github.com/pkg/errors"
)

var propertyRegex = regexp.MustCompile(`^\${(.*)}$`)
var artifactRegex = regexp.MustCompile(`^(.+):(.+):(.+)$`)

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
	Properties   Properties `xml:"properties"`
	Dependencies []*Artifact `xml:"dependencies>dependency"`
}

type Properties struct {
	Values []Property `xml:",any"`
}

type Property struct {
	XMLName  xml.Name
	Value string `xml:",innerxml"`
}

func NewArtifact(artifact string) *Artifact {
	a := &Artifact{}

	ms := artifactRegex.FindStringSubmatch(artifact)
	if len(ms) > 3 {
		a.GroupID = ms[1]
		a.ArtifactID = ms[2]
		a.Version = ms[3]
	}

	return a
}

func (a *Artifact) AsString() string {
	return fmt.Sprintf("%s:%s:%s", a.GroupID, a.ArtifactID, a.Version)
}

func (a *Artifact) IsValid() bool {
	// TODO: use regex to check IDs and version syntax correctly
	return a.GroupID != "" && a.ArtifactID != "" && a.Version != ""
}

func (a *Artifact) SearchPath() (string, error) {
	if !a.IsValid() {
		return "", errors.New("invalid POM definition")
	}
	return fmt.Sprintf("%s/%s/%s/%s-%s.pom",
		strings.Replace(a.GroupID, ".", "/", -1), a.ArtifactID, a.Version, a.ArtifactID, a.Version),
		nil
}

func (a *Artifact) findPropertyValue(property string) string {
	for _, prop := range a.Properties.Values {
		if prop.XMLName.Local == property {
			return prop.Value
		}
	}
	return ""
}

func UnmarshalPOM(contents []byte) (*Artifact, error) {
	pom := &Artifact{}
	reader := bytes.NewReader(contents)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(pom); err != nil {
		return nil, errors.Wrapf(err, "error parsing POM")
	}

	// read GroupID from parent block if need be
	if pom.GroupID == "" && pom.Parent != nil {
		pom.GroupID = pom.Parent.GroupID
	}

	// interpolate Maven properties for Versions
	for _, dep := range pom.Dependencies {
		ms := propertyRegex.FindStringSubmatch(dep.Version)
		if len(ms) > 1 {
			prop := ms[1]
			propVal := pom.findPropertyValue(prop)
			if propVal == "" {
				return nil, errors.Errorf("error parsing POM : error interpolating Maven properties")
			}
			dep.Version = propVal
		}
	}

	return pom, nil
}

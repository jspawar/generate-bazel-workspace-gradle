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
	XMLName      xml.Name
	GroupID      string      `xml:"groupId"`
	ArtifactID   string      `xml:"artifactId"`
	Version      string      `xml:"version"`
	Scope        string      `xml:"scope,omitempty"`
	Repository   string      `xml:"-"`
	SHA          string      `xml:"-"`
	Parent       *Artifact   `xml:"parent,omitempty"`
	ModelVersion string      `xml:"modelVersion,omitempty"`
	Properties   Properties  `xml:"properties,omitempty"`
	Dependencies []*Artifact `xml:"dependencies>dependency,omitempty"`
}

type Properties struct {
	Values []Property `xml:",any"`
}

type Property struct {
	XMLName xml.Name
	Value   string `xml:",innerxml"`
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

func (a *Artifact) GetMavenCoords() string {
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

func (a *Artifact) GetBazelRule() string {
	groupID := strings.Replace(a.GroupID, ".", "_", -1)
	artifactID := strings.Replace(a.ArtifactID, "-", "_", -1)
	return fmt.Sprintf("%s_%s", groupID, artifactID)
}

func (a *Artifact) InterpolateFromProperties(interpolate string) (string, error) {
	ms := propertyRegex.FindStringSubmatch(interpolate)
	if len(ms) > 1 {
		prop := ms[1]
		propVal := a.findPropertyValue(prop)
		if propVal == "" {
			return "", errors.Errorf(
				"error parsing POM : value not found to interpolate Maven property [%s]", prop)
		}
		return propVal, nil
	} else {
		return interpolate, nil
	}
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

	// read group ID from parent block if need be
	if pom.GroupID == "" && pom.Parent != nil {
		pom.GroupID = pom.Parent.GroupID
	}

	// read version from parent block if need be
	if pom.Version == "" && pom.Parent != nil {
		pom.Version = pom.Parent.Version
	}

	// read parent Maven properties if need be
	if pom.Parent != nil && pom.Parent.Properties.Values != nil && len(pom.Parent.Properties.Values) > 0 {
		if pom.Properties.Values == nil {
			pom.Properties.Values = pom.Parent.Properties.Values
		} else {
			pom.Properties.Values = append(pom.Properties.Values, pom.Parent.Properties.Values...)
		}
	}

	// interpolate Maven properties for Versions
	for _, dep := range pom.Dependencies {
		interpolatedVersion, err := pom.InterpolateFromProperties(dep.Version)
		if err != nil {
			return nil, err
		}
		dep.Version = interpolatedVersion
	}

	// throw error if deserialized POM is invalid
	if !pom.IsValid() {
		return nil, errors.Errorf("error parsing POM [%s] : input POM was invalid", pom.GetMavenCoords())
	}

	return pom, nil
}

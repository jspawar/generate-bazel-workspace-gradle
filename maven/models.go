package maven

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/net/html/charset"
	"regexp"
	"strings"
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

func (a *Artifact) SearchPath() string {
	// TODO: return with leading forward slash?
	return fmt.Sprintf("%s/%s/%s/%s-%s.pom",
		strings.Replace(a.GroupID, ".", "/", -1), a.ArtifactID, a.Version, a.ArtifactID, a.Version)
}

func (a *Artifact) GetBazelRule() string {
	groupID := strings.Replace(a.GroupID, ".", "_", -1)
	artifactID := strings.Replace(a.ArtifactID, "-", "_", -1)
	return fmt.Sprintf("%s_%s", groupID, artifactID)
}

func (a *Artifact) InterpolateFromParent() {
	// read group ID from parent block if need be
	if a.GroupID == "" && a.Parent != nil {
		a.GroupID = a.Parent.GroupID
	}

	// read version from parent block if need be
	if a.Version == "" && a.Parent != nil {
		a.Version = a.Parent.Version
	}
}

// TODO: conform to spec for interpolation: http://maven.apache.org/ref/current/maven-model-builder/
// TODO: up to what version of spec above to not conform to?
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
	return pom, nil
}

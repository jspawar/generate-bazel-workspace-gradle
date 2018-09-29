package maven

import (
	"net/http"
	"github.com/pkg/errors"
	"fmt"
	"io/ioutil"
	"encoding/xml"
)

//go:generate counterfeiter . RemoteRepository
type RemoteRepository interface {
	FetchRemoteArtifact(artifact *Artifact, remoteRepository string) (*Artifact, error)
}

type remoteRepository struct {}

func NewRemoteRepository() RemoteRepository {
	return &remoteRepository{}
}

// TODO: fix assumption of no trailing "/" on repo URL
func (r *remoteRepository) FetchRemoteArtifact(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	remoteArtifact, err := r.doFetch(artifact, remoteRepository)
	if err != nil {
		return nil, err
	}

	// inheritance assembly
	r.doInherit(remoteArtifact)

	// parent resolution
	if remoteArtifact.Parent != nil {
		remoteParent, err := r.FetchRemoteArtifact(remoteArtifact.Parent, remoteRepository)
		if err != nil {
			return nil, err
		}
		remoteArtifact.Parent = remoteParent
	}

	// model interpolation
	if err := r.doInterpolation(remoteArtifact); err != nil {
		return nil, err
	}

	// TODO: validate here or in separate place? also, delegate validation to validation class?
	if !remoteArtifact.IsValid() {
		return nil, errors.Errorf("error parsing POM [%s] : invalid POM definition", remoteArtifact.GetMavenCoords())
	}

	return remoteArtifact, nil
}

func (r *remoteRepository) doFetch(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s", remoteRepository, artifact.SearchPath()))
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to find POM [%s] in configured search repositories",
			artifact.GetMavenCoords())
	}
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"failed to find POM [%s] in configured search repositories",
			artifact.GetMavenCoords()))
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	remoteArtifact, err := UnmarshalPOM(bs)
	if err != nil {
		panic(err)
	}
	return remoteArtifact, nil
}

func (r *remoteRepository) doInherit(artifact *Artifact) {
	artifact.InterpolateFromParent()
	artifact.Properties.Values = append(artifact.Properties.Values, Property{
		XMLName: xml.Name{Local: "project.version"},
		Value:   artifact.Version,
	})
}

func (r *remoteRepository) doInterpolation(artifact *Artifact) error {
	// read parent Maven properties if need be
	if artifact.Parent != nil && artifact.Parent.Properties.Values != nil && len(artifact.Parent.Properties.Values) > 0 {
		if artifact.Properties.Values == nil {
			artifact.Properties.Values = artifact.Parent.Properties.Values
		} else {
			artifact.Properties.Values = append(artifact.Properties.Values, artifact.Parent.Properties.Values...)
		}
	}
	// interpolate
	for _, dep := range artifact.Dependencies {
		interpolatedVersion, err := artifact.InterpolateFromProperties(dep.Version)
		if err != nil {
			return err
		}
		dep.Version = interpolatedVersion
	}
	return nil
}

package maven

import (
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

//go:generate counterfeiter . RemoteRepository
type RemoteRepository interface {
	FetchRemoteModel(artifact *Artifact, remoteRepository string) (*Artifact, error)
	CheckRemoteJAR(artifact *Artifact, remoteRepository string) (string, error)
}

type remoteRepository struct{}

func NewRemoteRepository() RemoteRepository {
	return &remoteRepository{}
}

// TODO: fix assumption of no trailing "/" on repo URL
func (r *remoteRepository) FetchRemoteModel(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	// get latest version if needed
	if artifact.Version == "" {
		latestVersion, err := r.fetchLatestVersion(artifact, remoteRepository)
		if err != nil {
			return nil, err
		}
		artifact.Version = latestVersion
	}

	remoteArtifact, err := r.doFetch(artifact, remoteRepository)
	if err != nil {
		return nil, err
	}

	// inheritance assembly
	r.doInherit(remoteArtifact)

	// parent resolution
	if remoteArtifact.Parent != nil {
		remoteParent, err := r.FetchRemoteModel(remoteArtifact.Parent, remoteRepository)
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

func (r *remoteRepository) CheckRemoteJAR(artifact *Artifact, remoteRepository string) (string, error) {
	return "", nil
}

func (r *remoteRepository) doFetch(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s", remoteRepository, artifact.PathToPOM()))
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
		return nil, err
	}
	remoteArtifact, err := UnmarshalPOM(bs)
	if err != nil {
		return nil, err
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
	// ensure properties themselves have been interpolated
	artifact.InterpolatePropertiesFromProperties()
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

func (r *remoteRepository) fetchLatestVersion(artifact *Artifact, remoteRepository string) (string, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s", remoteRepository, artifact.MetadataPath()))
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to find metadata for POM [%s] in configured search repositories",
			artifact.GetMavenCoords())
	}
	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf(
			"failed to find metadata for POM [%s] in configured search repositories",
			artifact.GetMavenCoords()))
	}
	defer res.Body.Close()

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	metadata, err := UnmarshalMetadata(bs)
	if err != nil {
		return "", err
	}

	// return most recent "release" version if available, else refer to "latest"
	if metadata.Release == "" {
		// return value in "version" if "release" and "latest" aren't available
		if metadata.Latest == "" {
			return metadata.Version, nil
		}
		return metadata.Latest, nil
	}
	return metadata.Release, nil
}

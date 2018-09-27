package maven

import (
	"net/http"
	"github.com/pkg/errors"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter . RemoteRepository
type RemoteRepository interface {
	FetchRemoteArtifact(artifact *Artifact, remoteRepository string) (*Artifact, error)
}

type remoteRepository struct {}

func NewRemoteRepository() RemoteRepository {
	return &remoteRepository{}
}

// TODO: test drive this
// TODO: fix assumption of no trailing "/" on repo URL
func (r *remoteRepository) FetchRemoteArtifact(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	remoteArtifact, err := r.doFetch(artifact, remoteRepository)
	if err != nil {
		return nil, err
	}

	// fetch parent if present
	if artifact.Parent != nil {
		remoteParent, err := r.doFetch(artifact.Parent, remoteRepository)
		if err != nil {
			return nil, err
		}
		remoteArtifact.Parent = remoteParent
	}

	return remoteArtifact, nil
}

func (r *remoteRepository) doFetch(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	searchPath, err := artifact.SearchPath()
	if err != nil {
		return nil, err
	}

	res, err := http.Get(fmt.Sprintf("%s/%s", remoteRepository, searchPath))
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to fetch POM [%s] from configured search repositories",
			artifact.GetMavenCoords())
	}
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"failed to fetch POM [%s] from configured search repositories",
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

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

func (r *remoteRepository) FetchRemoteArtifact(artifact *Artifact, remoteRepository string) (*Artifact, error) {
	searchPath, err := artifact.SearchPath()
	if err != nil {
		return nil, err
	}
	res, err := http.Get(remoteRepository + searchPath)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Failed to fetch POM [%s] from configured search repositories",
			artifact.AsString())
	}
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"Failed to fetch POM [%s] from configured search repositories",
			artifact.AsString()))
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

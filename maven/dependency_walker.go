package maven

import (
	"errors"
	"fmt"
	"net/http"
)

type DependencyWalker struct {
	Repositories []string
}

func (w *DependencyWalker) TraversePOM(pom *MavenPom) ([]MavenArtifact, error) {
	repository := w.Repositories[0]
	searchPath, err := pom.SearchPath()
	if err != nil {
		panic(err)
	}

	res, err := http.Get(repository + searchPath)
	if err != nil || res.StatusCode != 200 {
		fmt.Println("Request URL : ", res.Request.URL)
		fmt.Println("Response status : ", res.StatusCode)
		panic(err)
	}

	deps := make([]MavenArtifact, 0)
	//deps, err := w.VerifyDependencies(pom.Dependencies)
	//if err != nil {
	//panic(err)
	//}

	deps = append(deps, MavenArtifact{
		GroupID:    pom.GroupID,
		ArtifactID: pom.ArtifactID,
		Version:    pom.Version,
		Repository: repository,
	})

	return deps, nil
}

func (w *DependencyWalker) VerifyDependencies(deps []MavenArtifact) ([]MavenArtifact, error) {
	return nil, errors.New("To be implemented")
}

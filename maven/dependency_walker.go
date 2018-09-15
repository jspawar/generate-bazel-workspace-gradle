package maven

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jspawar/generate-bazel-workspace-gradle/logging"
	"github.com/pkg/errors"
)

var (
	logger = logging.Logger
)

type DependencyWalker struct {
	Repositories []string
	cache        map[string]string
}

func (w *DependencyWalker) TraversePOM(pom *Artifact) ([]Artifact, error) {
	repository := w.Repositories[0]
	searchPath, err := pom.SearchPath()
	if err != nil {
		panic(err)
	}

	logger.Debug().Msgf("Searching for POM in repository : %s", repository)
	res, err := http.Get(repository + searchPath)
	if err != nil || res.StatusCode != 200 {
		panic(err)
	}
	defer res.Body.Close()

	pom.Repository = repository
	deps := []Artifact{*pom}
	w.cache = map[string]string{deps[0].AsString(): repository}

	logger.Debug().Msg("Traversing dependencies...")
	for _, dep := range pom.Dependencies {
		logger.Debug().Msgf("Traversing dependency : %s", dep.AsString())
		traversedDeps, err := w.traverseArtifact(*dep)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Failed to fetch dependency for POM [%s] from configured search repositories",
				pom.AsString())
		}
		deps = append(deps, traversedDeps...)
	}

	return deps, nil
}

func (w *DependencyWalker) traverseArtifact(artifact Artifact) ([]Artifact, error) {
	// check cache to avoid unnecessary traversal
	if _, isCached := w.cache[artifact.AsString()]; isCached {
		logger.Debug().Msgf("Artifact already discovered : %s", artifact.AsString())
		return nil, nil
	}
	// TODO: move this check up?
	if !artifact.IsValid() {
		return nil, errors.New("Invalid Maven artifact definition : " + artifact.AsString())
	}

	repository := w.Repositories[0]
	searchPath, err := artifact.SearchPath()
	if err != nil {
		panic(err)
	}

	logger.Debug().Msgf("Searching for artifact [%s] in repository : %s", artifact.AsString(), repository)
	res, err := http.Get(repository + searchPath)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Error attempting to find artifact [%s] in any of the search repositories",
			artifact.AsString(), repository)
	}
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf(
			"Failed to find artifact [%s] in any of the search repositories",
			artifact.AsString()))
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	body := string(bs)
	pomPath := fmt.Sprintf("%s-%s.pom", artifact.ArtifactID, artifact.Version)
	if !strings.Contains(body, pomPath) {
		return nil, errors.New("Failed to find POM in any repository for artifact : " + artifact.AsString())
	}
	// can safely add this artifact to result slice
	w.cache[artifact.AsString()] = repository
	artifact.Repository = repository
	deps := []Artifact{artifact}

	logger.Debug().Msgf("Reading POM at : %s", repository+searchPath+pomPath)
	res.Body.Close()
	res, err = http.Get(repository + searchPath + pomPath)
	if err != nil || res.StatusCode != 200 {
		panic(err)
	}
	defer res.Body.Close()

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	artifactPom, err := UnmarshalPOM(bs)
	if err != nil {
		panic(err)
	}

	for _, dep := range artifactPom.Dependencies {
		traversedDeps, err := w.traverseArtifact(*dep)
		if err != nil {
			return nil, err
		}
		// TODO: add these to the input artifact's dependency list?
		deps = append(deps, traversedDeps...)
	}

	return deps, nil
}

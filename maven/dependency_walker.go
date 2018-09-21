package maven

import (
	_ "github.com/jspawar/generate-bazel-workspace-gradle/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	logger = zap.S()
)

// TODO: refactor this to instead have an array of `RemoteRepository` instances which are constructed with the actual remote's URL
type DependencyWalker struct {
	Repositories []string
	RemoteRepository
	cache        map[string]string
}

func (w *DependencyWalker) TraversePOM(pom *Artifact) ([]Artifact, error) {
	repository := w.Repositories[0]

	logger.Debugf("Searching for POM in repository : %s", repository)
	remotePom, err := w.RemoteRepository.FetchRemoteArtifact(pom, repository)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Failed to traverse POM [%s] with configured search repositories",
			pom.GetMavenCoords())
	}
	pom = remotePom

	// initialize cache
	pom.Repository = repository
	deps := []Artifact{*pom}
	w.cache = map[string]string{deps[0].GetMavenCoords(): repository}

	logger.Debug("Traversing dependencies...")
	for _, dep := range pom.Dependencies {
		logger.Debugf("Traversing dependency : %s", dep.GetMavenCoords())
		traversedDeps, err := w.traverseArtifact(*dep)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Failed to fetch POM [%s] from configured search repositories",
				dep.GetMavenCoords())
		}
		if traversedDeps != nil {
			deps = append(deps, traversedDeps...)
		}
	}

	return deps, nil
}

func (w *DependencyWalker) traverseArtifact(artifact Artifact) ([]Artifact, error) {
	// check cache to avoid unnecessary traversal
	if _, isCached := w.cache[artifact.GetMavenCoords()]; isCached {
		logger.Debugf("Artifact already discovered : %s", artifact.GetMavenCoords())
		return nil, nil
	}

	repository := w.Repositories[0]

	logger.Debugf("Searching for artifact [%s] in repository : %s", artifact.GetMavenCoords(), repository)
	remoteArtifact, err := w.RemoteRepository.FetchRemoteArtifact(&artifact, repository)
	if err != nil {
		return nil, err
	}
	artifact = *remoteArtifact

	// can safely add this artifact to result slice
	w.cache[artifact.GetMavenCoords()] = repository
	artifact.Repository = repository
	deps := []Artifact{artifact}

	for _, dep := range artifact.Dependencies {
		traversedDeps, err := w.traverseArtifact(*dep)
		if err != nil {
			return nil, err
		}
		// TODO: add these to the input artifact's dependency list?
		deps = append(deps, traversedDeps...)
	}

	return deps, nil
}

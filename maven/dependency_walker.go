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
	cache        map[string]string
	RemoteRepository
}

func (w *DependencyWalker) TraversePOM(pom *Artifact) (*Artifact, error) {
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
	deps := make([]*Artifact, 0)
	// TODO: replace keys for cache with something w/o version like "GetBazelRule" for now
	w.cache = map[string]string{pom.GetMavenCoords(): repository}

	logger.Debug("Traversing dependencies...")
	for _, dep := range pom.Dependencies {
		logger.Debugf("Traversing dependency : %s", dep.GetMavenCoords())
		traversedDep, err := w.traverseArtifact(dep)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Failed to fetch POM [%s] from configured search repositories",
				dep.GetMavenCoords())
		}
		if traversedDep != nil {
			deps = append(deps, traversedDep)
		}
	}
	pom.Dependencies = deps

	return pom, nil
}

func (w *DependencyWalker) traverseArtifact(artifact *Artifact) (*Artifact, error) {
	// check cache to avoid unnecessary traversal
	// TODO: replace keys for cache with something w/o version like "GetBazelRule" for now
	if _, isCached := w.cache[artifact.GetMavenCoords()]; isCached {
		logger.Debugf("Artifact already discovered : %s", artifact.GetMavenCoords())
		// TODO: sufficient to return nil and not append this to list of dependencies for caller?
		return nil, nil
	}

	repository := w.Repositories[0]

	logger.Debugf("Searching for artifact [%s] in repository : %s", artifact.GetMavenCoords(), repository)
	remoteArtifact, err := w.RemoteRepository.FetchRemoteArtifact(artifact, repository)
	if err != nil {
		return nil, err
	}
	artifact = remoteArtifact

	// can safely add this artifact to result slice
	// TODO: replace keys for cache with something w/o version like "GetBazelRule" for now
	w.cache[artifact.GetMavenCoords()] = repository
	artifact.Repository = repository
	deps := make([]*Artifact, 0)

	for _, dep := range artifact.Dependencies {
		traversedDep, err := w.traverseArtifact(dep)
		if err != nil {
			return nil, err
		}
		// only append to list of remote dependencies if it hasn't been discovered already
		if traversedDep != nil {
			deps = append(deps, traversedDep)
		}
	}
	artifact.Dependencies = deps

	return artifact, nil
}

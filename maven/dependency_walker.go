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
	remotePom, err := w.RemoteRepository.FetchRemoteModel(pom, repository)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Failed to traverse POM [%s] with configured search repositories",
			pom.GetMavenCoords())
	}
	pom = remotePom

	// initialize cache
	pom.Repository = repository
	deps := make([]*Artifact, 0)
	w.cache = map[string]string{pom.GetBazelRule(): repository}

	logger.Debug("Traversing dependencies...")
	for _, dep := range pom.Dependencies {
		// ignore test dependencies for now
		if !dep.Optional && pom.Scope != "test" {
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
	}
	pom.Dependencies = deps

	return pom, nil
}

func (w *DependencyWalker) traverseArtifact(artifact *Artifact) (*Artifact, error) {
	// check cache to avoid unnecessary traversal
	if _, isCached := w.cache[artifact.GetBazelRule()]; isCached {
		logger.Debugf("Artifact already discovered : %s", artifact.GetMavenCoords())
		// TODO: sufficient to return nil and not append this to list of dependencies for caller?
		return nil, nil
	}

	repository := w.Repositories[0]

	logger.Infof("Searching for artifact [%s] in repository : %s", artifact.GetMavenCoords(), repository)
	remoteArtifact, err := w.RemoteRepository.FetchRemoteModel(artifact, repository)
	if err != nil {
		return nil, err
	}
	artifact = remoteArtifact

	// can safely add this artifact to result slice
	w.cache[artifact.GetBazelRule()] = repository
	artifact.Repository = repository
	deps := make([]*Artifact, 0)

	for _, dep := range artifact.Dependencies {
		// ignore test dependencies for now
		if !dep.Optional && artifact.Scope != "test" {
			traversedDep, err := w.traverseArtifact(dep)
			if err != nil {
				return nil, err
			}
			// only append to list of remote dependencies if it hasn't been discovered already
			if traversedDep != nil {
				deps = append(deps, traversedDep)
			}
		}
	}
	artifact.Dependencies = deps

	return artifact, nil
}

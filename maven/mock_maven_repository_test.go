package maven_test

import (
	"fmt"
	"net/http"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	"encoding/xml"
)

// TODO: write tests for this test utility?
type MockMavenRepository struct {
	server    *http.Server
	Artifacts []maven.Artifact
}

func (m *MockMavenRepository) Start() {
	mux := http.NewServeMux()
	for _, a := range m.Artifacts {
		m.registerArtifactHandler(mux, a)
	}

	m.server = &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error closing out server:", err)
		}
	}()
}

func (m *MockMavenRepository) Stop() {
	if m.server != nil {
		m.server.Close()
	}
}

func (m *MockMavenRepository) registerArtifactHandler(mux *http.ServeMux, artifact maven.Artifact) error {
	// serve parent of artifact
	if artifact.Parent != nil {
		if err := m.registerArtifactHandler(mux, *artifact.Parent); err != nil {
			return err
		}
	}

	path, err := artifact.SearchPath()
	if err != nil {
		return err
	}

	artifact.XMLName.Local = "project"
	bs, err := xml.MarshalIndent(artifact, "", "    ")
	if err != nil {
		return err
	}

	mux.Handle("/"+path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/xml")
		_, err = w.Write([]byte(xml.Header))
		_, err = w.Write(bs)
	}))

	// serve dependencies of artifact
	for _, dep := range artifact.Dependencies {
		if err := m.registerArtifactHandler(mux, *dep); err != nil {
			return err
		}
	}

	return nil
}

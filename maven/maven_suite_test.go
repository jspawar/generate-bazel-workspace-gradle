package maven_test

import (
	"testing"

	"encoding/xml"
	"github.com/jspawar/generate-bazel-workspace-gradle/maven"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"net/http"
	"net/http/httptest"
)

func TestMaven(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maven Suite")
}

var _ = BeforeSuite(func() {
	format.TruncatedDiff = false
})

func initMockServer(mocks []maven.Artifact, metadata... maven.Metadata) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mocks == nil || len(mocks) < 1 {
			w.WriteHeader(404)
			return
		}

		// serve all input mock responses
		for _, a := range mocks {
			p := a.SearchPath()
			if "/"+p == r.URL.Path {
				// serialize POM object and return
				bs, err := xml.Marshal(a)
				if err != nil {
					w.WriteHeader(500)
					return
				}
				if _, err := w.Write(bs); err != nil {
					w.WriteHeader(500)
				}
				w.Header().Add("Content-Type", "text/xml")
				return
			}
		}

		// serve metadata
		for _, m := range metadata {
			a := &maven.Artifact{GroupID: m.GroupID, ArtifactID: m.ArtifactID}
			p := a.MetadataPath()
			if "/"+p == r.URL.Path {
				// serialize metadata object and return
				bs, err := xml.Marshal(m)
				if err != nil {
					w.WriteHeader(500)
					return
				}
				if _, err := w.Write(bs); err != nil {
					w.WriteHeader(500)
				}
				w.Header().Add("Content-Type", "text/xml")
				return
			}
		}

		// else request doesn't match configured mocks
		w.WriteHeader(404)
	}))
}

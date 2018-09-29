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

func initMockServer(mockResponse *maven.Artifact) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := mockResponse

		// return "not found" if mock response not configured
		if mockResponse == nil {
			w.WriteHeader(404)
			return
		}
		// return "not found" if request doesn't match with mock response or its parent
		p := mockResponse.SearchPath()
		if mockResponse.Parent == nil {
			if "/"+p != r.URL.Path {
				w.WriteHeader(404)
				return
			}
		} else {
			pp := mockResponse.Parent.SearchPath()
			if "/"+p != r.URL.Path && "/"+pp != r.URL.Path {
				w.WriteHeader(404)
				return
			} else if "/"+pp == r.URL.Path {
				res = mockResponse.Parent
			}
		}

		// serialize POM object and return
		bs, err := xml.Marshal(res)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		if _, err := w.Write(bs); err != nil {
			w.WriteHeader(500)
		}
		w.Header().Add("Content-Type", "text/xml")
	}))
}

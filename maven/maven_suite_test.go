package maven_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestMaven(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maven Suite")
}

var _ = BeforeSuite(func() {
	initLogger()
})

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
}

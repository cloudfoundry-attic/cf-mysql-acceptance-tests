package service_test

import (
	"testing"

	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var (
	integrationConfig = helpers.LoadConfig()
)

func TestService(t *testing.T) {
	helpers.PrepareAndRunTests("Service", &integrationConfig, t)
}

func appURI(appname string) string {
	return "http://" + appname + "." + integrationConfig.AppsDomain
}

func curling(args ...string) func() *gexec.Session {
	return func() *gexec.Session {
		return Curl(args...)
	}
}

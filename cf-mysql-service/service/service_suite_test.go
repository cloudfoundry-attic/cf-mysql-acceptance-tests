package service_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
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

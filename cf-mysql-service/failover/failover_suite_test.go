package failover_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var (
	integrationConfig = helpers.LoadConfig()
)

func TestFailover(t *testing.T) {
	helpers.PrepareAndRunTests("Failover", &integrationConfig, t)
}

func appUri(appname string) string {
	return "http://" + appname + "." + integrationConfig.AppsDomain
}

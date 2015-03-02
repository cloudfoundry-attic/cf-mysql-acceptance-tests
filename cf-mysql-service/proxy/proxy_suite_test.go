package proxy_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var integrationConfig = helpers.LoadConfig()

func TestService(t *testing.T) {
	helpers.PrepareAndRunTests("Proxy", &integrationConfig, t)
}

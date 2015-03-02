package dashboard_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var (
	integrationConfig = helpers.LoadConfig()
)

func TestDashboard(t *testing.T) {
	helpers.PrepareAndRunTests("Dashboard", &integrationConfig, t)
}

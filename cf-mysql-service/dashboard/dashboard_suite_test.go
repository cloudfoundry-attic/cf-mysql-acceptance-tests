package dashboard_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var IntegrationConfig = helpers.LoadConfig()

func TestDashboard(t *testing.T) {
	helpers.PrepareAndRunTests("Dashboard", &IntegrationConfig, t)
}

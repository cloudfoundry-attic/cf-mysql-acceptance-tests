package failover_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var IntegrationConfig = helpers.LoadConfig()

func TestFailover(t *testing.T) {
	helpers.PrepareAndRunTests("Failover", &IntegrationConfig, t)
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)
})

func AppUri(appname string) string {
	return "http://" + appname + "." + IntegrationConfig.AppsDomain
}

package proxy_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	helpers "../../helpers"
)

var IntegrationConfig = helpers.LoadConfig()

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", "Proxy"), []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	context_setup.TimeoutScale = IntegrationConfig.TimeoutScale
	SetDefaultEventuallyTimeout(context_setup.ScaledTimeout(10 * time.Second))
})

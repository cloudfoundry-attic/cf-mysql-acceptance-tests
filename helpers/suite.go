package helpers

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	context_setup "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

var TestConfig MysqlIntegrationConfig
var TestEnv context_setup.TestEnvironment

func PrepareAndRunTests(packageName string, t *testing.T) {
    var err error
    TestConfig, err = LoadConfig()
    if err != nil {
        panic("Loading config: " + err.Error())
    }

    err = ValidateConfig(&TestConfig)
    if err != nil {
        panic("Validating config: " + err.Error())
    }

	if TestConfig.SmokeTestsOnly {
		ginkgoconfig.GinkgoConfig.FocusString = "Service instance lifecycle"
	}

	var skipStrings []string

	if ginkgoconfig.GinkgoConfig.SkipString != "" {
		skipStrings = append(skipStrings, ginkgoconfig.GinkgoConfig.SkipString)
	}

	if !TestConfig.IncludeDashboardTests {
		skipStrings = append(skipStrings, "CF Mysql Dashboard")
	}

	if !TestConfig.IncludeFailoverTests {
		skipStrings = append(skipStrings, "CF MySQL Failover")
	}

	if len(skipStrings) > 0 {
		ginkgoconfig.GinkgoConfig.SkipString = strings.Join(skipStrings, "|")
	}

    TestEnv = context_setup.NewTestEnvironment(context_setup.NewContext(TestConfig.IntegrationConfig, "MySQLATS"))

    BeforeEach(TestEnv.BeforeEach)
    AfterEach(TestEnv.AfterEach)

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", packageName), []Reporter{junitReporter})
}

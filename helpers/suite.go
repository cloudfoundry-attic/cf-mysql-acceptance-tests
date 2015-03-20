package helpers

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
)

var TestConfig MysqlIntegrationConfig
var TestContext services.Context

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
		ginkgoconfig.GinkgoConfig.FocusString = "Allows users"
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

	TestContext = services.NewContext(TestConfig.Config, "MySQLATS")

	BeforeEach(TestContext.Setup)
	AfterEach(TestContext.Teardown)

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", packageName), []Reporter{junitReporter})
}

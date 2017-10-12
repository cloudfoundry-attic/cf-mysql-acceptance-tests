package lifecycle_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	"os"
)

var _ = Describe("P-MySQL Lifecycle Tests", func() {
	var sinatraPath = "../../assets/sinatra_app"
	var springPath = "../../assets/cipher_finder"

	var enableServiceAccessToOrg func(string, string)
	var assertAppIsRunning func(string)
	var createBindAndStartApp func(string, string, string, string)
	var cleanupServiceInstance func(string, string)

	It("Lists all public plans in cf marketplace", func() {
		marketplaceCmd := runner.NewCmdRunner(Cf("m"), helpers.TestContext.LongTimeout()).Run()
		marketplaceOutput := marketplaceCmd.Out.Contents()
		for _, plan := range helpers.TestConfig.Plans {
			if plan.Private == false {
				Expect(marketplaceOutput).To(MatchRegexp("%v.*%v", helpers.TestConfig.ServiceName, plan.Name))
			}
		}
	})

	It("Does not list any private plans in cf marketplace", func() {
		marketplaceCmd := runner.NewCmdRunner(Cf("m"), helpers.TestContext.LongTimeout()).Run()
		marketplaceOutput := marketplaceCmd.Out.Contents()
		for _, plan := range helpers.TestConfig.Plans {
			if plan.Private == true {
				Expect(marketplaceOutput).ToNot(MatchRegexp("%v.*%v", helpers.TestConfig.ServiceName, plan.Name))
			}
		}
	})

	Describe("When pushing an app", func() {
		var appName, serviceInstanceName string
		var plan helpers.Plan

		BeforeEach(func() {
			appName = RandomName()
			serviceInstanceName = RandomName()

			if len(helpers.TestConfig.Plans) > 0 {
				plan = helpers.TestConfig.Plans[0]
			} else {
				Skip("Skipping due to lack of plans.")
			}

			enableServiceAccessToOrg(helpers.TestConfig.ServiceName, helpers.TestContext.RegularUserContext().Org)
		})

		AfterEach(func() {
			cleanupServiceInstance(appName, serviceInstanceName)
		})

		It("Allows users to create, bind, write to, read from, unbind, and destroy a service instance for the each plan", func() {
			pushCmd := runner.NewCmdRunner(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-b", "ruby_buildpack", "-d", helpers.TestConfig.AppsDomain, "-no-start"), helpers.TestContext.LongTimeout()).Run()
			Expect(pushCmd).To(Say("OK"))

			uri := fmt.Sprintf("%s/service/mysql/%s/mykey", helpers.TestConfig.AppURI(appName), serviceInstanceName)

			createBindAndStartApp(helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName, appName)

			fmt.Printf("\n*** Posting to url: %s\n", uri)
			curlCmd := runner.NewCmdRunner(runner.Curl("-k", "-d", "myvalue", uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("myvalue"))

			fmt.Printf("\n*** Curling url: %s\n", uri)
			curlCmd = runner.NewCmdRunner(runner.Curl("-k", uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("myvalue"))
		})

		It("Guarantees a TLS connection to a simple Spring app", func() {
			if !helpers.TestConfig.EnableTlsTests {
				Skip("Skipping TLS tests as TLS is not enabled.")
			}

			os.MkdirAll(fmt.Sprintf("%s/build/libs/", springPath), 0700)
			os.Link("/var/vcap/packages/acceptance-tests/cipher_finder/cipher_finder.jar", fmt.Sprintf("%s/build/libs/cipher_finder.jar", springPath))

			// cf push cipher-finder -no-start
			pushCmd := runner.NewCmdRunner(Cf("push", appName, "-m", "1G", "-f", fmt.Sprintf("%s/manifest.yml", springPath), "-d", helpers.TestConfig.AppsDomain, "-b", "java_buildpack", "-no-start"), helpers.TestContext.LongTimeout()).Run()
			Expect(pushCmd).To(Say("OK"))

			// create-service & bind-service & start & assertAppIsRunning
			createBindAndStartApp(helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName, appName)

			// curl app on only endpoint to return active connection's cipher
			uri := fmt.Sprintf("%s/ciphers", helpers.TestConfig.AppURI(appName))
			fmt.Printf("\n*** GET curl to url: %s\n", uri)
			curlCmd := runner.NewCmdRunner(runner.Curl("-k", uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(`{"cipher_used":"([^"]+)"}`))
		})
	})

	enableServiceAccessToOrg = func(serviceName string, org string) {
		AsUser(helpers.TestContext.AdminUserContext(), helpers.TestContext.ShortTimeout(), func() {
			runner.NewCmdRunner(Cf("enable-service-access", serviceName, "-o", org), helpers.TestContext.ShortTimeout()).Run()
		})
	}

	createBindAndStartApp = func(serviceName string, planName string, serviceInstanceName string, appName string) {
		runner.NewCmdRunner(Cf("create-service", serviceName, planName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()

		runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
		runner.NewCmdRunner(Cf("start", appName), helpers.TestContext.LongTimeout()).Run()
		assertAppIsRunning(appName)

	}

	assertAppIsRunning = func(appName string) {
		pingURI := helpers.TestConfig.AppURI(appName) + "/ping"
		fmt.Println("\n*** Checking that the app is responding at url: ", pingURI)

		runner.NewCmdRunner(runner.Curl("-k", pingURI), helpers.TestContext.ShortTimeout()).WithAttempts(3).WithOutput("OK").Run()
	}

	cleanupServiceInstance = func(appName string, serviceInstanceName string) {
		runner.NewCmdRunner(Cf("unbind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
		runner.NewCmdRunner(Cf("delete-service", "-f", serviceInstanceName), helpers.TestContext.LongTimeout()).Run()

		runner.NewCmdRunner(Cf("delete", appName, "-f"), helpers.TestContext.LongTimeout()).Run()
	}
})

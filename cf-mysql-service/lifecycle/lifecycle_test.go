package lifecycle_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"os"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

var _ = Describe("P-MySQL Lifecycle Tests", func() {
	var sinatraPath = "../../assets/sinatra_app"
	var springPath = "../../assets/cipher_finder"

	var enableServiceAccessToOrg func(string, string)
	var createBindAndStartApp func(string, string, string, string, helpers.Pinger)
	var cleanupServiceInstance func(string, string)

	It("Lists all public plans in cf marketplace", func() {
		marketplaceCmd := cf.Cf("m").Wait(helpers.TestContext.LongTimeout())
		Expect(marketplaceCmd).To(Exit(0))

		marketplaceOutput := marketplaceCmd.Out.Contents()
		for _, plan := range helpers.TestConfig.Plans {
			if plan.Private == false {
				Expect(marketplaceOutput).To(MatchRegexp("%v.*%v", helpers.TestConfig.ServiceName, plan.Name))
			}
		}
	})

	It("Does not list any private plans in cf marketplace", func() {
		if helpers.TestConfig.CFConfig.UseExistingOrganization {
			Skip("Skipping private plan test due to use of existing org")
		}

		marketplaceCmd := cf.Cf("m").Wait(helpers.TestContext.LongTimeout())
		Expect(marketplaceCmd).To(Exit(0))

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
		var sinatraAppClient helpers.SinatraAppClient
		var cipherFinderAppClient helpers.CipherFinderClient

		BeforeEach(func() {
			appName = generator.PrefixedRandomName("lifecycle", "")
			serviceInstanceName = generator.PrefixedRandomName("lifecycle", "")

			if len(helpers.TestConfig.Plans) > 0 {
				plan = helpers.TestConfig.Plans[0]
			} else {
				Skip("Skipping due to lack of plans.")
			}

			enableServiceAccessToOrg(helpers.TestConfig.ServiceName, helpers.TestContext.RegularUserContext().Org)
			sinatraAppClient = helpers.NewSinatraAppClient(helpers.TestConfig.AppURI(appName), serviceInstanceName, helpers.TestConfig.CFConfig.SkipSSLValidation)
			cipherFinderAppClient = helpers.NewCipherFinderClient(helpers.TestConfig.AppURI(appName), helpers.TestConfig.CFConfig.SkipSSLValidation)
		})

		AfterEach(func() {
			cleanupServiceInstance(appName, serviceInstanceName)
		})

		It("Allows users to create, bind, write to, read from, unbind, and destroy a service instance for the each plan", func() {
			Expect(cf.Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-b", "ruby_buildpack", "-d", helpers.TestConfig.CFConfig.AppsDomain, "--no-start").
				Wait(helpers.TestContext.LongTimeout())).
				To(Exit(0))

			createBindAndStartApp(helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName, appName, sinatraAppClient)

			fmt.Printf("\n*** Posting to app\n")
			msg, err := sinatraAppClient.Set("mykey", "myvalue")
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring("myvalue"))

			fmt.Printf("\n*** Curling app\n")
			msg, err = sinatraAppClient.Get("mykey")
			Expect(msg).To(ContainSubstring("myvalue"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Guarantees a TLS connection to a simple Spring app", func() {
			if !helpers.TestConfig.EnableTlsTests {
				Skip("Skipping TLS tests as TLS is not enabled.")
			}

			os.MkdirAll(fmt.Sprintf("%s/build/libs/", springPath), 0700)
			os.Link("/var/vcap/packages/acceptance-tests/cipher_finder/cipher_finder.jar", fmt.Sprintf("%s/build/libs/cipher_finder.jar", springPath))

			// cf push cipher-finder --no-start
			Expect(cf.Cf("push", appName, "-m", "1G", "-f", fmt.Sprintf("%s/manifest.yml", springPath), "-d", helpers.TestConfig.CFConfig.AppsDomain, "--no-start").
				Wait(helpers.TestContext.LongTimeout())).
				To(Exit(0))

			// create-service & bind-service & start & assertAppIsRunning
			createBindAndStartApp(helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName, appName, cipherFinderAppClient)

			fmt.Printf("\n*** GET curl to url\n")
			cipher, err := cipherFinderAppClient.Ciphers()
			Expect(err).NotTo(HaveOccurred())
			Expect(cipher).To(Equal("AES256-SHA256"))
		})

		enableServiceAccessToOrg = func(serviceName string, org string) {
			workflowhelpers.AsUser(helpers.TestContext.AdminUserContext(), helpers.TestContext.ShortTimeout(), func() {
				cf.Cf("enable-service-access", serviceName, "-o", org).Wait(helpers.TestContext.ShortTimeout())
			})
		}

		createBindAndStartApp = func(serviceName string, planName string, serviceInstanceName string, appName string, appClient helpers.Pinger) {
			Expect(cf.Cf("create-service", serviceName, planName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("bind-service", appName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("start", appName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			err := appClient.Ping()
			Expect(err).NotTo(HaveOccurred())
		}

		cleanupServiceInstance = func(appName string, serviceInstanceName string) {
			cf.Cf("unbind-service", appName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())
			cf.Cf("delete-service", "-f", serviceInstanceName).Wait(helpers.TestContext.LongTimeout())

			cf.Cf("delete", appName, "-f").Wait(helpers.TestContext.LongTimeout())
		}
	})
})

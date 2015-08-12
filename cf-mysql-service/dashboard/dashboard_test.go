package dashboard_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	"fmt"
	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"time"
)

var _ = Describe("CF Mysql Dashboard", func() {
	var (
		page                *Page
		driver              *WebDriver
		dashboardUrl        string
		username            string
		password            string
		serviceInstanceName string
	)

	var getDashboardUrl = func(serviceInstanceInfo map[string]interface{}) string {
		resources := serviceInstanceInfo["resources"].([]interface{})
		resource := resources[0].(map[string]interface{})
		entity := resource["entity"].(map[string]interface{})

		return entity["dashboard_url"].(string)
	}

	BeforeEach(func() {

		driver = PhantomJS()
		Expect(driver.Start()).To(Succeed())

		var err error
		page, err = driver.NewPage()
		Expect(err).ToNot(HaveOccurred())

		serviceInstanceName = RandomName()
		planName := helpers.TestConfig.Plans[0].Name

		By("Creating service")
		runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, planName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()

		By("Verifing service instance exists")
		var serviceInstanceInfo map[string]interface{}
		serviceInfoCmd := runner.NewCmdRunner(Cf("curl", "/v2/service_instances?q=name:"+serviceInstanceName, "-k"), helpers.TestContext.ShortTimeout()).Run()
		err = json.Unmarshal(serviceInfoCmd.Buffer().Contents(), &serviceInstanceInfo)
		Expect(err).ShouldNot(HaveOccurred())

		dashboardUrl = getDashboardUrl(serviceInstanceInfo)
		regularUserContext := helpers.TestContext.RegularUserContext()
		username = regularUserContext.Username
		password = regularUserContext.Password
	})

	AfterEach(func() {
		By("Stopping Webdriver")
		Expect(page.Destroy()).To(Succeed())

		Expect(driver.Stop()).To(Succeed())

		runner.NewCmdRunner(Cf("delete-service", "-f", serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
	})

	It("Login via dashboard url", func() {
		By("navigate to dashboard url", func() {
			time.Sleep(time.Second * 10)
			err := page.Navigate(dashboardUrl)
			Expect(err).ToNot(HaveOccurred())
			content, err := page.HTML()
			Expect(err).ToNot(HaveOccurred())
			fmt.Printf("Login Page: %s", content)
			Eventually(page.Find("h1"), time.Second*5).Should(HaveText("Welcome!"))
		})

		By("submit login credentials", func() {
			Expect(page.Find("input[name=username]").Fill(username)).To(Succeed())
			Expect(page.Find("input[name=password]").Fill(password)).To(Succeed())
			Expect(page.Find("form").Submit()).To(Succeed())
		})

		By("authorize broker application", func() {
			Eventually(page.Find("h1"), time.Second*5).Should(HaveText("Application Authorization"))
			Expect(page.Find("button#authorize").Click()).To(Succeed())
		})

		By("end up on dashboard", func() {
			Eventually(page, time.Second*5).Should(HaveTitle("MySQL Management Dashboard"))
		})
	})
})

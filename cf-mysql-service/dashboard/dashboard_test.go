package dashboard_test

import (
	"encoding/json"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	"github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core"
	. "github.com/sclevine/agouti/dsl"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Feature("CF Mysql Dashboard", func() {
	var (
		page                Page
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

	Background(func() {
		StartChrome()

		serviceInstanceName = RandomName()
		planName := integrationConfig.Plans[0].Name

		Step("Creating service")
		ExecWithTimeout(Cf("create-service", integrationConfig.ServiceName, planName, serviceInstanceName), integrationConfig.LongTimeout())

		Step("Verifing service instance exists")
		var serviceInstanceInfo map[string]interface{}
		serviceInfoCmd := ExecWithTimeout(Cf("curl", "/v2/service_instances?q=name:"+serviceInstanceName), integrationConfig.ShortTimeout())
		err := json.Unmarshal(serviceInfoCmd.Buffer().Contents(), &serviceInstanceInfo)
		Expect(err).ShouldNot(HaveOccurred())

		dashboardUrl = getDashboardUrl(serviceInstanceInfo)
		username = context_setup.RegularUserContext.Username
		password = context_setup.RegularUserContext.Password

		page = CreatePage()
		page.Size(640, 480)
	})

	AfterEach(func() {
		page.Destroy()

		ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), integrationConfig.LongTimeout())
		StopWebdriver()
	})

	Scenario("Login via dashboard url", func() {
		Step("navigate to dashboard url", func() {
			page.Navigate(dashboardUrl)
			Eventually(page.Find("h1")).Should(HaveText("Welcome!"))
		})

		Step("submit login credentials", func() {
			Fill(page.Find("input[name=username]"), username)
			Fill(page.Find("input[name=password]"), password)
			Submit(page.Find("form"))
		})

		Step("authorize broker application", func() {
			Eventually(page.Find("h1")).Should(HaveText("Application Authorization"))
			Click(page.Find("button#authorize"))
		})

		Step("end up on dashboard", func() {
			Eventually(page).Should(HaveTitle("MySQL Management Dashboard"))
		})
	})
})

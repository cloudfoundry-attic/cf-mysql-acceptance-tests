package service_test

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var _ = Describe("P-MySQL Service", func() {
	var sinatraPath = "../../assets/sinatra_app"

	AssertAppIsRunning := func(appName string) {
		pingUri := appURI(appName) + "/ping"
		fmt.Println("Checking that the app is responding at url: ", pingUri)
		curlCmd := ExecWithTimeout(Curl(pingUri), integrationConfig.ShortTimeout())
		Expect(curlCmd).To(Say("OK"))
	}

	It("Registers a route", func() {
		uri := fmt.Sprintf("http://%s/v2/catalog", integrationConfig.BrokerHost)

		fmt.Printf("Curling url: %s\n", uri)
		curlCmd := ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
		Expect(curlCmd).To(Say("HTTP Basic: Access denied."))
	})

	Describe("Service instance lifecycle", func() {
		var appName string

		BeforeEach(func() {
			appName = RandomName()
			pushCmd := ExecWithTimeout(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), integrationConfig.LongTimeout())
			Expect(pushCmd).To(Say("OK"))
		})

		AfterEach(func() {
			ExecWithTimeout(Cf("delete", appName, "-f"), integrationConfig.LongTimeout())
		})

		AssertLifeCycleBehavior := func(PlanName string) {
			It("Allows users to create, bind, write to, read from, unbind, and destroy a service instance for a plan", func() {
				serviceInstanceName := RandomName()
				uri := appURI(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"

				ExecWithTimeout(Cf("create-service", integrationConfig.ServiceName, PlanName, serviceInstanceName), integrationConfig.LongTimeout())

				ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
				ExecWithTimeout(Cf("start", appName), integrationConfig.LongTimeout())
				AssertAppIsRunning(appName)

				fmt.Printf("Posting to url: %s\n", uri)
				curlCmd := ExecWithTimeout(Curl("-d", "myvalue", uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("myvalue"))

				fmt.Printf("Curling url: %s\n", uri)
				curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("myvalue"))

				ExecWithTimeout(Cf("unbind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
				ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), integrationConfig.LongTimeout())
			})
		}

		Context("using a new service instance", func() {
			for _, plan := range integrationConfig.Plans {
				AssertLifeCycleBehavior(plan.Name)
			}
		})
	})

	Describe("Enforcing MySQL storage and connection quota", func() {
		var appName string
		var serviceInstanceName string
		var quotaEnforcerSleepTime time.Duration

		BeforeEach(func() {
			appName = RandomName()
			serviceInstanceName = RandomName()
			quotaEnforcerSleepTime = 10 * time.Second

			ExecWithTimeout(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), integrationConfig.LongTimeout())
		})

		AfterEach(func() {
			ExecWithTimeout(Cf("unbind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("delete", appName, "-f"), integrationConfig.LongTimeout())
		})

		CreatesBindsAndStartsApp := func(PlanName string) {
			ExecWithTimeout(Cf("create-service", integrationConfig.ServiceName, PlanName, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("start", appName), integrationConfig.LongTimeout())
			AssertAppIsRunning(appName)
		}

		ExceedQuota := func(MaxStorageMb int, appName, serviceInstanceName string) {
			uri := appURI(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"
			writeUri := appURI(appName) + "/service/mysql/" + serviceInstanceName + "/write-bulk-data"

			fmt.Println("*** Exceeding quota")

			mbToWrite := 10
			loopIterations := (MaxStorageMb / mbToWrite)
			if MaxStorageMb%mbToWrite == 0 {
				loopIterations += 1
			}

			for i := 0; i < loopIterations; i += 1 {
				curlCmd := ExecWithTimeout(Curl("-v", "-d", strconv.Itoa(mbToWrite), writeUri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("Database now contains"))
			}

			fmt.Println("*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			value := RandomName()[:20]
			fmt.Println("*** Proving we cannot write")
			curlCmd := ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
		}

		AssertStorageQuotaBehavior := func(PlanName string, MaxStorageMb int) {
			It("enforces the storage quota for the plan", func() {
				CreatesBindsAndStartsApp(PlanName)

				uri := appURI(appName) + "/service/mysql/context_setup.ScaledTimeout(timeout), retryInterval" + serviceInstanceName + "/mykey"
				deleteUri := appURI(appName) + "/service/mysql/" + serviceInstanceName + "/delete-bulk-data"
				firstValue := RandomName()[:20]
				secondValue := RandomName()[:20]

				fmt.Println("*** Proving we can write")
				curlCmd := ExecWithTimeout(Curl("-d", firstValue, uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(firstValue))

				fmt.Println("*** Proving we can read")
				curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(firstValue))

				ExceedQuota(MaxStorageMb, appName, serviceInstanceName)

				fmt.Println("*** Proving we can read")
				curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(firstValue))

				fmt.Println("*** Deleting below quota")
				curlCmd = ExecWithTimeout(Curl("-d", "20", deleteUri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("Database now contains"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we can write")
				curlCmd = ExecWithTimeout(Curl("-d", secondValue, uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(secondValue))

				fmt.Println("*** Proving we can read")
				curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(secondValue))
			})
		}

		AssertConnectionQuotaBehavior := func(PlanName string, MaxUserConnections int) {
			It("enforces the connection quota for the plan", func() {
				CreatesBindsAndStartsApp(PlanName)

				uri := appURI(appName) + "/connections/mysql/" + serviceInstanceName + "/"
				over_maximum_connection_num := MaxUserConnections + 1

				fmt.Println("*** Proving we can use the max num of connections")

				curlCmd := ExecWithTimeout(Curl(uri+strconv.Itoa(MaxUserConnections)), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("success"))

				fmt.Println("*** Proving the connection quota is enforced")
				curlCmd = ExecWithTimeout(Curl(uri+strconv.Itoa(over_maximum_connection_num)), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("Error"))
			})
		}

		Context("for each plan", func() {
			for _, plan := range integrationConfig.Plans {
				AssertStorageQuotaBehavior(plan.Name, plan.MaxStorageMb)
				AssertConnectionQuotaBehavior(plan.Name, plan.MaxUserConnections)
			}
		})

		Describe("Upgrading a service instance", func() {
			It("upgrades the instance and enforces the new quota", func() {
				plan := integrationConfig.Plans[0]
				newPlan := integrationConfig.Plans[1]
				CreatesBindsAndStartsApp(plan.Name)
				ExceedQuota(plan.MaxStorageMb, appName, serviceInstanceName)

				fmt.Println("*** Upgrading service instance")
				cfCmd := ExecWithTimeout(Cf("update-service", serviceInstanceName, "-p", newPlan.Name), integrationConfig.LongTimeout())
				Expect(cfCmd).To(Say("OK"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we can write")
				uri := appURI(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"
				value := RandomName()[:20]
				curlCmd := ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say(value))

				ExceedQuota(newPlan.MaxStorageMb, appName, serviceInstanceName)
			})
		})
	})
})

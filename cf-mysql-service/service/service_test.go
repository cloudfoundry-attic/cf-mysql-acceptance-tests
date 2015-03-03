package service_test

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

const (
	// The quota enforcer sleeps for one second between iterations,
	// so sleeping for 20 seconds is sufficient for it have enforced all quotas
	quotaEnforcerSleepTime = 20 * time.Second
)

var _ = Describe("P-MySQL Service", func() {
	sinatraPath := "../../assets/sinatra_app"

	assertAppIsRunning := func(appName string) {
		pingUri := appURI(appName) + "/ping"
		fmt.Println("\n*** Checking that the app is responding at url: ", pingUri)
		curlCmd := ExecWithTimeout(Curl(pingUri), integrationConfig.ShortTimeout())
		Expect(curlCmd).To(Say("OK"))
	}

	It("Registers a route", func() {
		uri := fmt.Sprintf("http://%s/v2/catalog", integrationConfig.BrokerHost)

		fmt.Printf("\n*** Curling url: %s\n", uri)
		curlCmd := ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
		Expect(curlCmd).To(Say("HTTP Basic: Access denied."))
		fmt.Println("Expected failure occured")
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

		Describe("for each plan", func() {
			for _, plan := range integrationConfig.Plans {
				It("Allows users to create, bind, write to, read from, unbind, and destroy a service instance for a plan", func() {
					serviceInstanceName := RandomName()
					uri := fmt.Sprintf("%s/service/mysql/%s/mykey", appURI(appName), serviceInstanceName)

					ExecWithTimeout(Cf("create-service", integrationConfig.ServiceName, plan.Name, serviceInstanceName), integrationConfig.LongTimeout())

					ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
					ExecWithTimeout(Cf("start", appName), integrationConfig.LongTimeout())
					assertAppIsRunning(appName)

					fmt.Printf("\n*** Posting to url: %s\n", uri)
					curlCmd := ExecWithTimeout(Curl("-d", "myvalue", uri), integrationConfig.ShortTimeout())
					Expect(curlCmd).To(Say("myvalue"))

					fmt.Printf("\n*** Curling url: %s\n", uri)
					curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
					Expect(curlCmd).To(Say("myvalue"))

					ExecWithTimeout(Cf("unbind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
					ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), integrationConfig.LongTimeout())
				})
			}
		})
	})

	Describe("Enforcing MySQL storage and connection quota", func() {
		var appName string
		var serviceInstanceName string
		var serviceURI string
		var plan helpers.Plan

		BeforeEach(func() {
			appName = RandomName()
			serviceInstanceName = RandomName()
			plan = integrationConfig.Plans[0]

			serviceURI = fmt.Sprintf("%s/service/mysql/%s", appURI(appName), serviceInstanceName)
			ExecWithTimeout(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), integrationConfig.LongTimeout())
		})

		JustBeforeEach(func() {
			fmt.Printf("Creating service with serviceName: %s, planName: %s, serviceInstanceName: %s\n", integrationConfig.ServiceName, plan.Name, serviceInstanceName)
			ExecWithTimeout(Cf("create-service", integrationConfig.ServiceName, plan.Name, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("start", appName), integrationConfig.LongTimeout())
			assertAppIsRunning(appName)
		})

		AfterEach(func() {
			ExecWithTimeout(Cf("unbind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), integrationConfig.LongTimeout())
			ExecWithTimeout(Cf("delete", appName, "-f"), integrationConfig.LongTimeout())
		})

		ExceedLimit := func(maxStorageMb int) {
			writeUri := fmt.Sprintf("%s/write-bulk-data", serviceURI)

			fmt.Printf("\n*** Exceeding limit of %d\n", maxStorageMb)
			mbToWrite := 10
			loopIterations := (maxStorageMb / mbToWrite)

			for i := 0; i < loopIterations; i++ {
				curlCmd := ExecWithTimeout(Curl("-v", "-d", strconv.Itoa(mbToWrite), writeUri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("Database now contains"))
			}

			remainder := maxStorageMb % mbToWrite
			if remainder != 0 {
				curlCmd := ExecWithTimeout(Curl("-v", "-d", strconv.Itoa(remainder), writeUri), integrationConfig.ShortTimeout())
				Expect(curlCmd).To(Say("Database now contains"))
			}

			// Write a little bit more to guarantee we are over quota
			// as opposed to being exactly at quota,
			// We are not interested in the output because we know we will be over quota.
			ExecWithTimeout(Curl("-v", "-d", strconv.Itoa(1), writeUri), integrationConfig.ShortTimeout())
		}

		// We only need to validate the storage quota enforcer operates as expected over the first plan.
		// Especially important given the plan can be of any size, and we don't want to fill up large databases.
		It("enforces the storage quota for the plan", func() {
			uri := fmt.Sprintf("%s/mykey", serviceURI)
			deleteUri := fmt.Sprintf("%s/delete-bulk-data", serviceURI)
			firstValue := RandomName()[:20]
			secondValue := RandomName()[:20]

			fmt.Println("\n*** Proving we can write")
			curlCmd := ExecWithTimeout(Curl("-d", firstValue, uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say(firstValue))

			fmt.Println("\n*** Proving we can read")
			curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say(firstValue))

			ExceedLimit(plan.MaxStorageMb)

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we cannot write (expect app to fail)")
			value := RandomName()[:20]
			curlCmd = ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
			fmt.Println("Expected failure occured")

			fmt.Println("\n*** Proving we can read")
			curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say(firstValue))

			fmt.Println("\n*** Deleting below quota")
			curlCmd = ExecWithTimeout(Curl("-d", "20", deleteUri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say("Database now contains"))

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we can write")
			curlCmd = ExecWithTimeout(Curl("-d", secondValue, uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say(secondValue))

			fmt.Println("\n*** Proving we can read")
			curlCmd = ExecWithTimeout(Curl(uri), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say(secondValue))
		})

		It("enforces the connection quota for the plan", func() {
			connectionsURI := fmt.Sprintf("%s/connections/mysql/%s/", appURI(appName), serviceInstanceName)

			fmt.Println("\n*** Proving we can use the max num of connections")
			curlCmd := ExecWithTimeout(Curl(connectionsURI+strconv.Itoa(plan.MaxUserConnections)), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say("success"))

			fmt.Println("\n*** Proving the connection quota is enforced")
			curlCmd = ExecWithTimeout(Curl(connectionsURI+strconv.Itoa(plan.MaxUserConnections+1)), integrationConfig.ShortTimeout())
			Expect(curlCmd).To(Say("Error"))
		})

		Describe("Migrating a service instance between plans of different storage quota", func() {
			Context("when upgrading to a larger storage quota", func() {
				var newPlan helpers.Plan

				BeforeEach(func() {
					newPlan = integrationConfig.Plans[1]
				})

				It("enforces the new quota", func() {
					uri := fmt.Sprintf("%s/mykey", serviceURI)
					ExceedLimit(plan.MaxStorageMb)

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we cannot write (expect app to fail)")
					value := RandomName()[:20]
					curlCmd := ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
					Expect(curlCmd).To(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
					fmt.Println("Expected failure occured")

					fmt.Println("\n*** Upgrading service instance")
					cfCmd := ExecWithTimeout(Cf("update-service", serviceInstanceName, "-p", newPlan.Name), integrationConfig.LongTimeout())
					Expect(cfCmd).To(Say("OK"))

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we can write")
					value = RandomName()[:20]
					curlCmd = ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
					Expect(curlCmd).To(Say(value))
				})
			})

			Context("when attempting to downgrade to a smaller storage quota", func() {
				var smallPlan helpers.Plan

				BeforeEach(func() {
					plan = integrationConfig.Plans[1]
					smallPlan = integrationConfig.Plans[0]
				})

				Context("when storage usage is over smaller quota", func() {
					It("disallows downgrade", func() {
						ExceedLimit(smallPlan.MaxStorageMb)

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value := RandomName()[:20]
						uri := fmt.Sprintf("%s/mykey", serviceURI)
						curlCmd := ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
						Expect(curlCmd).To(Say(value))

						fmt.Println("\n*** Downgrading service instance (Expect failure)")
						cfCmd := ExecWithTimeoutForExitCode(Cf("update-service", serviceInstanceName, "-p", smallPlan.Name), integrationConfig.LongTimeout(), 1)
						Expect(cfCmd).To(Say("Service broker error"))
						fmt.Println("Expected failure occured")
					})
				})

				Context("when storage usage is under smaller quota", func() {
					It("allows downgrade", func() {
						ExceedLimit(0)

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value := RandomName()[:20]
						uri := fmt.Sprintf("%s/mykey", serviceURI)
						curlCmd := ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
						Expect(curlCmd).To(Say(value))

						fmt.Println("\n*** Downgrading service instance")
						cfCmd := ExecWithTimeout(Cf("update-service", serviceInstanceName, "-p", smallPlan.Name), integrationConfig.LongTimeout())
						Expect(cfCmd).To(Say("OK"))

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value = RandomName()[:20]
						curlCmd = ExecWithTimeout(Curl("-d", value, uri), integrationConfig.ShortTimeout())
						Expect(curlCmd).To(Say(value))
					})
				})
			})
		})
	})
})

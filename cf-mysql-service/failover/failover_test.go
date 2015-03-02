package failover_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/sclevine/agouti/dsl"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/partition"
)

const (
	firstKey    = "mykey"
	firstValue  = "myvalue"
	secondKey   = "mysecondkey"
	secondValue = "mysecondvalue"
	planName    = "100mb"

	sinatraPath = "../../assets/sinatra_app"
)

var appName string

func createAndBindService(serviceName, serviceInstanceName, planName string) {
	By("Creating service instance")
	ExecWithTimeout(Cf("create-service", serviceName, planName, serviceInstanceName), integrationConfig.LongTimeout())

	By("Binding app to service instance")
	ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), integrationConfig.LongTimeout())

	By("Restarting app")
	ExecWithTimeout(Cf("restart", appName), integrationConfig.LongTimeout())
}

func assertAppIsRunning(appName string) {
	pingURI := appUri(appName) + "/ping"
	curlCmd := ExecWithTimeout(Curl(pingURI), integrationConfig.ShortTimeout())
	Expect(curlCmd).Should(Say("OK"))
}

func assertWriteToDB(key, value, uri string) {
	curlURI := fmt.Sprintf("%s/%s", uri, key)
	curlCmd := ExecWithTimeout(Curl("-d", value, curlURI), integrationConfig.ShortTimeout())
	Expect(curlCmd).Should(Say(value))
}

func assertReadFromDB(key, value, uri string) {
	curlURI := fmt.Sprintf("%s/%s", uri, key)
	curlCmd := ExecWithTimeout(Curl(curlURI), integrationConfig.ShortTimeout())
	Expect(curlCmd).Should(Say(value))
}

var _ = Feature("CF MySQL Failover", func() {
	BeforeEach(func() {
		appName = generator.RandomName()

		Step("Push an app", func() {
			ExecWithTimeout(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), integrationConfig.LongTimeout())
		})
	})

	Context("when the mysql node is partitioned", func() {
		BeforeEach(func() {
			Expect(integrationConfig.MysqlNodes).NotTo(BeNil())
			Expect(len(integrationConfig.MysqlNodes)).To(BeNumerically(">=", 1))
		})

		AfterEach(func() {
			// Re-introducing a mariadb node once partitioned is unsafe
			// See https://www.pivotaltracker.com/story/show/81974864
			// partition.Off(IntegrationConfig.MysqlNodes[0].SshTunnel)
		})

		Scenario("write/read data before the partition and successfully writes and read it after", func() {
			planName := "100mb"
			serviceInstanceName := generator.RandomName()
			instanceURI := appUri(appName) + "/service/mysql/" + serviceInstanceName

			Step("Create & bind a DB", func() {
				createAndBindService(integrationConfig.ServiceName, serviceInstanceName, planName)
				assertAppIsRunning(appName)
			})

			Step("Start App", func() {
				ExecWithTimeout(Cf("start", appName), integrationConfig.LongTimeout())
				assertAppIsRunning(appName)
			})

			Step("Write a key-value pair to DB", func() {
				assertWriteToDB(firstKey, firstValue, instanceURI)
			})

			Step("Read value from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
			})

			Step("Take down mysql node", func() {
				partition.On(
					integrationConfig.MysqlNodes[0].SshTunnel,
					integrationConfig.MysqlNodes[0].Ip,
				)
			})

			Step("Restart sinatra app to reset connections", func() {
				fmt.Println("Restarting app")
				ExecWithTimeout(Cf("restart", appName), integrationConfig.LongTimeout())
				fmt.Println("Checking whether app is running")
				assertAppIsRunning(appName)
			})

			Step("Write a second key-value pair to DB", func() {
				assertWriteToDB(secondKey, secondValue, instanceURI)
			})

			Step("Read both values from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
				assertReadFromDB(secondKey, secondValue, instanceURI)
			})
		})
	})

	Context("Broker failure", func() {
		var broker0SshTunnel, broker1SshTunnel string

		BeforeEach(func() {
			Expect(integrationConfig.Brokers).NotTo(BeNil())
			Expect(len(integrationConfig.Brokers)).To(BeNumerically(">=", 2))

			broker0SshTunnel = integrationConfig.Brokers[0].SshTunnel
			broker1SshTunnel = integrationConfig.Brokers[1].SshTunnel
		})

		AfterEach(func() {
			partition.Off(broker0SshTunnel)
			partition.Off(broker1SshTunnel)
		})

		Scenario("Broker failure", func() {
			serviceInstanceName := generator.RandomName()
			instanceURI := appUri(appName) + "/service/mysql/" + serviceInstanceName

			// Remove partitions in case previous test did not cleanup correctly
			partition.Off(broker0SshTunnel)
			partition.Off(broker1SshTunnel)

			Step("Take down first broker instance", func() {
				partition.On(broker0SshTunnel, integrationConfig.Brokers[0].Ip)
			})

			Step("Create & bind a DB", func() {
				createAndBindService(integrationConfig.ServiceName, serviceInstanceName, planName)
			})

			Step("Write a key-value pair to DB", func() {
				assertWriteToDB(firstKey, firstValue, instanceURI)
			})

			Step("Read valuefrom DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
			})

			Step("Bring back first broker instance", func() {
				partition.Off(broker0SshTunnel)
			})

			Step("Take down second broker instance", func() {
				partition.On(broker1SshTunnel, integrationConfig.Brokers[1].Ip)
			})

			Step("Create & bind a DB again", func() {
				serviceInstanceName := generator.RandomName()
				createAndBindService(integrationConfig.ServiceName, serviceInstanceName, planName)
			})

			Step("Write a second key-value pair to DB", func() {
				assertWriteToDB(secondKey, secondValue, instanceURI)
			})

			Step("Read both values from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
				assertReadFromDB(secondKey, secondValue, instanceURI)
			})

			Step("Bring back second broker instance", func() {
				partition.Off(broker1SshTunnel)
			})
		})
	})
})

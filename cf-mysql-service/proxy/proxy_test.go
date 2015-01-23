package proxy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var _ = Describe("P-MySQL Proxy", func() {

	It("rejects connections without credentials", func() {
		uri := "http://haproxy-0.p-mysql." + IntegrationConfig.AppsDomain

		Eventually(Curl(uri)).Should(Say("Unauthorized"))
	})

	It("rejects connections with incorrect credentials", func() {
		uri := "http://fakeuser:fakepassword@haproxy-0.p-mysql." + IntegrationConfig.AppsDomain

		Eventually(Curl(uri)).Should(Say("Unauthorized"))
	})

	It("allows connections with correct credentials", func() {
		uri := "http://admin:password@haproxy-0.p-mysql." + IntegrationConfig.AppsDomain

		Eventually(Curl(uri)).Should(Say("Statistics Report for HAProxy"))
	})
})

package proxy_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"io/ioutil"
	"net/url"
)

var _ = Describe("P-MySQL Proxy", func() {

	It("prompts for Basic Auth creds when they aren't provided", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			resp, err := http.Get(url)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}
	})

	It("does not accept bad Basic Auth creds", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth("bad_username", "bad_password")
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}
	})

	It("accepts valid Basic Auth creds", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(
				helpers.TestConfig.Proxy.APIUsername,
				helpers.TestConfig.Proxy.APIPassword,
			)
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	})

	It("has links to the dashboard urls in the aggregator page", func() {
		req, err := http.NewRequest("GET", helpers.TestConfig.Proxy.AggregatorUrl, nil)
		req.SetBasicAuth(
			helpers.TestConfig.Proxy.APIUsername,
			helpers.TestConfig.Proxy.APIPassword,
		)
		resp, err := http.DefaultClient.Do(req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		for _, uri := range helpers.TestConfig.Proxy.DashboardUrls {
			tmphost, _ := url.Parse(uri)
			hostname := tmphost.Host
			Expect(body).To(ContainSubstring(hostname))
		}
	})
})

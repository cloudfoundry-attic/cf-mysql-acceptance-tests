package proxy_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var _ = Describe("P-MySQL Proxy", func() {
	var (
		urlProxy0 string
		urlProxy1 string
	)

	BeforeEach(func() {
		urlProxy0 = fmt.Sprintf(
			"https://proxy-0.%s/v0/backends",
			helpers.TestConfig.Proxy.ExternalHost,
		)

		urlProxy1 = fmt.Sprintf(
			"https://proxy-1.%s/v0/backends",
			helpers.TestConfig.Proxy.ExternalHost,
		)
	})

	var _ = Context("urlProxy0", func() {

		It("prompts for Basic Auth creds when they aren't provided", func() {
			resp, err := http.Get(urlProxy0)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("does not accept bad Basic Auth creds", func() {
			req, err := http.NewRequest("GET", urlProxy0, nil)
			req.SetBasicAuth("bad_username", "bad_password")
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		})

		It("accepts valid Basic Auth creds", func() {
			req, err := http.NewRequest("GET", urlProxy0, nil)
			req.SetBasicAuth(
				helpers.TestConfig.Proxy.APIUsername,
				helpers.TestConfig.Proxy.APIPassword,
			)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

		})

	})

	var _ = Context("urlProxy1", func() {

		It("prompts for Basic Auth creds when they aren't provided", func() {
			resp, err := http.Get(urlProxy1)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("does not accept bad Basic Auth creds", func() {
			req, err := http.NewRequest("GET", urlProxy1, nil)
			req.SetBasicAuth("bad_username", "bad_password")
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))

		})

		It("accepts valid Basic Auth creds", func() {
			req, err := http.NewRequest("GET", urlProxy1, nil)
			req.SetBasicAuth(
				helpers.TestConfig.Proxy.APIUsername,
				helpers.TestConfig.Proxy.APIPassword,
			)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

		})

	})
})

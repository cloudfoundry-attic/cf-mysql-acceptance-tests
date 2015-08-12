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
		url string
	)

	BeforeEach(func() {
		url = fmt.Sprintf(
			"https://proxy-0.%s/v0/backends",
			helpers.TestConfig.Proxy.ExternalHost,
		)
	})

	It("prompts for Basic Auth creds when they aren't provided", func() {
		resp, err := http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
	})

	It("does not accept bad Basic Auth creds", func() {
		req, err := http.NewRequest("GET", url, nil)
		req.SetBasicAuth("bad_username", "bad_password")
		resp, err := http.DefaultClient.Do(req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
	})

	It("accepts valid Basic Auth creds", func() {
		req, err := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(
			helpers.TestConfig.Proxy.APIUsername,
			helpers.TestConfig.Proxy.APIPassword,
		)
		resp, err := http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})

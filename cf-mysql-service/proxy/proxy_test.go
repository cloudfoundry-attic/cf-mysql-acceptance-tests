package proxy_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("P-MySQL Proxy", func() {
	var url string

	BeforeEach(func() {
		url = fmt.Sprintf(
			"http://proxy-0.%s/v0/backends",
			IntegrationConfig.Proxy.ExternalHost,
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
		client := &http.Client{}
		resp, err := client.Do(req)

		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
	})

	It("accepts valid Basic Auth creds", func() {
		req, err := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(
			IntegrationConfig.Proxy.APIUsername,
			IntegrationConfig.Proxy.APIPassword,
		)
		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})

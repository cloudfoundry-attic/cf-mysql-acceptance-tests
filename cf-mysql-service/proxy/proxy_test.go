package proxy_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("P-MySQL Proxy", func() {
	var url string

	BeforeEach(func() {
		url = "http://haproxy-0.p-mysql." + IntegrationConfig.Proxy.Domain
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
		req.SetBasicAuth("admin", IntegrationConfig.Proxy.Password)
		client := &http.Client{}
		resp, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})

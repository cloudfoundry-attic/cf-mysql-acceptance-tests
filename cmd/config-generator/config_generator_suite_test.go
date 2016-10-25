package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"testing"
	"time"
)

var binPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/cmd/configure")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	binPath = string(data)

	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestConfigGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configuration Generator Command Suite")
}

package standalone_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"

	. "github.com/onsi/gomega"
	"github.com/nu7hatch/gouuid"
	"fmt"
	"strings"
)

func TestService(t *testing.T) {
	helpers.PrepareAndRunTests("Standalone", t, false)
}

func uuidWithUnderscores(prefix string) string {
	id, err := uuid.NewV4()
	Expect(err).ToNot(HaveOccurred())
	idString := fmt.Sprintf("%s_%s", prefix, id.String())
	return strings.Replace(idString, "-", "_", -1)
}
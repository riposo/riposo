package config_test

import (
	"testing"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/config"
)

var _ = Describe("Config", func() {
	It("parses", func() {
		conf, err := Parse()
		Expect(err).NotTo(HaveOccurred())
		Expect(conf).To(BeAssignableToTypeOf(&Config{}))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/config")
}

package config_test

import (
	"testing"

	"github.com/riposo/riposo/internal/config"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Config", func() {
	It("parses", func() {
		conf, err := config.Parse()
		Expect(err).NotTo(HaveOccurred())
		Expect(conf).To(BeAssignableToTypeOf(&config.Config{}))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/config")
}

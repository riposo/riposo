package config_test

import (
	"testing"
	"time"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/config"
)

var _ = Describe("Config", func() {
	It("applies defaults", func() {
		conf, err := Parse("", MapEnv{})
		Expect(err).NotTo(HaveOccurred())
		Expect(conf).To(BeAssignableToTypeOf(&Config{}))
		Expect(conf.RetryAfter).To(Equal(30 * time.Second))
		Expect(conf.ID.Factory).To(Equal("nanoid"))
		Expect(conf.Auth.Methods).To(Equal([]string{"basic"}))
		Expect(conf.Pagination.MaxLimit).To(Equal(10_000))
		Expect(conf.Server.Address).To(Equal(":8888"))
		Expect(conf.Server.ShutdownTimeout).To(Equal(5 * time.Second))
		Expect(conf.EOS.Time).To(BeZero())
	})

	It("parse env", func() {
		env := MapEnv{
			"RIPOSO_SERVER_ADDRESS":      ":8889",
			"RIPOSO_PERMISSION_DEFAULTS": `{"bucket:create":[foo,bar], "bucket:read":[system.Everyone]}`,
			"RIPOSO_EOS_TIME":            "2042-12-24T19:17:13Z",
		}

		conf, err := Parse("", env)
		Expect(err).NotTo(HaveOccurred())
		Expect(conf).To(BeAssignableToTypeOf(&Config{}))
		Expect(conf.ID.Factory).To(Equal("nanoid"))
		Expect(conf.Pagination.MaxLimit).To(Equal(10_000))
		Expect(conf.Server.Address).To(Equal(":8889"))
		Expect(conf.Permission.Defaults).To(Equal(map[string][]string{
			"bucket:create": {"foo", "bar"},
			"bucket:read":   {"system.Everyone"},
		}))
		Expect(conf.EOS.Time).To(BeTemporally("==", time.Date(2042, 12, 24, 19, 17, 13, 0, time.UTC)))
	})

	It("parses files", func() {
		conf, err := Parse("testdata/config.yml", MapEnv{})
		Expect(err).NotTo(HaveOccurred())
		Expect(conf).To(BeAssignableToTypeOf(&Config{}))
		Expect(conf.ID.Factory).To(Equal("nanoid"))
		Expect(conf.Pagination.MaxLimit).To(Equal(100))
		Expect(conf.Server.Address).To(Equal(":8889"))
		Expect(conf.Server.ShutdownTimeout).To(Equal(10 * time.Second))
		Expect(conf.Temp.Dir).To(Equal("/dev/shm"))
		Expect(conf.Permission.Defaults).To(Equal(map[string][]string{
			"bucket:create": {"system.Authenticated"},
			"bucket:read":   {"bar", "foo"},
		}))
		Expect(conf.Rules).To(HaveLen(2))
		Expect(conf.EOS.Time).To(BeTemporally("==", time.Date(2042, 12, 24, 17, 29, 37, 0, time.UTC)))
	})

	It("prioritises env", func() {
		env := MapEnv{
			"RIPOSO_SERVER_ADDRESS": ":8899",
		}

		conf, err := Parse("testdata/config.yml", env)
		Expect(err).NotTo(HaveOccurred())
		Expect(conf.Server.Address).To(Equal(":8899"))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/config")
}

package cache_test

import (
	"context"
	"os"
	"testing"

	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/conn/cache/testdata"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/conn/postgres/cache"
)

var _ = Describe("Backend", func() {
	var link testdata.LikeBackend

	BeforeEach(func() {
		link.Backend = instance
	})

	testdata.BehavesLikeBackend(&link)
})

// --------------------------------------------------------------------

var instance cache.Backend

var _ = BeforeSuite(func() {
	dsn := "postgres://127.0.0.1/riposo_test?timezone=UTC"
	if val := os.Getenv("POSTGRES_DSN"); val != "" {
		dsn = val
	}

	var err error
	instance, err = Connect(context.Background(), dsn)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if instance != nil {
		Expect(instance.Close()).To(Succeed())
	}
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/conn/postgres/cache")
}

package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/conn/storage/testdata"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/conn/postgres/storage"
)

var _ = Describe("Backend", func() {
	var ctx = context.Background()
	var link testdata.LikeBackend

	BeforeEach(func() {
		instance.(reloadable).ReloadHelpers(mock.Helpers())
		link.Backend = instance
	})

	testdata.BehavesLikeBackend(&link)

	It("supports CONTAINS ANY filters", func() {
		tx, err := instance.Begin(ctx)
		Expect(err).NotTo(HaveOccurred())
		defer tx.Rollback()

		_, _, err = testdata.StdSeeds(tx)
		Expect(err).NotTo(HaveOccurred())

		Expect(testdata.FilterScope(tx, "contains_any_ary", "x")).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", "x,y,z")).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", "5,6,7")).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", "w,false")).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", `{"z":8}`)).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", "null")).To(ConsistOf("EPR.ID"))
		Expect(testdata.FilterScope(tx, "contains_any_ary", "true")).To(BeEmpty())
		Expect(testdata.FilterScope(tx, "contains_any_ary", "a,b,c")).To(BeEmpty())
		Expect(testdata.FilterScope(tx, "contains_any_ary", "[]")).To(BeEmpty())
		Expect(testdata.FilterScope(tx, "contains_any_ary", "{}")).To(BeEmpty())
	})
})

// --------------------------------------------------------------------

type reloadable interface {
	ReloadHelpers(riposo.Helpers)
}

var instance storage.Backend

var _ = BeforeSuite(func() {
	dsn := "postgres://127.0.0.1/riposo_test?timezone=UTC"
	if val := os.Getenv("POSTGRES_DSN"); val != "" {
		dsn = val
	}

	var err error
	instance, err = Connect(context.Background(), dsn, mock.Helpers())
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if instance != nil {
		Expect(instance.Close()).To(Succeed())
	}
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/conn/postgres/storage")
}

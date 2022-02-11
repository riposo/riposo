package storage_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/conn/storage/testdata"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/params"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/conn/memory/storage"
)

var _ = Describe("Backend", func() {
	var subject storage.Backend
	var link testdata.LikeBackend

	BeforeEach(func() {
		subject = New(clock.New(), mock.Helpers())
		link.Backend = subject
		link.SkipACID = true
		link.SkipFilters = []params.Operator{
			params.OperatorContains,
			params.OperatorContainsAny,
		}
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	testdata.BehavesLikeBackend(&link)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/conn/memory/storage")
}

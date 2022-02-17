package permission_test

import (
	"testing"

	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/conn/permission/testdata"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/conn/memory/permission"
)

var _ = Describe("Backend", func() {
	var subject permission.Backend
	var link testdata.LikeBackend

	BeforeEach(func() {
		subject = New()
		link.Backend = subject
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	Describe("common", func() {
		testdata.BehavesLikeBackend(&link)
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/conn/memory/permission")
}

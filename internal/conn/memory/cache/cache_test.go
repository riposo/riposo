package cache_test

import (
	"testing"

	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/conn/cache/testdata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/riposo/riposo/internal/conn/memory/cache"
)

var _ = Describe("Backend", func() {
	var subject cache.Backend
	var link testdata.LikeBackend

	BeforeEach(func() {
		subject = New()
		link.Backend = subject
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	testdata.BehavesLikeBackend(&link)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/conn/memory/cache")
}

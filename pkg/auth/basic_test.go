package auth_test

import (
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Basic", func() {
	var subject auth.Method
	var txn *api.Txn

	BeforeEach(func() {
		txn = mock.Txn()

		pass, err := txn.Helpers.SlowHash("s3cret")
		Expect(err).NotTo(HaveOccurred())
		Expect(txn.Store.Create("/accounts/*", &schema.Object{
			ID:    "testuser",
			Extra: []byte(`{"password":"` + pass + `"}`),
		})).To(Succeed())

		subject = auth.Basic()
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(txn.Abort()).To(Succeed())
	})

	It("authenticates", func() {
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("testuser", "s3cret")
		Expect(subject.Authenticate(req)).To(Equal(&api.User{ID: "account:testuser"}))
	})

	It("does not authenticate without authorization", func() {
		req := mock.Request(txn, "GET", "/", nil)

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(auth.ErrUnauthenticated))
		Expect(err).To(MatchError(`no basic auth credentials`))
	})

	It("does not authenticate unknown users", func() {
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("unknown", "s3cret")

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(auth.ErrUnauthenticated))
		Expect(err).To(MatchError(`unknown user account`))
	})

	It("rejects bad credentials", func() {
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("testuser", "wrongpass")

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(auth.ErrUnauthenticated))
		Expect(err).To(MatchError(`invalid password`))
	})
})

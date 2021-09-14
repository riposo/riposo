package auth_test

import (
	"net/http"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/mock"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/auth"
)

var _ = Describe("OneOf", func() {
	var subject Method
	var txn *api.Txn

	BeforeEach(func() {
		txn = mock.Txn()
		subject = OneOf(mockAuth("alice"), mockAuth("bob"))
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(txn.Abort()).To(Succeed())
	})

	It("authenticates", func() {
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("alice", "")
		Expect(subject.Authenticate(req)).To(Equal(&api.User{ID: "mock:alice"}))

		req.SetBasicAuth("bob", "")
		Expect(subject.Authenticate(req)).To(Equal(&api.User{ID: "mock:bob"}))
	})

	It("does not authenticate without authorization", func() {
		req := mock.Request(txn, "GET", "/", nil)

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(ErrUnauthenticated))
		Expect(err).To(MatchError(`no credentials`))
	})

	It("does not authenticate unknown users", func() {
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("claire", "")

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(ErrUnauthenticated))
		Expect(err).To(MatchError(`unknown user`))
	})

	It("fails when empty", func() {
		subject = OneOf()
		req := mock.Request(txn, "GET", "/", nil)
		req.SetBasicAuth("alice", "")

		_, err := subject.Authenticate(req)
		Expect(err).To(MatchError(ErrUnauthenticated))
		Expect(err).To(MatchError(`no authentication methods enabled`))
	})
})

type mockAuth string

func (m mockAuth) Authenticate(r *http.Request) (*api.User, error) {
	user, _, ok := r.BasicAuth()
	if !ok {
		return nil, Errorf("no credentials")
	} else if user != string(m) {
		return nil, Errorf("unknown user")
	}
	return &api.User{ID: "mock:" + user}, nil
}

func (mockAuth) Close() error {
	return nil
}

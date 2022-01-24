package api_test

import (
	"context"

	memory "github.com/riposo/riposo/internal/conn/memory/permission"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/api"
)

var _ = Describe("Authz", func() {
	var subject Authz
	var backend permission.Backend
	var tx permission.Transaction
	var ctx = context.Background()

	ACE := func(perm string, path riposo.Path) permission.ACE {
		return permission.ACE{Perm: perm, Path: path}
	}

	BeforeEach(func() {
		backend = memory.New()
		subject = Authz{
			"static:create": {"system.Authenticated"},
		}

		var err error
		tx, err = backend.Begin(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(tx.Rollback()).To(Succeed())
		Expect(backend.Close()).To(Succeed())
	})

	It("verifies", func() {
		Expect(tx.AddACEPrincipal("alice", ACE("write", "/bucket/bat"))).To(Succeed())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/accounts/ant"),
		})).To(BeFalse())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("read", "/bucket/bat"),
		})).To(BeFalse())
		Expect(subject.Verify(tx, []string{"bob"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeFalse())
	})

	It("supports static principals", func() {
		Expect(subject.Verify(tx, []string{"system.Authenticated"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeFalse())
		Expect(subject.Verify(tx, []string{"system.Authenticated"}, []permission.ACE{
			ACE("static:create", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"system.Authenticated"}, []permission.ACE{
			ACE("static:create", ""),
		})).To(BeTrue())
	})

	It("requires at least one principal to match", func() {
		Expect(tx.AddACEPrincipal("alice", ACE("write", "/bucket/bat"))).To(Succeed())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"alice", "bob"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"bob"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeFalse())
		Expect(subject.Verify(tx, []string{"bob", "system.Authenticated"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
			ACE("static:create", ""),
		})).To(BeTrue())
	})

	It("requires at least one entity to match", func() {
		Expect(tx.AddACEPrincipal("alice", ACE("write", "/bucket/bat"))).To(Succeed())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/accounts/ant"),
			ACE("write", "/bucket/bat"),
		})).To(BeTrue())
		Expect(subject.Verify(tx, []string{"alice"}, []permission.ACE{
			ACE("write", "/accounts/ant"),
			ACE("read", "/bucket/bat"),
		})).To(BeFalse())
	})

	It("refuses empty sets", func() {
		Expect(tx.AddACEPrincipal("alice", ACE("write", "/bucket/bat"))).To(Succeed())
		Expect(subject.Verify(tx, nil, []permission.ACE{ACE("write", "/bucket/bat")})).To(BeFalse())
		Expect(subject.Verify(tx, []string{"alice"}, nil)).To(BeFalse())
	})
})

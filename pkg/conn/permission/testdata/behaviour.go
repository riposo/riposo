package testdata

import (
	"context"
	"sort"

	Ψ "github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

type testableTx interface {
	NumEntries() (int64, error)
}

// LikeBackend test link.
type LikeBackend struct {
	permission.Backend
}

// BehavesLikeBackend contains common Backend behaviour
func BehavesLikeBackend(link *LikeBackend) {
	var subject permission.Backend
	var tx permission.Transaction
	var ctx = context.Background()

	ACE := func(perm string, path riposo.Path) permission.ACE {
		return permission.ACE{Perm: perm, Path: path}
	}

	Ψ.BeforeEach(func() {
		subject = link.Backend

		var err error
		tx, err = subject.Begin(ctx)
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
	})

	Ψ.AfterEach(func() {
		Ω.Expect(tx.Flush()).To(Ω.Succeed())
		Ω.Expect(tx.Rollback()).To(Ω.Succeed())
	})

	Ψ.It("connects and migrates", func() {
		Ω.Expect(subject).NotTo(Ω.BeNil())
		Ω.Expect(tx).NotTo(Ω.BeNil())
	})

	Ψ.It("pings", func() {
		Ω.Expect(subject.Ping(ctx)).To(Ω.Succeed())
	})

	Ψ.It("flushes", func() {
		Ω.Expect(tx.AddACEPrincipal("alice", ACE("read", "/accounts/ant"))).To(Ω.Succeed())
		Ω.Expect(tx.(testableTx).NumEntries()).To(Ω.BeNumerically("==", 1))
		Ω.Expect(tx.Flush()).To(Ω.Succeed())
		Ω.Expect(tx.(testableTx).NumEntries()).To(Ω.BeNumerically("==", 0))
	})

	Ψ.It("manages user principals", func() {
		Ω.Expect(tx.AddUserPrincipal("b", []string{"alice", "bob"})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("a", []string{"alice", "bob"})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("b", []string{"bob"})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("c", []string{"alice"})).To(Ω.Succeed())
		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"a", "b", "c",
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals("bob")).To(Ω.ConsistOf(
			"a", "b",
			"bob", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals("claire")).To(Ω.ConsistOf(
			"claire", riposo.Authenticated, riposo.Everyone,
		))

		Ω.Expect(tx.RemoveUserPrincipal("a", []string{"alice", "claire"})).To(Ω.Succeed())
		Ω.Expect(tx.RemoveUserPrincipal("x", []string{"alice"})).To(Ω.Succeed())
		Ω.Expect(tx.PurgeUserPrincipals(nil)).To(Ω.Succeed()) // deletes nothing
		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"b", "c",
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals("claire")).To(Ω.ConsistOf(
			"claire", riposo.Authenticated, riposo.Everyone,
		))

		Ω.Expect(tx.PurgeUserPrincipals([]string{"b", "x"})).To(Ω.Succeed())
		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"c",
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals("bob")).To(Ω.ConsistOf(
			"a",
			"bob", riposo.Authenticated, riposo.Everyone,
		))

		Ω.Expect(tx.PurgeUserPrincipals([]string{"a", "c"})).To(Ω.Succeed())
		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals("bob")).To(Ω.ConsistOf(
			"bob", riposo.Authenticated, riposo.Everyone,
		))
	})

	Ψ.It("always appends key user principals", func() {
		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals(riposo.Authenticated)).To(Ω.ConsistOf(
			riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals(riposo.Everyone)).To(Ω.ConsistOf(
			riposo.Everyone,
		))

		Ω.Expect(tx.AddUserPrincipal("a", []string{"alice"})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("x", []string{"alice"})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("x", []string{riposo.Authenticated})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("y", []string{riposo.Authenticated})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("x", []string{riposo.Everyone})).To(Ω.Succeed())
		Ω.Expect(tx.AddUserPrincipal("z", []string{riposo.Everyone})).To(Ω.Succeed())

		Ω.Expect(tx.GetUserPrincipals("alice")).To(Ω.ConsistOf(
			"a", "x", "y", "z",
			"alice", riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals(riposo.Authenticated)).To(Ω.ConsistOf(
			"x", "y", "z",
			riposo.Authenticated, riposo.Everyone,
		))
		Ω.Expect(tx.GetUserPrincipals(riposo.Everyone)).To(Ω.ConsistOf(
			"x", "z",
			riposo.Everyone,
		))
	})

	Ψ.It("manages ACE principals", func() {
		Ω.Expect(tx.AddACEPrincipal("x", ACE("read", "/accounts/ant"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("y", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("y", ACE("write", "/buckets/bat"))).To(Ω.Succeed())

		Ω.Expect(tx.GetACEPrincipals(ACE("read", "/accounts/ant"))).To(Ω.ConsistOf("x"))
		Ω.Expect(tx.GetACEPrincipals(ACE("write", "/accounts/ant"))).To(Ω.BeEmpty())
		Ω.Expect(tx.GetACEPrincipals(ACE("write", "/buckets/bat"))).To(Ω.ConsistOf("x", "y"))
		Ω.Expect(tx.GetACEPrincipals(ACE("read", "/buckets/bat"))).To(Ω.BeEmpty())
		Ω.Expect(tx.GetACEPrincipals(ACE("write", "/collections/cod"))).To(Ω.BeEmpty())

		Ω.Expect(tx.RemoveACEPrincipal("y", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.RemoveACEPrincipal("z", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.RemoveACEPrincipal("x", ACE("write", "/collections/cod"))).To(Ω.Succeed())
		Ω.Expect(tx.RemoveACEPrincipal("y", ACE("read", "/unknown/uzo"))).To(Ω.Succeed())
		Ω.Expect(tx.GetACEPrincipals(ACE("write", "/buckets/bat"))).To(Ω.ConsistOf("x"))
	})

	Ψ.It("retrieves bulk principals", func() {
		Ω.Expect(tx.AddACEPrincipal("x", ACE("read", "/accounts/ant"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("z", ACE("write", "/accounts/ant"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("y", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("z", ACE("create", "/buckets/bat"))).To(Ω.Succeed())

		Ω.Expect(tx.GetAllACEPrincipals([]permission.ACE{
			ACE("read", "/accounts/ant"),
			ACE("write", "/buckets/bat"),
			ACE("read", "/buckets/bat"),
			ACE("write", "/buckets/bat"),
			ACE("create", "/unknown/uzo"),
		})).To(Ω.ConsistOf("x", "y"))

		Ω.Expect(tx.GetAllACEPrincipals(nil)).To(Ω.BeNil())
	})

	Ψ.It("creates permissions", func() {
		Ω.Expect(tx.GetPermissions("/accounts/ant")).To(MatchPermissions(nil))

		Ω.Expect(tx.CreatePermissions("/accounts/ant", schema.PermissionSet{
			"read":       []string{"x"},
			"sub:create": []string{"y"},
		})).To(Ω.Succeed())
		Ω.Expect(tx.CreatePermissions("/buckets/bat", schema.PermissionSet{
			"read":  []string{"z"},
			"write": []string{"x", "y"},
		})).To(Ω.Succeed())

		Ω.Expect(tx.GetPermissions("/accounts/ant")).To(MatchPermissions(schema.PermissionSet{
			"read":       []string{"x"},
			"sub:create": []string{"y"},
		}))
		Ω.Expect(tx.GetPermissions("/buckets/bat")).To(MatchPermissions(schema.PermissionSet{
			"read":  []string{"z"},
			"write": []string{"x", "y"},
		}))
		Ω.Expect(tx.GetPermissions("/collections/cod")).To(MatchPermissions(nil))

		// ignore duplicates, support partials
		Ω.Expect(tx.CreatePermissions("/accounts/ant", schema.PermissionSet{
			"read": []string{"x", "y"},
		})).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/accounts/ant")).To(MatchPermissions(schema.PermissionSet{
			"read":       []string{"x", "y"},
			"sub:create": []string{"y"},
		}))
	})

	Ψ.It("merges permissions", func() {
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("y", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("y", ACE("read", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("z", ACE("delete", "/buckets/bat"))).To(Ω.Succeed())

		Ω.Expect(tx.GetPermissions("/buckets/bat")).To(MatchPermissions(schema.PermissionSet{
			"write":  {"x", "y"},
			"read":   {"y"},
			"delete": {"z"},
		}))
		Ω.Expect(tx.MergePermissions("/buckets/bat", nil)).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/buckets/bat")).To(MatchPermissions(schema.PermissionSet{
			"write":  {"x", "y"},
			"read":   {"y"},
			"delete": {"z"},
		}))

		Ω.Expect(tx.MergePermissions("/buckets/bat", schema.PermissionSet{
			"write":  {"x", "z"},
			"delete": {},
			"new":    {"z"},
		})).To(Ω.Succeed())

		Ω.Expect(tx.GetPermissions("/buckets/bat")).To(MatchPermissions(schema.PermissionSet{
			"write": {"x", "z"},
			"read":  {"y"},
			"new":   {"z"},
		}))
	})

	Ψ.It("resets permissions locally", func() {
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/bat"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/bat/collections/boom"))).To(Ω.Succeed())

		Ω.Expect(tx.MergePermissions("/buckets/bat", schema.PermissionSet{"write": {}})).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/buckets/bat")).To(MatchPermissions(nil))
		Ω.Expect(tx.GetPermissions("/buckets/bat/collections/boom")).To(MatchPermissions(schema.PermissionSet{"write": {"x"}}))
	})

	Ψ.It("deletes permissions (recursively)", func() {
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/a"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/b"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/c"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/c/collections/x"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/cc"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("x", ACE("write", "/buckets/z/collections/x"))).To(Ω.Succeed())

		Ω.Expect(tx.DeletePermissions([]riposo.Path{"/buckets/a", "/buckets/b", "/buckets/z"})).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/buckets/a")).To(MatchPermissions(nil))
		Ω.Expect(tx.GetPermissions("/buckets/b")).To(MatchPermissions(nil))
		Ω.Expect(tx.GetPermissions("/buckets/c")).To(Ω.HaveLen(1))
		Ω.Expect(tx.GetPermissions("/buckets/z/collections/x")).To(MatchPermissions(nil))

		Ω.Expect(tx.DeletePermissions(nil)).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/buckets/c")).To(Ω.HaveLen(1))
		Ω.Expect(tx.GetPermissions("/buckets/c/collections/x")).To(Ω.HaveLen(1))
		Ω.Expect(tx.GetPermissions("/buckets/cc")).To(Ω.HaveLen(1))

		Ω.Expect(tx.DeletePermissions([]riposo.Path{"/buckets/c"})).To(Ω.Succeed())
		Ω.Expect(tx.GetPermissions("/buckets/c")).To(MatchPermissions(nil))
		Ω.Expect(tx.GetPermissions("/buckets/c/collections/x")).To(MatchPermissions(nil))
		Ω.Expect(tx.GetPermissions("/buckets/cc")).To(Ω.HaveLen(1))
	})

	Ψ.It("gets accessible paths", func() {
		Ω.Expect(tx.AddACEPrincipal("user", ACE("write", "/buckets/a"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("group", ACE("sub:create", "/buckets/a"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("user", ACE("read", "/buckets/b"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("other", ACE("read", "/buckets/b"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("other", ACE("write", "/buckets/c"))).To(Ω.Succeed())
		Ω.Expect(tx.AddACEPrincipal("group", ACE("read", "/buckets/c/collections/x"))).To(Ω.Succeed())

		pcps := []string{"user", "group"}

		// no entities
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, nil)).To(Ω.BeEmpty())

		// no principals
		Ω.Expect(tx.GetAccessiblePaths(nil, nil, []permission.ACE{ACE("read", "")})).To(Ω.BeEmpty())

		// no matches within path
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, []permission.ACE{ACE("read", "")})).To(Ω.BeEmpty())

		// wildcards match only immediate children
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, []permission.ACE{ACE("read", "*")})).To(Ω.BeEmpty())

		// wildcard match
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, []permission.ACE{
			ACE("read", "/buckets/*"),
			ACE("write", "/buckets/*"),
		})).To(Ω.ConsistOf([]riposo.Path{"/buckets/a", "/buckets/b"}))
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, []permission.ACE{
			ACE("write", "/buckets/*"),
		})).To(Ω.ConsistOf([]riposo.Path{"/buckets/a"}))

		// mixed match
		Ω.Expect(tx.GetAccessiblePaths(nil, pcps, []permission.ACE{
			ACE("sub:create", "/buckets/*"),
			ACE("read", "/buckets/c/collections/x"),
		})).To(Ω.ConsistOf([]riposo.Path{"/buckets/a", "/buckets/c/collections/x"}))
	})
}

// MatchPermissions matcher.
func MatchPermissions(exp schema.PermissionSet) types.GomegaMatcher {
	transform := func(set schema.PermissionSet) []string {
		res := make([]string, 0, len(set))
		for k, vv := range set {
			for _, v := range vv {
				res = append(res, k+":"+v)
			}
		}
		return res
	}

	return Ω.WithTransform(func(act schema.PermissionSet) []string {
		return transform(act)
	}, Ω.ConsistOf(transform(exp)))
}

// MatchPaths matcher.
func MatchPaths(exp map[riposo.Path][]string) types.GomegaMatcher {
	transform := func(paths map[riposo.Path][]string) []string {
		res := make([]string, 0, len(paths))
		for k, vv := range paths {
			for _, v := range vv {
				res = append(res, v+"@"+k.String())
			}
		}
		sort.Strings(res)
		return res
	}

	return Ω.WithTransform(func(actual map[riposo.Path][]string) []string {
		return transform(actual)
	}, Ω.Equal(transform(exp)))
}

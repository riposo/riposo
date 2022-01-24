package testdata

import (
	"context"
	"strings"
	"time"

	Ψ "github.com/bsm/ginkgo/v2"
	Ω "github.com/bsm/gomega"
	"github.com/riposo/riposo/pkg/conn/cache"
)

type testableTx interface {
	NumEntries() (int64, error)
}

// LikeBackend test link.
type LikeBackend struct {
	cache.Backend
}

// BehavesLikeBackend contains common Backend behaviour
func BehavesLikeBackend(link *LikeBackend) {
	var subject cache.Backend
	var tx cache.Transaction
	var ctx = context.Background()

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
	})

	Ψ.It("pings", func() {
		Ω.Expect(subject.Ping(ctx)).To(Ω.Succeed())
	})

	Ψ.It("flushes", func() {
		Ω.Expect(tx.Set("key", []byte("val"), time.Now().Add(time.Hour))).To(Ω.Succeed())
		Ω.Expect(tx.(testableTx).NumEntries()).To(Ω.BeNumerically("==", 1))
		Ω.Expect(tx.Flush()).To(Ω.Succeed())
		Ω.Expect(tx.(testableTx).NumEntries()).To(Ω.BeNumerically("==", 0))
	})

	Ψ.It("gets/sets", func() {
		Ω.Expect(tx.Set("key", []byte("val"), time.Now().Add(time.Hour))).To(Ω.Succeed())
		Ω.Expect(tx.Get("key")).To(Ω.Equal([]byte("val")))

		_, err := tx.Get("unknown")
		Ω.Expect(err).To(Ω.MatchError(cache.ErrNotFound))
	})

	Ψ.It("allows blank values", func() {
		Ω.Expect(tx.Set("key", nil, time.Now().Add(time.Hour))).To(Ω.Succeed())
		Ω.Expect(tx.Get("key")).To(Ω.BeEmpty())
	})

	Ψ.It("limits key length values", func() {
		val := []byte("val")
		exp := time.Now().Add(time.Hour)
		Ω.Expect(tx.Set("", val, exp)).To(Ω.MatchError("key is invalid"))
		Ω.Expect(tx.Set(strings.Repeat("本", 257), val, exp)).To(Ω.MatchError("key is invalid"))
		Ω.Expect(tx.Set(strings.Repeat("本", 256), val, exp)).To(Ω.Succeed())
	})

	Ψ.It("fails when retrieving expired", func() {
		Ω.Expect(tx.Set("key", []byte("val"), time.Now().Add(-time.Second))).To(Ω.Succeed())
		_, err := tx.Get("key")
		Ω.Expect(err).To(Ω.MatchError(cache.ErrNotFound))
	})

	Ψ.It("deletes", func() {
		Ω.Expect(tx.Set("key", []byte("val"), time.Now().Add(time.Hour))).To(Ω.Succeed())
		Ω.Expect(tx.Del("key")).To(Ω.Succeed())
		Ω.Expect(tx.Del("key")).To(Ω.MatchError(cache.ErrNotFound))
	})

	Ψ.It("fails when deleting expired", func() {
		Ω.Expect(tx.Set("key", []byte("val"), time.Now().Add(-time.Second))).To(Ω.Succeed())
		Ω.Expect(tx.Del("key")).To(Ω.MatchError(cache.ErrNotFound))
	})
}

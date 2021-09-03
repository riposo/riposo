package testdata

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"

	Ψ "github.com/bsm/ginkgo"
	Ω "github.com/bsm/gomega"
	"github.com/bsm/gomega/types"
)

type testableTx interface {
	NumEntries() (int64, error)
}

// LikeBackend test link.
type LikeBackend struct {
	storage.Backend
	SkipFilters []params.Operator
}

// BehavesLikeBackend contains common store behaviour
func BehavesLikeBackend(link *LikeBackend) {
	var subject storage.Backend
	var tx storage.Transaction
	var ctx = context.Background()

	BeRecent := func() types.GomegaMatcher {
		return Ω.BeNumerically("~", riposo.CurrentEpoch(), 1000)
	}
	mustDelete := func(path riposo.Path) {
		_, err := tx.Delete(path)
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
	}
	NumEntries := func() (int, error) {
		n, err := tx.(testableTx).NumEntries()
		return int(n), err
	}

	Ψ.BeforeEach(func() {
		subject = link.Backend

		var err error
		tx, err = subject.Begin(ctx)
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
	})

	Ψ.AfterEach(func() {
		Ω.Expect(tx.Rollback()).To(Ω.Or(Ω.Succeed(), Ω.MatchError(sql.ErrTxDone)))
	})

	Ψ.It("connects and migrates", func() {
		Ω.Expect(subject).NotTo(Ω.BeNil())
	})

	Ψ.It("pings", func() {
		Ω.Expect(subject.Ping(ctx)).To(Ω.Succeed())
	})

	Ψ.It("flushes", func() {
		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(1))

		Ω.Expect(tx.Flush()).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(0))
		Ω.Expect(tx.Flush()).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(0))
	})

	Ψ.It("gets mod-times", func() {
		// only accept node paths
		_, err := tx.ModTime("/objects/foo")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// should be 0 when no nodes
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.BeNumerically("==", 0))

		// create an object
		o1 := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())

		// should be set and static
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.Equal(o1.ModTime))
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.Equal(o1.ModTime))

		// delete object
		o2, err := tx.Delete("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(o2.ModTime).To(Ω.BeNumerically(">", o1.ModTime))

		// should update
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.BeNumerically("==", o2.ModTime))
	})

	Ψ.It("checks if objects exist", func() {
		o1 := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())

		Ω.Expect(tx.Exists("/objects/EPR.ID")).To(Ω.BeTrue())
		Ω.Expect(tx.Exists("/objects/missing")).To(Ω.BeFalse())

		// reject node paths
		_, err := tx.Get("/objects/*")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))
	})

	Ψ.It("gets objects", func() {
		o1 := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())

		o2, err := tx.Get("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(o2.ID).To(Ω.Equal(o1.ID))
		Ω.Expect(o2.ModTime).To(Ω.Equal(o1.ModTime))
		Ω.Expect(o2.Extra).To(Ω.MatchJSON(`{}`))
		Ω.Expect(o2).NotTo(Ω.BeIdenticalTo(o1))

		// reject node paths
		_, err = tx.Get("/objects/*")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// handle not-found
		_, err = tx.Get("/objects/missing-id")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrNotFound))
	})

	Ψ.It("gets objects for update", func() {
		o1 := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())

		h, err := tx.GetForUpdate("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(h.Object().ID).To(Ω.Equal(o1.ID))
		Ω.Expect(h.Object()).NotTo(Ω.BeIdenticalTo(o1))

		// reject node paths
		_, err = tx.GetForUpdate("/objects/*")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// handle not-found
		_, err = tx.GetForUpdate("/objects/missing-id")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrNotFound))
	})

	Ψ.It("creates objects", func() {
		// only accept node paths
		Ω.Expect(tx.Create("/objects/foo", nil)).To(Ω.MatchError(storage.ErrInvalidPath))

		o1 := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())
		Ω.Expect(o1.ID).To(Ω.Equal("EPR.ID"))
		Ω.Expect(o1.ModTime).To(BeRecent())
		Ω.Expect(o1.Extra).To(Ω.MatchJSON(`{}`))
		Ω.Expect(NumEntries()).To(Ω.Equal(1))

		// maintain the node's epoch
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.Equal(o1.ModTime))

		// duplicate
		o2 := &schema.Object{ID: "EPR.ID"}
		Ω.Expect(tx.Create("/objects/*", o2)).To(Ω.Equal(storage.ErrObjectExists))

		// previously deleted
		mustDelete("/objects/EPR.ID")
		Ω.Expect(NumEntries()).To(Ω.Equal(0))

		Ω.Expect(tx.Create("/objects/*", o2)).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(1))
		Ω.Expect(o2.ModTime).To(Ω.BeNumerically(">", o1.ModTime))
	})

	Ψ.It("creates in parallel", func() {
		wg := new(sync.WaitGroup)
		for t := 0; t < 5; t++ {
			wg.Add(1)

			go func() {
				defer Ψ.GinkgoRecover()
				defer wg.Done()

				tx2, err := subject.Begin(ctx)
				Ω.Expect(err).NotTo(Ω.HaveOccurred())
				defer tx2.Rollback()

				for i := 0; i < 20; i++ {
					Ω.Expect(tx2.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
				}
				Ω.Expect(tx2.Commit()).To(Ω.Succeed())
			}()
		}
		wg.Wait()

		tx2, err := subject.Begin(ctx)
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		defer tx2.Rollback()

		objs, err := tx2.ListAll(nil, "/objects/*", storage.ListOptions{})
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(objs).To(Ω.HaveLen(100))

		Ω.Expect(tx2.Flush()).To(Ω.Succeed())
		Ω.Expect(tx2.Commit()).To(Ω.Succeed())

		epochs := make(map[riposo.Epoch]int)
		for _, obj := range objs {
			epochs[obj.ModTime]++
		}
		Ω.Expect(epochs).To(Ω.HaveLen(100))
	})

	Ψ.It("updates objects", func() {
		obj := &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", obj)).To(Ω.Succeed())

		// update
		h1, err := tx.GetForUpdate("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(tx.Update(h1)).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(1))

		o1, err := tx.Get("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(o1.ModTime).To(Ω.BeNumerically(">", obj.ModTime))

		// always increment timestamps
		h2, err := tx.GetForUpdate("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(tx.Update(h2)).To(Ω.Succeed())
		Ω.Expect(NumEntries()).To(Ω.Equal(1))

		o2, err := tx.Get("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(o2.ModTime).To(Ω.BeNumerically(">", o1.ModTime))
	})

	Ψ.It("deletes individual objects", func() {
		// seed!
		o1 := &schema.Object{Extra: []byte(`{"meta":true}`)}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())
		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/objects/*", &schema.Object{ID: "EPR.IDX"})).To(Ω.Succeed())

		n1 := &schema.Object{}
		Ω.Expect(tx.Create(riposo.Path("/objects/EPR.ID/nested/*"), n1)).To(Ω.Succeed())
		Ω.Expect(tx.Create(riposo.Path("/objects/EPR.ID/nested/*"), &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create(riposo.Path("/objects/OTHER/nested/*"), &schema.Object{})).To(Ω.Succeed())

		// confirm
		time.Sleep(time.Millisecond)
		Ω.Expect(NumEntries()).To(Ω.Equal(6))

		// handle not-found
		_, err := tx.Delete("/objects/unknown")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrNotFound))

		// not delete nested if exact not found
		_, err = tx.Delete("/objects/OTHER")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrNotFound))
		Ω.Expect(NumEntries()).To(Ω.Equal(6))

		// reject invalid paths
		_, err = tx.Delete("/objects/*")
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// delete o1 (+ 2 nested)
		d1, err := tx.Delete("/objects/EPR.ID")
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(d1.ModTime).To(Ω.BeNumerically(">", o1.ModTime))
		Ω.Expect(d1.Deleted).To(Ω.BeTrue())
		Ω.Expect(d1.Extra).To(Ω.MatchJSON(`{"meta":true}`))
		Ω.Expect(NumEntries()).To(Ω.Equal(3))

		// maintain the node's epoch
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.Equal(d1.ModTime))

		// maintain nested node's epoch
		Ω.Expect(tx.ModTime("/objects/EPR.ID/nested/*")).To(Ω.BeNumerically(">=", d1.ModTime))

		// retain deleted records
		Ω.Expect(tx.ListAll(nil, "/objects/*", storage.ListOptions{
			Include: storage.IncludeAll,
		})).To(Ω.And(
			Ω.HaveLen(3),
			Ω.ContainElement(d1),
		))
	})

	Ψ.It("counts matching objects", func() {
		// seed!
		Ω.Expect(tx.Create("/parents/a/objects/*", &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/parents/a/objects/*", &schema.Object{Extra: []byte(`{"x": 11, "y": 33}`)})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/parents/b/objects/*", &schema.Object{Extra: []byte(`{"x": "v", "y": 22}`)})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/others/*", &schema.Object{})).To(Ω.Succeed())

		// only accept node paths
		_, err := tx.CountAll("/objects/foo", storage.CountOptions{})
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// exact
		Ω.Expect(tx.CountAll("/parents/a/objects/*", storage.CountOptions{})).To(Ω.Equal(int64(2)))
		Ω.Expect(tx.CountAll("/objects/*", storage.CountOptions{})).To(Ω.Equal(int64(0)))
		Ω.Expect(tx.CountAll("/parents/x/objects/*", storage.CountOptions{})).To(Ω.Equal(int64(0)))
		Ω.Expect(tx.CountAll("/parents/a/unknowns/*", storage.CountOptions{})).To(Ω.Equal(int64(0)))

		// conditions
		Ω.Expect(tx.CountAll("/parents/a/objects/*", storage.CountOptions{Condition: params.Condition{
			params.ParseFilter("has_x", "true"),
		}})).To(Ω.Equal(int64(1)))
	})

	Ψ.It("lists matching objects", func() {
		// seed!
		Ω.Expect(tx.Create("/parents/a/objects/*", &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/parents/a/objects/*", &schema.Object{Extra: []byte(`{"x": 11, "y": 33}`)})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/parents/b/objects/*", &schema.Object{Extra: []byte(`{"x": "v", "y": 22}`)})).To(Ω.Succeed())
		Ω.Expect(tx.Create("/others/*", &schema.Object{})).To(Ω.Succeed())

		// only accept node paths
		_, err := tx.ListAll(nil, "/objects/foo", storage.ListOptions{})
		Ω.Expect(err).To(Ω.MatchError(storage.ErrInvalidPath))

		// exact
		Ω.Expect(tx.ListAll(nil, "/parents/a/objects/*", storage.ListOptions{})).To(Ω.HaveLen(2))
		Ω.Expect(tx.ListAll(nil, "/objects/*", storage.ListOptions{})).To(Ω.BeEmpty())
		Ω.Expect(tx.ListAll(nil, "/parents/x/objects/*", storage.ListOptions{})).To(Ω.BeEmpty())
		Ω.Expect(tx.ListAll(nil, "/parents/a/unknowns/*", storage.ListOptions{})).To(Ω.BeEmpty())

		// limit
		Ω.Expect(tx.ListAll(nil, "/parents/a/objects/*", storage.ListOptions{Limit: 1})).To(Ω.HaveLen(1))

		// conditions
		Ω.Expect(tx.ListAll(nil, "/parents/a/objects/*", storage.ListOptions{Condition: params.Condition{
			params.ParseFilter("has_x", "true"),
		}})).To(Ω.HaveLen(1))
	})

	Ψ.It("deletes multiple objects", func() {
		// seed!
		o1, o2 := &schema.Object{Extra: []byte(`{"meta": true}`)}, &schema.Object{}
		Ω.Expect(tx.Create("/objects/*", o1)).To(Ω.Succeed())                            // EPR.ID
		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())              // ITR.ID
		Ω.Expect(tx.Create("/objects/*", o2)).To(Ω.Succeed())                            // MXR.ID
		Ω.Expect(tx.Create("/others/*", &schema.Object{})).To(Ω.Succeed())               // Q3R.ID
		Ω.Expect(tx.Create("/objects/*", &schema.Object{ID: "EPR.IDX"})).To(Ω.Succeed()) // EPR.IDX
		Ω.Expect(tx.Create(riposo.Path("/objects/MXR.ID/nested/*"), &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create(riposo.Path("/objects/MXR.ID/nested/*"), &schema.Object{})).To(Ω.Succeed())
		Ω.Expect(tx.Create(riposo.Path("/objects/OTHER/nested/*"), &schema.Object{})).To(Ω.Succeed())

		// confirm
		time.Sleep(time.Millisecond)
		Ω.Expect(NumEntries()).To(Ω.Equal(8))

		// ignore missing
		Ω.Expect(tx.DeleteAll([]riposo.Path{"/objects/foo"})).To(Ω.Equal(riposo.Epoch(0)))
		Ω.Expect(tx.DeleteAll([]riposo.Path{"/objects/*"})).To(Ω.Equal(riposo.Epoch(0)))
		Ω.Expect(NumEntries()).To(Ω.Equal(8))

		// may delete only nested
		Ω.Expect(tx.DeleteAll([]riposo.Path{"/objects/OTHER"})).To(Ω.Equal(riposo.Epoch(0)))
		Ω.Expect(NumEntries()).To(Ω.Equal(7))

		// delete ITR.ID + MXR.ID (+ 2 nested MXR.ID)
		modTime1, err := tx.DeleteAll([]riposo.Path{
			"/objects/ITR.ID", // deletes ITR.ID
			"/objects/MISSING",
			"/objects/MXR.ID", // deletes MXR.ID + 2 nested
			"/objects/Q3R.ID",
		})
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(modTime1).To(Ω.BeNumerically(">", o2.ModTime))
		Ω.Expect(NumEntries()).To(Ω.Equal(3))

		// delete EPR.ID
		modTime2, err := tx.DeleteAll([]riposo.Path{"/objects/EPR.ID"})
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		Ω.Expect(modTime2).To(Ω.BeNumerically(">", modTime1))
		Ω.Expect(NumEntries()).To(Ω.Equal(2))

		// maintain the node's epoch
		Ω.Expect(tx.ModTime("/objects/*")).To(Ω.Equal(modTime2))

		// retain deleted records
		Ω.Expect(tx.ListAll(nil, "/objects/*", storage.ListOptions{
			Include: storage.IncludeAll,
		})).To(Ω.And(
			Ω.HaveLen(4),
			Ω.ContainElement(&schema.Object{
				ID:      "EPR.ID",
				ModTime: modTime2,
				Deleted: true,
				Extra:   []byte(`{"meta": true}`),
			}),
		))
	})

	Ψ.It("purges deleted", func() {
		// seed & delete
		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
		mustDelete("/objects/EPR.ID")

		time.Sleep(3 * time.Millisecond)
		t1 := riposo.CurrentEpoch()

		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
		mustDelete("/objects/ITR.ID")
		Ω.Expect(tx.Create("/objects/*", &schema.Object{})).To(Ω.Succeed())
		mustDelete("/objects/MXR.ID")

		includeAll := storage.ListOptions{Include: storage.IncludeAll}
		Ω.Expect(ListScope(tx, includeAll)).To(Ω.HaveLen(3))

		// purge everything older than t1
		Ω.Expect(tx.Purge(t1)).To(Ω.Equal(int64(1)))
		Ω.Expect(ListScope(tx, includeAll)).To(Ω.HaveLen(2))

		// purge everything else
		Ω.Expect(tx.Purge(0)).To(Ω.Equal(int64(2)))
		Ω.Expect(NumEntries()).To(Ω.Equal(0))
	})

	Ψ.Describe("listing", func() {
		var o1, o2 *schema.Object

		etoa := func(epoch riposo.Epoch) string {
			return strconv.FormatInt(int64(epoch), 10)
		}

		Ψ.BeforeEach(func() {
			var err error
			o1, o2, err = StdSeeds(tx)
			Ω.Expect(err).NotTo(Ω.HaveOccurred())
		})

		Ψ.Describe("inclusion", func() {
			Ψ.BeforeEach(func() {
				mustDelete(riposo.Path("/objects/" + o2.ID))
			})

			Ψ.It("includes only live by default", func() {
				Ω.Expect(ListScope(tx, storage.ListOptions{})).To(Ω.ConsistOf("EPR.ID"))
			})

			Ψ.It("allows to include all", func() {
				Ω.Expect(ListScope(tx, storage.ListOptions{
					Include: storage.IncludeAll,
				})).To(Ω.ConsistOf("EPR.ID", "ITR.ID"))
			})
		})

		Ψ.Describe("sorting", func() {
			sorted := func(s string) ([]string, error) { return SortScope(tx, s) }

			Ψ.It("sorts by ID", func() {
				Ω.Expect(sorted("id")).To(Ω.Equal([]string{"EPR.ID", "ITR.ID"}))
				Ω.Expect(sorted("-id")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))
			})

			Ψ.It("sorts by last modified", func() {
				Ω.Expect(sorted("last_modified")).To(Ω.Equal([]string{"EPR.ID", "ITR.ID"}))
				Ω.Expect(sorted("-last_modified")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))
			})

			Ψ.It("sorts by any other attribute", func() {
				Ω.Expect(sorted("num")).To(Ω.Equal([]string{"EPR.ID", "ITR.ID"}))
				Ω.Expect(sorted("-num")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))

				Ω.Expect(sorted("str")).To(Ω.Equal([]string{"EPR.ID", "ITR.ID"}))
				Ω.Expect(sorted("-str")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))

				// NULL > NOT NULL
				Ω.Expect(sorted("sub.num")).To(Ω.Equal([]string{"EPR.ID", "ITR.ID"}))
				Ω.Expect(sorted("-sub.num")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))

				// ensure mixed type consistency
				mix1, err := sorted("mix")
				Ω.Expect(err).NotTo(Ω.HaveOccurred())
				mix2, err := sorted("-mix")
				Ω.Expect(err).NotTo(Ω.HaveOccurred())
				for i, j := 0, len(mix2)-1; i < j; i, j = i+1, j-1 {
					mix2[i], mix2[j] = mix2[j], mix2[i]
				}
				Ω.Expect(mix1).To(Ω.Equal(mix2))
			})

			Ψ.It("supports multiple sorts", func() {
				Ω.Expect(sorted("unk,sub.ok,-id")).To(Ω.Equal([]string{"ITR.ID", "EPR.ID"}))
			})
		})

		Ψ.Describe("pagination", func() {
			parse := func(field, value string) params.Filter { return params.ParseFilter(field, value) }
			paginate := func(pagination ...params.Condition) ([]string, error) {
				return ListScope(tx, storage.ListOptions{Pagination: pagination})
			}

			Ψ.It("paginates", func() {
				Ω.Expect(paginate()).To(Ω.HaveLen(2))

				Ω.Expect(paginate(
					params.Condition{}, // ignore blanks
				)).To(Ω.HaveLen(2))

				Ω.Expect(paginate(
					params.Condition{parse("min_id", "AAA"), parse("min_last_modified", "0")},
				)).To(Ω.HaveLen(2))

				Ω.Expect(paginate(
					params.Condition{parse("id", "EPR.ID"), parse("last_modified", etoa(o1.ModTime))},
					params.Condition{parse("id", "ITR.ID"), parse("last_modified", etoa(o2.ModTime))},
				)).To(Ω.HaveLen(2))

				Ω.Expect(paginate(
					params.Condition{parse("id", "EPR.ID"), parse("id", "ITR.ID")},
				)).To(Ω.BeEmpty())

				Ω.Expect(paginate(
					params.Condition{parse("id", "unknown")},
					params.Condition{parse("min_id", "ITR.ID")},
				)).To(Ω.ConsistOf("ITR.ID"))

				Ω.Expect(paginate(
					params.Condition{parse("id", "unknown")},
					params.Condition{}, // ignore blanks
					params.Condition{parse("min_id", "ITR.ID")},
				)).To(Ω.ConsistOf("ITR.ID"))
			})
		})

		Ψ.Describe("conditions", func() {
			filter := func(f, v string) ([]string, error) { return FilterScope(tx, f, v) }
			succeed := func() types.GomegaMatcher {
				return Ω.SatisfyAny(Ω.BeEmpty(), Ω.HaveLen(1), Ω.HaveLen(2))
			}
			conditionalSupport := func(op params.Operator) func(types.GomegaMatcher) types.GomegaMatcher {
				for _, skip := range link.SkipFilters {
					if op == skip {
						return func(_ types.GomegaMatcher) types.GomegaMatcher { return Ω.BeEmpty() }
					}
				}
				return func(m types.GomegaMatcher) types.GomegaMatcher { return m }
			}

			Ψ.It("filters via EQ", func() {
				// ID
				Ω.Expect(filter("id", "EPR.ID")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("id", "ITR.ID")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("id", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("id", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("id", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("id", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("id", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("id", "{}")).To(Ω.BeEmpty())

				// Last-Modified
				Ω.Expect(filter("last_modified", etoa(o1.ModTime))).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("last_modified", etoa(o2.ModTime))).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("last_modified", etoa(o1.ModTime-1))).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("last_modified", "{}")).To(Ω.BeEmpty())

				// Strings
				Ω.Expect(filter("str", "k")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("mix", "val")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("str", "null")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("str", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("str", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("str", "123")).To(Ω.BeEmpty())
				Ω.Expect(filter("str", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("str", "{}")).To(Ω.BeEmpty())

				// Numerics
				Ω.Expect(filter("num", "33")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("num", "66")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("num", "99")).To(Ω.BeEmpty())
				Ω.Expect(filter("num", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("num", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("num", `"xx"`)).To(Ω.BeEmpty())
				Ω.Expect(filter("num", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("num", "{}")).To(Ω.BeEmpty())
				Ω.Expect(filter("sub.num", "11")).To(Ω.ConsistOf("EPR.ID"))

				// Boolean
				Ω.Expect(filter("yes", "true")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("mix", "true")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("yes", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("yes", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("yes", "{}")).To(Ω.BeEmpty())
				Ω.Expect(filter("sub.ok", "true")).To(Ω.HaveLen(2))

				// Objects
				Ω.Expect(filter("sub", `{"ok": true}`)).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("sub", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("sub", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("sub", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("sub", `{}`)).To(Ω.BeEmpty())
				Ω.Expect(filter("sub", `[]`)).To(Ω.BeEmpty())

				// Arrays
				Ω.Expect(filter("ary", `["x", 7, null, false, {"z": 8}]`)).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("ary", "null")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("ary", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("ary", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("ary", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("ary", "[]")).To(Ω.BeEmpty())

				// Missing
				Ω.Expect(filter("unk", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("unk", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("unk", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("unk", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("unk", `{}`)).To(Ω.BeEmpty())
				Ω.Expect(filter("unk", `[]`)).To(Ω.BeEmpty())
			})

			Ψ.It("filters via NOT", func() {
				// ID
				Ω.Expect(filter("not_id", "EPR.ID")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("not_id", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_id", "null")).To(Ω.HaveLen(2))

				// Last-Modified
				Ω.Expect(filter("not_last_modified", etoa(o1.ModTime))).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("not_last_modified", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_last_modified", "null")).To(Ω.HaveLen(2))

				// Data
				Ω.Expect(filter("not_str", "k")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("not_str", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_str", "null")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("not_num", "33")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("not_num", "99")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_num", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_sub", "[]")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_sub", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_sub", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_ary", "[]")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_ary", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_ary", "null")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("not_unk", "33")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_unk", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("not_unk", "null")).To(Ω.BeEmpty())
			})

			Ψ.It("filters via LIKE", func() {
				// ID
				Ω.Expect(filter("like_id", "EPR.ID")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("like_id", "EPR*")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("like_id", "*ID")).To(Ω.HaveLen(2))
				Ω.Expect(filter("like_id", "*R*ID")).To(Ω.HaveLen(2))
				Ω.Expect(filter("like_id", "R.")).To(Ω.HaveLen(2))
				Ω.Expect(filter("like_id", "IP*")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "[]")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_id", "{}")).To(Ω.BeEmpty())

				// Last-Modified
				Ω.Expect(filter("like_last_modified", "*")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_last_modified", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_last_modified", "xx")).To(Ω.BeEmpty())

				// Data
				Ω.Expect(filter("like_str", "k")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("like_mix", "*")).To(Ω.HaveLen(2))
				Ω.Expect(filter("like_mix", "v*")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("like_str", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_str", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_num", "33")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("like_num", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_num", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_unk", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_unk", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("like_unk", "null")).To(Ω.BeEmpty())
			})

			Ψ.It("filters via HAS", func() {
				// ID
				Ω.Expect(filter("has_id", "true")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_id", "false")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_id", "0")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_id", "1")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_id", "xx")).To(Ω.BeEmpty())

				// Last-Modified
				Ω.Expect(filter("has_last_modified", "true")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_last_modified", "false")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_last_modified", "0")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_last_modified", "1")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_last_modified", "xx")).To(Ω.BeEmpty())

				// Data
				Ω.Expect(filter("has_str", "true")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("has_str", "false")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("has_num", "true")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_num", "false")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_mix", "true")).To(Ω.HaveLen(2))
				Ω.Expect(filter("has_mix", "false")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_ary", "true")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("has_ary", "false")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("has_unk", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("has_unk", "false")).To(Ω.HaveLen(2))
			})

			Ψ.It("filters via GT", func() {
				// ID
				Ω.Expect(filter("gt_id", "EPR.ID")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_id", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_id", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_id", "true")).To(succeed())
				Ω.Expect(filter("gt_id", "false")).To(succeed())
				Ω.Expect(filter("gt_id", "33")).To(succeed())
				Ω.Expect(filter("gt_id", "[]")).To(succeed())
				Ω.Expect(filter("gt_id", "{}")).To(succeed())

				// Last-Modified
				Ω.Expect(filter("gt_last_modified", etoa(o1.ModTime))).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_last_modified", etoa(o1.ModTime-1))).To(Ω.HaveLen(2))
				Ω.Expect(filter("gt_last_modified", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("gt_last_modified", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_last_modified", "true")).To(succeed())
				Ω.Expect(filter("gt_last_modified", "false")).To(succeed())
				Ω.Expect(filter("gt_last_modified", "[]")).To(succeed())
				Ω.Expect(filter("gt_last_modified", "{}")).To(succeed())

				// Strings (NULL > NOT NULL)
				Ω.Expect(filter("gt_str", "k")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_str", "a")).To(Ω.HaveLen(2))
				Ω.Expect(filter("gt_str", "z")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_str", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_str", "true")).To(succeed())
				Ω.Expect(filter("gt_str", "false")).To(succeed())
				Ω.Expect(filter("gt_str", "123")).To(succeed())
				Ω.Expect(filter("gt_str", `[]`)).To(succeed())
				Ω.Expect(filter("gt_str", `{}`)).To(succeed())

				// Numerics
				Ω.Expect(filter("gt_num", "33")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_num", "11")).To(Ω.HaveLen(2))
				Ω.Expect(filter("gt_num", "99")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_num", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("gt_num", "true")).To(succeed())
				Ω.Expect(filter("gt_num", "xx")).To(succeed())

				// Booleans (NULL > NOT NULL)
				Ω.Expect(filter("gt_yes", "true")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_yes", "false")).To(Ω.HaveLen(2))
				Ω.Expect(filter("gt_non", "true")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("gt_non", "false")).To(Ω.ConsistOf("ITR.ID"))

				// Arrays
				Ω.Expect(filter("gt_ary", "null")).To(Ω.BeEmpty())

				// Objects
				Ω.Expect(filter("gt_sub", "null")).To(Ω.BeEmpty())
			})

			Ψ.It("filters via LT", func() {
				// ID
				Ω.Expect(filter("lt_id", "EPR.ID")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_id", "xx")).To(Ω.HaveLen(2))
				Ω.Expect(filter("lt_id", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("lt_id", "true")).To(succeed())
				Ω.Expect(filter("lt_id", "false")).To(succeed())
				Ω.Expect(filter("lt_id", "33")).To(succeed())
				Ω.Expect(filter("lt_id", "[]")).To(succeed())
				Ω.Expect(filter("lt_id", "{}")).To(succeed())

				// Last-Modified
				Ω.Expect(filter("lt_last_modified", etoa(o1.ModTime))).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_last_modified", etoa(o2.ModTime))).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("lt_last_modified", "xx")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_last_modified", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("lt_last_modified", "true")).To(succeed())
				Ω.Expect(filter("lt_last_modified", "false")).To(succeed())
				Ω.Expect(filter("lt_last_modified", "[]")).To(succeed())
				Ω.Expect(filter("lt_last_modified", "{}")).To(succeed())

				// Strings
				Ω.Expect(filter("lt_str", "k")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_str", "a")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_str", "z")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("lt_str", "null")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("lt_str", "true")).To(succeed())
				Ω.Expect(filter("lt_str", "false")).To(succeed())
				Ω.Expect(filter("lt_str", "123")).To(succeed())
				Ω.Expect(filter("lt_str", `[]`)).To(succeed())
				Ω.Expect(filter("lt_str", `{}`)).To(succeed())

				// Numerics
				Ω.Expect(filter("lt_num", "33")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_num", "11")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_num", "99")).To(Ω.HaveLen(2))
				Ω.Expect(filter("lt_num", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("lt_num", "true")).To(succeed())
				Ω.Expect(filter("lt_num", "xx")).To(succeed())

				// Booleans
				Ω.Expect(filter("lt_yes", "true")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_yes", "false")).To(Ω.BeEmpty())
				Ω.Expect(filter("lt_non", "true")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("lt_non", "false")).To(Ω.BeEmpty())

				// Arrays
				Ω.Expect(filter("lt_ary", "null")).To(Ω.ConsistOf("EPR.ID"))

				// Objects
				Ω.Expect(filter("lt_sub", "null")).To(Ω.HaveLen(2))
			})

			Ψ.It("filters via MIN", func() {
				// No need to edge cases, covered by GT tests.
				Ω.Expect(filter("min_id", "EPR.ID")).To(Ω.HaveLen(2))
				Ω.Expect(filter("min_id", "H")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("min_id", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("min_last_modified", etoa(o1.ModTime))).To(Ω.HaveLen(2))
				Ω.Expect(filter("min_last_modified", "null")).To(Ω.BeEmpty())
				Ω.Expect(filter("min_str", "k")).To(Ω.HaveLen(2))
				Ω.Expect(filter("min_str", "null")).To(Ω.ConsistOf("ITR.ID"))
				Ω.Expect(filter("min_num", "33")).To(Ω.HaveLen(2))
				Ω.Expect(filter("min_num", "11")).To(Ω.HaveLen(2))
				Ω.Expect(filter("min_num", "99")).To(Ω.BeEmpty())
				Ω.Expect(filter("min_num", "null")).To(Ω.BeEmpty())
			})

			Ψ.It("filters via MAX", func() {
				// No need to edge cases, covered by GT tests.
				Ω.Expect(filter("max_id", "EPR.ID")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("max_id", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("max_last_modified", etoa(o1.ModTime))).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("max_last_modified", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("max_str", "k")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("max_str", "null")).To(Ω.HaveLen(2))
				Ω.Expect(filter("max_num", "33")).To(Ω.ConsistOf("EPR.ID"))
				Ω.Expect(filter("max_num", "11")).To(Ω.BeEmpty())
				Ω.Expect(filter("max_num", "99")).To(Ω.HaveLen(2))
				Ω.Expect(filter("max_num", "null")).To(Ω.HaveLen(2))
			})

			Ψ.It("filters via IN", func() {
				ifSupported := conditionalSupport(params.OperatorIN)

				// ID
				Ω.Expect(filter("in_id", "EPR.ID,ITR.ID")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("in_id", "X,EPR.ID,Z")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("in_id", "X,Y,Z")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_id", "")).To(ifSupported(Ω.BeEmpty()))

				// Last-Modified
				Ω.Expect(filter("in_last_modified", etoa(o1.ModTime))).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("in_last_modified", etoa(o2.ModTime))).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("in_last_modified", "1,2,3")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_last_modified", "a,b,c")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_last_modified", "")).To(ifSupported(Ω.BeEmpty()))

				// Strings
				Ω.Expect(filter("in_str", "k,l,m")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("in_str", "x,y,z")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_str", "null")).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("in_str", "true")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_str", "1,2,3")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_str", "k,null,m")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("in_str", "")).To(ifSupported(Ω.BeEmpty()))

				// Numerics
				Ω.Expect(filter("in_num", "33")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("in_num", "11,33,66")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("in_num", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_num", "true")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_num", "x,y,z")).To(ifSupported(Ω.BeEmpty()))

				// Booleans
				Ω.Expect(filter("in_yes", "x,true,3")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("in_yes", "x,false,3")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_non", "x,true,3")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_non", "x,false,3")).To(ifSupported(Ω.ConsistOf("EPR.ID")))

				// Arrays
				Ω.Expect(filter("in_ary", "null")).To(ifSupported(Ω.ConsistOf("ITR.ID")))

				// Objects
				Ω.Expect(filter("in_sub", "null")).To(ifSupported(Ω.BeEmpty()))

				// Missing
				Ω.Expect(filter("in_unk", "x,y,z")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("in_unk", "x,null,z")).To(ifSupported(Ω.HaveLen(2)))
			})

			Ψ.It("filters via EXCLUDE", func() {
				ifSupported := conditionalSupport(params.OperatorEXCLUDE)

				// ID
				Ω.Expect(filter("exclude_id", "EPR.ID,ITR.ID")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("exclude_id", "X,EPR.ID,Z")).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("exclude_id", "X,Y,Z")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_id", "")).To(ifSupported(Ω.HaveLen(2)))

				// Last-Modified
				Ω.Expect(filter("exclude_last_modified", etoa(o1.ModTime))).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("exclude_last_modified", etoa(o2.ModTime))).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("exclude_last_modified", "1,2,3")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_last_modified", "a,b,c")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_last_modified", "")).To(ifSupported(Ω.HaveLen(2)))

				// Strings
				Ω.Expect(filter("exclude_str", "k,l,m")).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("exclude_str", "x,y,z")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_str", "null")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("exclude_str", "true")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_str", "1,2,3")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_str", "k,null,m")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("exclude_str", "")).To(ifSupported(Ω.HaveLen(2)))

				// Numerics
				Ω.Expect(filter("exclude_num", "33")).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("exclude_num", "11,33,66")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("exclude_num", "null")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_num", "true")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_num", "x,y,z")).To(ifSupported(Ω.HaveLen(2)))

				// Booleans
				Ω.Expect(filter("exclude_yes", "x,true,3")).To(ifSupported(Ω.ConsistOf("ITR.ID")))
				Ω.Expect(filter("exclude_yes", "x,false,3")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_non", "x,true,3")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_non", "x,false,3")).To(ifSupported(Ω.ConsistOf("ITR.ID")))

				// Arrays
				Ω.Expect(filter("exclude_ary", "null")).To(ifSupported(Ω.ConsistOf("EPR.ID")))

				// Objects
				Ω.Expect(filter("exclude_sub", "null")).To(ifSupported(Ω.HaveLen(2)))

				// Missing
				Ω.Expect(filter("exclude_unk", "x,y,z")).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("exclude_unk", "x,null,z")).To(ifSupported(Ω.BeEmpty()))
			})

			Ψ.It("filters via CONTAINS", func() {
				ifSupported := conditionalSupport(params.OperatorContains)

				// ID - always false
				Ω.Expect(filter("contains_id", "EPR.ID")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_id", "ROC")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_id", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_id", "")).To(ifSupported(Ω.BeEmpty()))

				// Last-Modified - always false
				Ω.Expect(filter("contains_last_modified", etoa(o1.ModTime))).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_last_modified", "0")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_last_modified", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_last_modified", "")).To(ifSupported(Ω.BeEmpty()))

				// Strings
				Ω.Expect(filter("contains_str", `k`)).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_str", `123`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_str", `null`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_str", `true`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_str", `"xx"`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_str", ``)).To(ifSupported(Ω.BeEmpty()))

				// Numerics
				Ω.Expect(filter("contains_num", "33")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_num", "99")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_num", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_num", "true")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_num", `"xx"`)).To(ifSupported(Ω.BeEmpty()))

				// Booleans
				Ω.Expect(filter("contains_yes", "true")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_yes", "false")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_yes", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_non", "true")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_non", "false")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_non", "null")).To(ifSupported(Ω.BeEmpty()))

				// Arrays
				Ω.Expect(filter("contains_ary", "x")).To(ifSupported(Ω.ConsistOf("EPR.ID")))

				// Objects
				Ω.Expect(filter("contains_sub", `{"ok": true}`)).To(ifSupported(Ω.HaveLen(2)))
				Ω.Expect(filter("contains_sub", `{"num": 11}`)).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_sub", `{"num": 12}`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_sub", `{"unk": true}`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_sub", `11`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_sub", `true`)).To(ifSupported(Ω.BeEmpty()))

				// Missing - always false
				Ω.Expect(filter("contains_unk", "xx")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_unk", "null")).To(ifSupported(Ω.BeEmpty()))
			})

			Ψ.It("filters via CONTAINS_ANY", func() {
				ifSupported := conditionalSupport(params.OperatorContainsAny)

				// ID - always false
				Ω.Expect(filter("contains_any_id", "EPR.ID")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_id", "EPR.ID,ITR.ID")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_id", "null")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_id", "")).To(ifSupported(Ω.BeEmpty()))

				// Last-Modified - always false
				Ω.Expect(filter("contains_any_last_modified", etoa(o1.ModTime))).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_last_modified", "")).To(ifSupported(Ω.BeEmpty()))

				// Strings - always false
				Ω.Expect(filter("contains_any_str", `k`)).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_str", "")).To(ifSupported(Ω.BeEmpty()))

				// Numerics - always false
				Ω.Expect(filter("contains_any_num", `33`)).To(ifSupported(Ω.BeEmpty()))

				// Arrays
				Ω.Expect(filter("contains_any_ary", "x")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", "x,y,z")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", "5,6,7")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", "w,false")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", `{"z":8}`)).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", "null")).To(ifSupported(Ω.ConsistOf("EPR.ID")))
				Ω.Expect(filter("contains_any_ary", "8,9")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_ary", "u,v,w")).To(ifSupported(Ω.BeEmpty()))
				Ω.Expect(filter("contains_any_ary", "true")).To(ifSupported(Ω.BeEmpty()))

				// Objects - always false
				Ω.Expect(filter("contains_any_sub", `{"ok": true}`)).To(ifSupported(Ω.BeEmpty()))

				// Missing - always false
				Ω.Expect(filter("contains_any_unk", `xx`)).To(ifSupported(Ω.BeEmpty()))
			})
		})
	})
}

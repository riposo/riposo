package storage

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/benbjohnson/clock"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

func init() {
	storage.Register("memory", func(_ context.Context, _ *url.URL, hlp riposo.Helpers) (storage.Backend, error) {
		return New(nil, hlp), nil
	})
}

type backend struct {
	cc   clock.Clock
	hlp  riposo.Helpers
	tree objectTree
	dead objectTree
	mu   sync.Mutex
}

// New inits a new in-memory store. Please use for development and testing only!
func New(cc clock.Clock, hlp riposo.Helpers) storage.Backend {
	if cc == nil {
		cc = clock.New()
	}

	return &backend{
		cc:   cc,
		hlp:  hlp,
		tree: make(objectTree),
		dead: make(objectTree),
	}
}

// Ping implements Backend interface.
func (*backend) Ping(_ context.Context) error {
	return nil
}

// Close implements Backend interface.
func (*backend) Close() error {
	return nil
}

// Begin implements Backend interface.
func (b *backend) Begin(_ context.Context) (storage.Transaction, error) {
	b.mu.Lock()
	return &transaction{b: b}, nil
}

func (b *backend) delete(path riposo.Path, epoch riposo.Epoch, requireExact bool, cb func(string, *schema.Object, bool)) {
	ns, objID := path.Split()

	// fetch node
	node := b.tree.GetNode(ns)
	if node == nil {
		return
	}

	// delete object
	obj := node.Del(objID, epoch)
	if obj != nil {
		cb(ns, obj, true)
		b.dead.FetchNode(ns, 0).ForcePut(obj)
	} else if requireExact {
		return
	}

	// delete nested
	for nns, node := range b.tree {
		if strings.HasPrefix(nns, path.String()) {
			for objID := range node.objects {
				if obj := node.Del(objID, epoch); obj != nil {
					cb(nns, obj, false)
					b.dead.FetchNode(nns, 0).ForcePut(obj)
				}
			}
		}
	}
}

// --------------------------------------------------------------------

type transaction struct {
	b *backend

	xtree objectTree
	xdead objectTree

	done, flushed bool
}

// Commit implements Transaction interface.
func (t *transaction) Commit() error {
	if t.done {
		return storage.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	return nil
}

// Rollback implements Transaction interface.
func (t *transaction) Rollback() error {
	if t.done {
		return storage.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	if t.flushed {
		t.b.tree = t.xtree
		t.b.dead = t.xdead
		return nil
	}

	for k, v := range t.xtree {
		if v == nil {
			delete(t.b.tree, k)
		} else {
			t.b.tree[k] = v
		}
	}
	for k, v := range t.xdead {
		if v == nil {
			delete(t.b.dead, k)
		} else {
			t.b.dead[k] = v
		}
	}
	return nil
}

// Flush implements Transaction interface.
func (t *transaction) Flush() error {
	if t.done {
		return storage.ErrTxDone
	}

	if !t.flushed {
		t.flushed = true
		t.xtree = t.b.tree
		t.xdead = t.b.dead
	}
	t.b.tree = make(objectTree)
	t.b.dead = make(objectTree)
	return nil
}

// ModTime implements Transaction interface.
func (t *transaction) ModTime(path riposo.Path) (riposo.Epoch, error) {
	if !path.IsNode() {
		return 0, storage.ErrInvalidPath
	}
	if t.done {
		return 0, storage.ErrTxDone
	}

	ns, _ := path.Split()
	if node := t.b.tree.GetNode(ns); node != nil {
		return node.modTime, nil
	}
	return 0, nil
}

// Exists implements Transaction interface.
func (t *transaction) Exists(path riposo.Path) (bool, error) {
	if path.IsNode() {
		return false, storage.ErrInvalidPath
	}
	if t.done {
		return false, storage.ErrTxDone
	}

	obj := t.b.tree.Get(path.Split())
	return obj != nil, nil
}

// Get implements Transaction interface.
func (t *transaction) Get(path riposo.Path) (*schema.Object, error) {
	if path.IsNode() {
		return nil, storage.ErrInvalidPath
	}
	if t.done {
		return nil, storage.ErrTxDone
	}

	obj := t.b.tree.Get(path.Split())
	if obj == nil {
		return nil, storage.ErrNotFound
	}
	return copyObject(obj), nil
}

// GetForUpdate implements Transaction interface.
func (t *transaction) GetForUpdate(path riposo.Path) (*schema.Object, error) {
	return t.Get(path)
}

// Create implements Transaction interface.
func (t *transaction) Create(path riposo.Path, obj *schema.Object) error {
	if !path.IsNode() {
		return storage.ErrInvalidPath
	}
	if t.done {
		return storage.ErrTxDone
	}

	now := riposo.EpochFromTime(t.b.cc.Now())
	ns, _ := path.Split()
	t.backup(ns)

	if obj.ID != "" {
		if exst := t.b.tree.Get(ns, obj.ID); exst != nil {
			return storage.ErrObjectExists
		}
	} else {
		obj.ID = t.b.hlp.NextID()
	}

	obj.Norm()
	t.b.dead.Unlink(ns, obj.ID)
	t.b.tree.FetchNode(ns, 0).Put(obj, now)
	return nil
}

// Update implements Transaction interface.
func (t *transaction) Update(path riposo.Path, obj *schema.Object) error {
	if t.done {
		return storage.ErrTxDone
	}

	now := riposo.EpochFromTime(t.b.cc.Now())
	ns, _ := path.Split()
	t.backup(ns)

	obj.Norm()
	t.b.dead.Unlink(ns, obj.ID)
	t.b.tree.FetchNode(ns, 0).Put(obj, now)
	return nil
}

// Delete implements Transaction interface.
func (t *transaction) Delete(path riposo.Path) (*schema.Object, error) {
	if path.IsNode() {
		return nil, storage.ErrInvalidPath
	}
	if t.done {
		return nil, storage.ErrTxDone
	}

	now := riposo.EpochFromTime(t.b.cc.Now())
	ns, _ := path.Split()
	t.backup(ns)

	var deleted *schema.Object
	t.b.delete(path, now, true, func(_ string, obj *schema.Object, exact bool) {
		if exact {
			deleted = obj
		}
	})
	if deleted == nil {
		return nil, storage.ErrNotFound
	}
	return deleted, nil
}

// ListAll implements Transaction interface.
func (t *transaction) ListAll(objs []*schema.Object, path riposo.Path, opt storage.ListOptions) ([]*schema.Object, error) {
	if !path.IsNode() {
		return nil, storage.ErrInvalidPath
	}
	if t.done {
		return nil, storage.ErrTxDone
	}

	ns, _ := path.Split()

	t.b.tree.Each(ns, opt.Condition, func(obj *schema.Object) {
		objs = append(objs, obj)
	})

	if opt.Include == storage.IncludeAll {
		t.b.dead.Each(ns, opt.Condition, func(obj *schema.Object) {
			objs = append(objs, obj)
		})
	}

	objs = paginationFilter(objs, opt.Pagination)
	if len(opt.Sort) != 0 {
		sort.Sort(&objectSlice{Slice: objs, Sort: opt.Sort})
	}
	if opt.Limit > 0 && len(objs) > opt.Limit {
		objs = objs[:opt.Limit]
	}
	return objs, nil
}

// CountAll implements Transaction interface.
func (t *transaction) CountAll(path riposo.Path, opt storage.CountOptions) (int64, error) {
	if !path.IsNode() {
		return 0, storage.ErrInvalidPath
	}
	if t.done {
		return 0, storage.ErrTxDone
	}

	ns, _ := path.Split()
	cnt := int64(0)

	t.b.tree.Each(ns, opt.Condition, func(obj *schema.Object) {
		cnt++
	})
	return cnt, nil
}

// DeleteAll implements Transaction interface.
func (t *transaction) DeleteAll(paths []riposo.Path) (modTime riposo.Epoch, deleted []riposo.Path, _ error) {
	for _, path := range paths {
		if path.IsNode() {
			return 0, nil, storage.ErrInvalidPath
		}
	}
	if t.done {
		return 0, nil, storage.ErrTxDone
	}
	if len(paths) == 0 {
		return 0, nil, nil
	}

	now := riposo.EpochFromTime(t.b.cc.Now())
	for _, path := range paths {
		ns, _ := path.Split()
		t.backup(ns)

		t.b.delete(path, now, false, func(ns string, obj *schema.Object, exact bool) {
			if exact && obj.ModTime > modTime {
				modTime = obj.ModTime
			}
			deleted = append(deleted, riposo.JoinPath(ns, obj.ID))
		})
	}
	return
}

// Purge implements Transaction interface.
func (t *transaction) Purge(olderThan riposo.Epoch) (cnt int64, err error) {
	if t.done {
		return 0, storage.ErrTxDone
	}

	for ns, node := range t.b.dead {
		t.backup(ns)

		for oid, obj := range node.objects {
			if olderThan == 0 || obj.ModTime < olderThan {
				delete(node.objects, oid)
				cnt++
			}
		}
		if len(node.objects) == 0 {
			delete(t.b.dead, ns)
		}
	}
	return
}

func (t *transaction) backup(ns string) {
	if t.flushed {
		return
	}

	if t.xtree == nil {
		t.xtree = make(objectTree)
	}
	if _, ok := t.xtree[ns]; !ok {
		t.xtree[ns] = t.b.tree[ns].Copy()
	}

	if t.xdead == nil {
		t.xdead = make(objectTree)
	}
	if _, ok := t.xdead[ns]; !ok {
		t.xdead[ns] = t.b.dead[ns].Copy()
	}
}

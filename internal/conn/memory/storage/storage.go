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
	storage.Register("memory", func(_ context.Context, _ *url.URL, hlp *riposo.Helpers) (storage.Backend, error) {
		return New(nil, hlp), nil
	})
}

type updateHandle struct {
	obj  *schema.Object
	path riposo.Path
}

func (t *updateHandle) Object() *schema.Object { return t.obj }

type backend struct {
	cc   clock.Clock
	hlp  *riposo.Helpers
	tree objectTree
	dead objectTree
	mu   sync.RWMutex
}

// New inits a new in-memory store. Please use for development and testing only!
func New(cc clock.Clock, hlp *riposo.Helpers) storage.Backend {
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
	return b, nil
}

// Commit implements Transaction interface.
func (b *backend) Commit() error {
	return nil
}

// Rollback implements Transaction interface.
func (b *backend) Rollback() error {
	return nil
}

// Flush implements Transaction interface.
func (b *backend) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tree = make(objectTree)
	b.dead = make(objectTree)
	return nil
}

// ModTime implements Transaction interface.
func (b *backend) ModTime(path riposo.Path) (riposo.Epoch, error) {
	if !path.IsNode() {
		return 0, storage.ErrInvalidPath
	}

	ns, _ := path.Split()

	b.mu.RLock()
	defer b.mu.RUnlock()

	if node := b.tree.GetNode(ns); node != nil {
		return node.modTime, nil
	}
	return 0, nil
}

// Exists implements Transaction interface.
func (b *backend) Exists(path riposo.Path) (bool, error) {
	if path.IsNode() {
		return false, storage.ErrInvalidPath
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	obj := b.tree.Get(path.Split())
	return obj != nil, nil
}

// Get implements Transaction interface.
func (b *backend) Get(path riposo.Path) (*schema.Object, error) {
	if path.IsNode() {
		return nil, storage.ErrInvalidPath
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	obj := b.tree.Get(path.Split())
	if obj == nil {
		return nil, storage.ErrNotFound
	}
	return copyObject(obj), nil
}

// GetForUpdate implements Transaction interface.
func (b *backend) GetForUpdate(path riposo.Path) (storage.UpdateHandle, error) {
	obj, err := b.Get(path)
	if err != nil {
		return nil, err
	}
	return &updateHandle{obj: obj, path: path}, nil
}

// Create implements Transaction interface.
func (b *backend) Create(path riposo.Path, obj *schema.Object) error {
	if !path.IsNode() {
		return storage.ErrInvalidPath
	}

	now := riposo.EpochFromTime(b.cc.Now())
	ns, _ := path.Split()

	b.mu.Lock()
	defer b.mu.Unlock()

	if obj.ID != "" {
		if exst := b.tree.Get(ns, obj.ID); exst != nil {
			return storage.ErrObjectExists
		}
	} else {
		obj.ID = b.hlp.NextID()
	}

	if len(obj.Extra) == 0 {
		obj.Extra = append(obj.Extra, '{', '}')
	}

	b.dead.Unlink(ns, obj.ID)
	b.tree.FetchNode(ns, 0).Put(obj, now)
	return nil
}

// Update implements Transaction interface.
func (b *backend) Update(h storage.UpdateHandle) error {
	uh := h.(*updateHandle)
	now := riposo.EpochFromTime(b.cc.Now())
	ns, _ := uh.path.Split()

	if len(uh.obj.Extra) == 0 {
		uh.obj.Extra = append(uh.obj.Extra, '{', '}')
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.dead.Unlink(ns, uh.obj.ID)
	b.tree.FetchNode(ns, 0).Put(uh.obj, now)
	return nil
}

// Delete implements Transaction interface.
func (b *backend) Delete(path riposo.Path) (*schema.Object, error) {
	if path.IsNode() {
		return nil, storage.ErrInvalidPath
	}

	now := riposo.EpochFromTime(b.cc.Now())

	b.mu.Lock()
	defer b.mu.Unlock()

	var deleted *schema.Object
	b.delete(path, now, true, func(obj *schema.Object, exact bool) {
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
func (b *backend) ListAll(objs []*schema.Object, path riposo.Path, opt storage.ListOptions) ([]*schema.Object, error) {
	if !path.IsNode() {
		return nil, storage.ErrInvalidPath
	}

	ns, _ := path.Split()

	b.mu.RLock()
	defer b.mu.RUnlock()

	b.tree.Each(ns, opt.Condition, func(obj *schema.Object) {
		objs = append(objs, obj)
	})

	if opt.Include == storage.IncludeAll {
		b.dead.Each(ns, opt.Condition, func(obj *schema.Object) {
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
func (b *backend) CountAll(path riposo.Path, opt storage.CountOptions) (int64, error) {
	if !path.IsNode() {
		return 0, storage.ErrInvalidPath
	}

	ns, _ := path.Split()
	cnt := int64(0)

	b.mu.RLock()
	defer b.mu.RUnlock()

	b.tree.Each(ns, opt.Condition, func(obj *schema.Object) {
		cnt++
	})
	return cnt, nil
}

// DeleteAll implements Transaction interface.
func (b *backend) DeleteAll(paths []riposo.Path) (modTime riposo.Epoch, _ error) {
	now := riposo.EpochFromTime(b.cc.Now())

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, path := range paths {
		b.delete(path, now, false, func(obj *schema.Object, exact bool) {
			if exact && obj.ModTime > modTime {
				modTime = obj.ModTime
			}
		})
	}
	return
}

// Purge implements Transaction interface.
func (b *backend) Purge(olderThan riposo.Epoch) (cnt int64, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if olderThan == 0 {
		cnt = int64(b.dead.Len())
		b.dead = make(objectTree)
		return
	}

	for ns, node := range b.dead {
		for oid, obj := range node.objects {
			if obj.ModTime < olderThan {
				delete(node.objects, oid)
				cnt++
			}
		}
		if len(node.objects) == 0 {
			delete(b.dead, ns)
		}
	}
	return
}

func (b *backend) delete(path riposo.Path, epoch riposo.Epoch, requireExact bool, cb func(*schema.Object, bool)) {
	ns, objID := path.Split()

	// fetch node
	node := b.tree.GetNode(ns)
	if node == nil {
		return
	}

	// delete object
	obj := node.Del(objID, epoch)
	if obj != nil {
		cb(obj, true)
		b.dead.FetchNode(ns, 0).ForcePut(obj)
	} else if requireExact {
		return
	}

	// delete nested
	for nns, node := range b.tree {
		if strings.HasPrefix(nns, path.String()) {
			for objID := range node.objects {
				if obj := node.Del(objID, epoch); obj != nil {
					cb(obj, false)
					b.dead.FetchNode(nns, 0).ForcePut(obj)
				}
			}
		}
	}
}

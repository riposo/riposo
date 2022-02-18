package cache

import (
	"context"
	"net/url"
	"sync"
	"time"

	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/riposo"
)

func init() {
	cache.Register("memory", func(context.Context, *url.URL, riposo.Helpers) (cache.Backend, error) {
		return New(), nil
	})
}

type item struct {
	val []byte
	exp time.Time
}

func (it *item) Copy() *item {
	if it == nil {
		return nil
	}

	val := make([]byte, len(it.val))
	copy(val, it.val)
	return &item{val: val, exp: it.exp}
}

func (it *item) Expired(now time.Time) bool {
	return it.exp.Before(now)
}

type backend struct {
	keys map[string]*item
	mu   sync.Mutex

	closer chan struct{}
}

// New inits a new in-memory cache backend. Please use for development and testing only!
func New() cache.Backend {
	b := &backend{
		keys:   make(map[string]*item),
		closer: make(chan struct{}),
	}
	go b.loop()
	return b
}

// Ping implements cache.Backend interface.
func (*backend) Ping(_ context.Context) error {
	return nil
}

// Begin implements cache.Backend interface.
func (b *backend) Begin(_ context.Context) (cache.Transaction, error) {
	b.mu.Lock()
	return &transaction{b: b}, nil
}

// Close implements cache.Backend interface.
func (b *backend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.closer:
	default:
		close(b.closer)
	}
	return nil
}

func (b *backend) loop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		select {
		case <-b.closer:
			return
		case now := <-t.C:
			b.reapExpired(now)
		}
	}
}

func (b *backend) reapExpired(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for key, it := range b.keys {
		if it.Expired(now) {
			delete(b.keys, key)
		}
	}
}

// --------------------------------------------------------------------

type transaction struct {
	b *backend

	xkeys map[string]*item

	done, flushed bool
}

// Commit implements cache.Transaction interface.
func (t *transaction) Commit() error {
	if t.done {
		return cache.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	return nil
}

// Rollback implements cache.Transaction interface.
func (t *transaction) Rollback() error {
	if t.done {
		return cache.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	if t.flushed {
		t.b.keys = t.xkeys
		return nil
	}

	for k, v := range t.xkeys {
		if v == nil {
			delete(t.b.keys, k)
		} else {
			t.b.keys[k] = v
		}
	}
	return nil
}

// Flush implements cache.Transaction interface.
func (t *transaction) Flush() error {
	if t.done {
		return cache.ErrTxDone
	}

	if !t.flushed {
		t.flushed = true
		t.xkeys = t.b.keys
	}
	t.b.keys = make(map[string]*item)
	return nil
}

// Get implements cache.Transaction interface.
func (t *transaction) Get(key string) ([]byte, error) {
	if t.done {
		return nil, cache.ErrTxDone
	}

	it, ok := t.b.keys[key]
	if !ok || it.Expired(time.Now()) {
		return nil, cache.ErrNotFound
	}

	val := make([]byte, len(it.val))
	copy(val, it.val)
	return val, nil
}

// Set implements cache.Transaction interface.
func (t *transaction) Set(key string, val []byte, exp time.Time) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	if t.done {
		return cache.ErrTxDone
	}

	t.backup(key)

	it, ok := t.b.keys[key]
	if !ok {
		it = new(item)
		t.b.keys[key] = it
	}

	it.exp = exp
	it.val = append(it.val[:0], val...)

	return nil
}

// Del implements cache.Transaction interface.
func (t *transaction) Del(key string) error {
	if t.done {
		return cache.ErrTxDone
	}

	t.backup(key)

	it, ok := t.b.keys[key]
	if ok {
		delete(t.b.keys, key)
	}

	if !ok || it.Expired(time.Now()) {
		return cache.ErrNotFound
	}
	return nil
}

func (t *transaction) backup(key string) {
	if t.flushed {
		return
	}

	if t.xkeys == nil {
		t.xkeys = make(map[string]*item)
	}

	if _, ok := t.xkeys[key]; !ok {
		t.xkeys[key] = t.b.keys[key].Copy()
	}
}

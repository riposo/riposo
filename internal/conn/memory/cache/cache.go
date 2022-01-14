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

func (it *item) Expired(now time.Time) bool { return it.exp.Before(now) }

type backend struct {
	keys map[string]*item
	mu   sync.RWMutex

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
	return b, nil
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

// Commit implements cache.Transaction interface.
func (*backend) Commit() error { return nil }

// Rollback implements cache.Transaction interface.
func (*backend) Rollback() error { return nil }

// Flush implements cache.Transaction interface.
func (b *backend) Flush() error {
	b.mu.Lock()
	b.keys = make(map[string]*item)
	b.mu.Unlock()

	return nil
}

// Get implements cache.Transaction interface.
func (b *backend) Get(key string) ([]byte, error) {
	b.mu.RLock()
	it, ok := b.keys[key]
	b.mu.RUnlock()

	if !ok || it.Expired(time.Now()) {
		return nil, cache.ErrNotFound
	}

	val := make([]byte, len(it.val))
	copy(val, it.val)
	return val, nil
}

// Set implements cache.Transaction interface.
func (b *backend) Set(key string, val []byte, exp time.Time) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	it, ok := b.keys[key]
	if !ok {
		it = new(item)
		b.keys[key] = it
	}

	it.exp = exp
	it.val = append(it.val[:0], val...)

	return nil
}

// Del implements cache.Transaction interface.
func (b *backend) Del(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	it, ok := b.keys[key]
	if ok {
		delete(b.keys, key)
	}

	if !ok || it.Expired(time.Now()) {
		return cache.ErrNotFound
	}
	return nil
}

func (b *backend) loop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	var acc []string
	for {
		select {
		case <-b.closer:
			return
		case now := <-t.C:
			b.reapExpired(now, acc[:0])
		}
	}
}

func (b *backend) reapExpired(now time.Time, acc []string) {
	b.mu.RLock()
	for key, it := range b.keys {
		if it.Expired(now) {
			acc = append(acc, key)
		}
	}
	b.mu.RUnlock()

	for _, key := range acc {
		b.mu.Lock()
		if it, ok := b.keys[key]; ok && it.Expired(now) {
			delete(b.keys, key)
		}
		b.mu.Unlock()
	}
}

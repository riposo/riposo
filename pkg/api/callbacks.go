package api

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// Callbacks are a set of callbacks around a model.
type Callbacks interface {
	OnCreate(txn *Txn, path riposo.Path) CreateCallback
	OnUpdate(txn *Txn, path riposo.Path) UpdateCallback
	OnPatch(txn *Txn, path riposo.Path) PatchCallback
	OnDelete(txn *Txn, path riposo.Path) DeleteCallback
	OnDeleteAll(txn *Txn, path riposo.Path) DeleteAllCallback
}

// CreateCallback instances run around create actions.
type CreateCallback interface {
	BeforeCreate(payload *schema.Resource) error
	AfterCreate(created *schema.Resource) error
}

// UpdateCallback instances run around update actions.
type UpdateCallback interface {
	BeforeUpdate(existing *schema.Object, payload *schema.Resource) error
	AfterUpdate(updated *schema.Resource) error
}

// PatchCallback instances run around patch actions.
type PatchCallback interface {
	BeforePatch(existing *schema.Object, payload *schema.Resource) error
	AfterPatch(patched *schema.Resource) error
}

// DeleteCallback instances run around delete actions.
type DeleteCallback interface {
	BeforeDelete(existing *schema.Object) error
	AfterDelete(deleted *schema.Object) error
}

// DeleteAllCallback instances run around multi-delete actions.
type DeleteAllCallback interface {
	BeforeDeleteAll(objIDs []string) error
	AfterDeleteAll(modTime riposo.Epoch, deleted []riposo.Path) error
}

// NoopCallbacks is an embeddable noop callback type.
type NoopCallbacks struct{}

func (NoopCallbacks) OnCreate(_ *Txn, _ riposo.Path) CreateCallback       { return nil }
func (NoopCallbacks) OnUpdate(_ *Txn, _ riposo.Path) UpdateCallback       { return nil }
func (NoopCallbacks) OnPatch(_ *Txn, _ riposo.Path) PatchCallback         { return nil }
func (NoopCallbacks) OnDelete(_ *Txn, _ riposo.Path) DeleteCallback       { return nil }
func (NoopCallbacks) OnDeleteAll(_ *Txn, _ riposo.Path) DeleteAllCallback { return nil }

// --------------------------------------------------------------------

// callbackChain registers callbacks with patterns.
type callbackChain struct {
	cbs []callbacksWithGlobs
}

// Len returns the number of registered callbacks.
func (r *callbackChain) Len() int {
	return len(r.cbs)
}

// Register registers callbacks with glob patterns.
func (r *callbackChain) Register(patterns []string, callbacks Callbacks) {
	globs := make([]callbackGlob, 0, len(patterns))
	for _, pat := range patterns {
		glob := parseCallbackGlob(pat)
		if glob.pattern != "" && doublestar.ValidatePathPattern(glob.pattern) {
			globs = append(globs, glob)
		}
	}
	if len(globs) == 0 {
		return
	}

	r.cbs = append(r.cbs, callbacksWithGlobs{
		Callbacks: callbacks,
		globs:     globs,
	})
}

// ForEach iterates over registered callbacks for a given path.
func (r *callbackChain) ForEach(path riposo.Path, fn func(Callbacks)) {
	s := path.String()
	for _, h := range r.cbs {
		if h.Match(s) {
			fn(h.Callbacks)
		}
	}
}

type callbacksWithGlobs struct {
	Callbacks
	globs []callbackGlob
}

func (cc callbacksWithGlobs) Match(s string) (include bool) {
	for _, pat := range cc.globs {
		// skip if we already have an inclusion match
		if include && !pat.exclude {
			continue
		}

		// skip if we already have an exclusion match
		if !include && pat.exclude {
			continue
		}

		// try match
		if ok, _ := doublestar.PathMatch(pat.pattern, s); ok {
			include = !pat.exclude
		}
	}
	return
}

type callbackGlob struct {
	pattern string
	exclude bool
}

func parseCallbackGlob(s string) callbackGlob {
	if strings.HasPrefix(s, "!") {
		return callbackGlob{pattern: s[1:], exclude: true}
	}
	return callbackGlob{pattern: s}
}

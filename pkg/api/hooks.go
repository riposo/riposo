package api

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// A Hook is a collection of callbacks around a model.
type Hook interface {
	BeforeCreate(txn *Txn, path riposo.Path, payload *schema.Resource) error
	AfterCreate(txn *Txn, path riposo.Path, created *schema.Resource) error

	BeforeUpdate(txn *Txn, path riposo.Path, existing *schema.Object, payload *schema.Resource) error
	AfterUpdate(txn *Txn, path riposo.Path, updated *schema.Resource) error

	BeforePatch(txn *Txn, path riposo.Path, existing *schema.Object, payload *schema.Resource) error
	AfterPatch(txn *Txn, path riposo.Path, patched *schema.Resource) error

	BeforeDelete(txn *Txn, path riposo.Path, existing *schema.Object) error
	AfterDelete(txn *Txn, path riposo.Path, deleted *schema.Object) error

	BeforeDeleteAll(txn *Txn, path riposo.Path, objIDs []string) error
	AfterDeleteAll(txn *Txn, path riposo.Path, objIDs []string, modTime riposo.Epoch, deleted []riposo.Path) error
}

// NoopHook is an embeddable model hook type.
type NoopHook struct{}

func (NoopHook) BeforeCreate(_ *Txn, _ riposo.Path, _ *schema.Resource) error {
	return nil
}
func (NoopHook) AfterCreate(_ *Txn, _ riposo.Path, _ *schema.Resource) error {
	return nil
}
func (NoopHook) BeforeUpdate(_ *Txn, _ riposo.Path, _ *schema.Object, _ *schema.Resource) error {
	return nil
}
func (NoopHook) AfterUpdate(_ *Txn, _ riposo.Path, _ *schema.Resource) error {
	return nil
}
func (NoopHook) BeforePatch(_ *Txn, _ riposo.Path, _ *schema.Object, _ *schema.Resource) error {
	return nil
}
func (NoopHook) AfterPatch(_ *Txn, _ riposo.Path, _ *schema.Resource) error {
	return nil
}
func (NoopHook) BeforeDelete(_ *Txn, _ riposo.Path, _ *schema.Object) error {
	return nil
}
func (NoopHook) AfterDelete(_ *Txn, _ riposo.Path, _ *schema.Object) error {
	return nil
}
func (NoopHook) BeforeDeleteAll(_ *Txn, _ riposo.Path, _ []string) error {
	return nil
}
func (NoopHook) AfterDeleteAll(_ *Txn, _ riposo.Path, _ []string, _ riposo.Epoch, _ []riposo.Path) error {
	return nil
}

// --------------------------------------------------------------------

// hookRegistry registers callbacks with patterns.
type hookRegistry struct {
	hooks []hook
}

// Len returns the number of registered hooks.
func (r *hookRegistry) Len() int {
	return len(r.hooks)
}

// Register registers callbacks with glob patterns.
func (r *hookRegistry) Register(patterns []string, callbacks Hook) {
	globs := make([]hookGlob, 0, len(patterns))
	for _, pat := range patterns {
		glob := parseHookGlob(pat)
		if glob.pattern != "" && doublestar.ValidatePathPattern(glob.pattern) {
			globs = append(globs, parseHookGlob(pat))
		}
	}
	if len(globs) == 0 {
		return
	}

	r.hooks = append(r.hooks, hook{
		globs: globs,
		Hook:  callbacks,
	})
}

// ForEach iterates over registered callbacks for a given path.
func (r *hookRegistry) ForEach(path riposo.Path, fn func(Hook) error) error {
	s := path.String()
	for _, h := range r.hooks {
		if h.Match(s) {
			if err := fn(h.Hook); err != nil {
				return err
			}
		}
	}
	return nil
}

type hook struct {
	globs []hookGlob
	Hook
}

func (h hook) Match(s string) (include bool) {
	for _, pat := range h.globs {
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

type hookGlob struct {
	pattern string
	exclude bool
}

func parseHookGlob(s string) hookGlob {
	if strings.HasPrefix(s, "!") {
		return hookGlob{pattern: s[1:], exclude: true}
	}
	return hookGlob{pattern: s}
}

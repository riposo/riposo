package api

import "github.com/riposo/riposo/pkg/riposo"

type HookChain interface {
	Len() int
	Register(patterns []string, callbacks Hook)
	ForEach(path riposo.Path, fn func(Hook) error) error
}

func NewHookChain() HookChain {
	return new(hookChain)
}

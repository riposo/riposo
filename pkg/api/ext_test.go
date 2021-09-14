package api

import "github.com/riposo/riposo/pkg/riposo"

type HookRegistry interface {
	Len() int
	Register(patterns []string, callbacks Hook)
	ForEach(path riposo.Path, fn func(Hook) error) error
}

func NewHookRegistry() HookRegistry {
	return new(hookRegistry)
}

package api

import "github.com/riposo/riposo/pkg/riposo"

type CallbackChain interface {
	Len() int
	Register(patterns []string, callbacks Callbacks)
	ForEach(path riposo.Path, fn func(Callbacks))
}

func NewCallbackChain() CallbackChain {
	return new(callbackChain)
}

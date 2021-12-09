package config

import (
	"github.com/riposo/riposo/pkg/identity"
	"github.com/riposo/riposo/pkg/slowhash"
)

type helpers struct {
	parseConfig parseFunc
	nextID      identity.Factory
	slowHash    slowhash.Generator
}

func (h *helpers) ParseConfig(target interface{}) error {
	return h.parseConfig(target)
}

func (h *helpers) NextID() string {
	return h.nextID()
}

func (h *helpers) SlowHash(plain string) (string, error) {
	return h.slowHash(plain)
}

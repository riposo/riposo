package riposo

import (
	"github.com/riposo/riposo/pkg/identity"
	"github.com/riposo/riposo/pkg/slowhash"
)

// HelpersOptions are used to configure helpers.
type HelpersOptions struct {
	Identity string // type of identity generator to use
	SlowHash string // type of slowhash to use
}

// Helpers provide access to elementary helper functions.
type Helpers struct {
	slowHash slowhash.Generator
	nextID   identity.Factory
}

// NewHelpers inits helpers.
func NewHelpers(opt HelpersOptions) (*Helpers, error) {
	slowHash, err := slowhash.Get(opt.SlowHash)
	if err != nil {
		return nil, err
	}

	nextID, err := identity.Get(opt.Identity)
	if err != nil {
		return nil, err
	}

	return CustomHelpers(slowHash, nextID), nil
}

// CustomHelpers inits custom helpers.
func CustomHelpers(slowHash slowhash.Generator, nextID identity.Factory) *Helpers {
	return &Helpers{
		slowHash: slowHash,
		nextID:   nextID,
	}
}

// SlowHash applies a password-hash to a plain string
// returning a cryptographically secure, hashed string.
func (e *Helpers) SlowHash(plain string) (string, error) {
	return e.slowHash(plain)
}

// NextID returns a new globally unique ID.
func (e *Helpers) NextID() string {
	return e.nextID()
}

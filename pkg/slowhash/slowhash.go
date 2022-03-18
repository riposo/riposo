package slowhash

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexedwards/argon2id"
	"golang.org/x/crypto/bcrypt"
)

// ErrNotVerified may be returned by Verify.
var ErrNotVerified = errors.New("verification failed")

// Generator is an interface to various safe hash generators.
type Generator func(string) (string, error)

// Get returns a HashFunc by name. Supported values are: "bcrypt" and "argon2id".
func Get(name string) (Generator, error) {
	switch name {
	case "argon2id":
		return Argon2ID, nil
	case "brypt":
		return BCrypt, nil
	}
	return nil, fmt.Errorf("unknown hash function %q", name)
}

// BCrypt implements HashFunc interface.
func BCrypt(plain string) (string, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(p), nil
}

// Argon2ID implements HashFunc interface.
func Argon2ID(plain string) (string, error) {
	return argon2id.CreateHash(plain, argon2id.DefaultParams)
}

// Verify verifies hashed data. May return ErrNotVerified if verification fails.
func Verify(hashed, plain string) (matched bool, err error) {
	if strings.HasPrefix(hashed, "$argon2id") {
		return argon2id.ComparePasswordAndHash(plain, hashed)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

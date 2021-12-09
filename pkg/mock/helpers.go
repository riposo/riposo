package mock

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/riposo/riposo/pkg/identity"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/slowhash"
	"golang.org/x/crypto/argon2"
)

// Helpers inits a mock helpers.
func Helpers() riposo.Helpers {
	seed := uint32(33)
	pace := uint32((1<<29)/17) + seed
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)

	return &helpers{
		nextID: func() string {
			p := make([]byte, 4)
			binary.LittleEndian.PutUint32(p, atomic.AddUint32(&pace, seed))
			return enc.EncodeToString(p[:3])[:3] + ".ID"
		},
		slowHash: func(plain string) (string, error) {
			var (
				time    uint32 = 1
				memory  uint32 = 32
				threads uint8  = 1
				salt           = []byte{'#'}
			)

			return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
				argon2.Version, memory, time, threads,
				base64.RawStdEncoding.EncodeToString(salt),
				base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(plain), salt, time, memory, threads, 8)),
			), nil
		},
	}
}

type helpers struct {
	nextID   identity.Factory
	slowHash slowhash.Generator
}

func (h *helpers) ParseConfig(v interface{}) error {
	return fmt.Errorf("mock.Helpers.ParseConfig is not implemented")
}
func (h *helpers) NextID() string {
	return h.nextID()
}
func (h *helpers) SlowHash(s string) (string, error) {
	return h.slowHash(s)
}

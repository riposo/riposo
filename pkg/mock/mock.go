package mock

import (
	"time"

	"github.com/benbjohnson/clock"
	memc "github.com/riposo/riposo/internal/conn/memory/cache"
	memp "github.com/riposo/riposo/internal/conn/memory/permission"
	mems "github.com/riposo/riposo/internal/conn/memory/storage"
	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/riposo"
)

// Conns returns an in-memory connections.
func Conns(hlp *riposo.Helpers) *conn.Set {
	if hlp == nil {
		hlp = Helpers()
	}

	return conn.Use(
		mems.New(Clock(), hlp),
		memp.New(),
		memc.New(),
	)
}

// Clock returns a mock clock.
func Clock() clock.Clock {
	cc := clock.NewMock()
	cc.Set(time.Unix(1515151515, 676_767_676))
	return cc
}

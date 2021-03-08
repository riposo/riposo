package postgres

import (
	_ "github.com/riposo/riposo/internal/conn/postgres/cache"      // cache backend
	_ "github.com/riposo/riposo/internal/conn/postgres/permission" // permission backend
	_ "github.com/riposo/riposo/internal/conn/postgres/storage"    // storage backend
)

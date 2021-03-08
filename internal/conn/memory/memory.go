package memory

import (
	_ "github.com/riposo/riposo/internal/conn/memory/cache"      // cache backend
	_ "github.com/riposo/riposo/internal/conn/memory/permission" // permission backend
	_ "github.com/riposo/riposo/internal/conn/memory/storage"    // storage backend
)

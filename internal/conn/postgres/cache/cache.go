package cache

import (
	"context"
	"database/sql"
	"embed"
	"net/url"
	"time"

	"github.com/riposo/riposo/internal/conn/postgres/common"
	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
)

const schemaVersion = 1

//go:embed schema.sql
var embedFS embed.FS

func init() {
	cache.Register("postgres", func(ctx context.Context, uri *url.URL, _ riposo.Helpers) (cache.Backend, error) {
		return Connect(ctx, uri.String())
	})
}

// --------------------------------------------------------------------

type conn struct {
	db   *sql.DB
	stop context.CancelFunc
	stmt struct {
		getKey, setKey, delKey, prune *sql.Stmt
	}
}

// Connect connects to a PostgreSQL server.
func Connect(ctx context.Context, dsn string) (cache.Backend, error) {
	// Connect to the DB.
	db, err := common.Connect(ctx, dsn, "cache_schema_version", schemaVersion, embedFS)
	if err != nil {
		return nil, err
	}

	// Create connection struct, prepare statements.
	cn := &conn{db: db}
	if cn.stmt.getKey, err = db.PrepareContext(ctx, sqlGetKey); err != nil {
		_ = cn.Close()
		return nil, err
	}
	if cn.stmt.setKey, err = db.PrepareContext(ctx, sqlSetKey); err != nil {
		_ = cn.Close()
		return nil, err
	}
	if cn.stmt.delKey, err = db.PrepareContext(ctx, sqlDelKey); err != nil {
		_ = cn.Close()
		return nil, err
	}
	if cn.stmt.prune, err = db.PrepareContext(ctx, sqlPrune); err != nil {
		_ = cn.Close()
		return nil, err
	}

	// Prune expired.
	if err := cn.pruneExpired(ctx, time.Now()); err != nil {
		_ = cn.Close()
		return nil, err
	}

	// Setup periodic prunning in the background.
	pruneCtx, stop := context.WithCancel(ctx)
	cn.stop = stop
	go cn.pruneLoop(pruneCtx)

	return cn, nil
}

// Ping implements cache.Backend interface.
func (cn *conn) Ping(ctx context.Context) error {
	return cn.db.PingContext(ctx)
}

// Begin implements cache.Backend interface.
func (cn *conn) Begin(ctx context.Context) (cache.Transaction, error) {
	tx, err := cn.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &transaction{Tx: tx, cn: cn, ctx: ctx}, nil
}

// Close implements cache.Backend.
func (cn *conn) Close() (err error) {
	if cn.stop != nil {
		cn.stop() // stop prune loop
	}

	if cn.stmt.getKey != nil {
		err = multierr.Append(err, cn.stmt.getKey.Close())
	}
	if cn.stmt.setKey != nil {
		err = multierr.Append(err, cn.stmt.setKey.Close())
	}
	if cn.stmt.delKey != nil {
		err = multierr.Append(err, cn.stmt.delKey.Close())
	}
	if cn.stmt.prune != nil {
		err = multierr.Append(err, cn.stmt.prune.Close())
	}
	return
}

func (cn *conn) pruneExpired(ctx context.Context, now time.Time) error {
	_, err := cn.stmt.prune.ExecContext(ctx, now.UTC())
	return err
}

func (cn *conn) pruneLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			_ = cn.pruneExpired(ctx, now)
		}
	}
}

// --------------------------------------------------------------------

type transaction struct {
	*sql.Tx
	cn  *conn
	ctx context.Context
}

// Flush implements cache.Transaction interface.
func (tx *transaction) Flush() error {
	_, err := tx.ExecContext(tx.ctx, `TRUNCATE cache_keys`)
	return err
}

// Get implements cache.Transaction.
func (tx *transaction) Get(key string) ([]byte, error) {
	if err := cache.ValidateKey(key); err != nil {
		return nil, err
	}

	var val []byte
	err := tx.StmtContext(tx.ctx, tx.cn.stmt.getKey).
		QueryRowContext(tx.ctx, key, time.Now().UTC()).
		Scan(&val)
	if err == sql.ErrNoRows {
		return nil, cache.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return val, nil
}

// Set implements cache.Transaction.
func (tx *transaction) Set(key string, val []byte, exp time.Time) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	_, err := tx.
		StmtContext(tx.ctx, tx.cn.stmt.setKey).
		ExecContext(tx.ctx, key, val, exp.UTC())
	return err
}

// Del implements cache.Transaction.
func (tx *transaction) Del(key string) error {
	if err := cache.ValidateKey(key); err != nil {
		return err
	}

	var s string
	err := tx.
		StmtContext(tx.ctx, tx.cn.stmt.delKey).
		QueryRowContext(tx.ctx, key, time.Now().UTC()).
		Scan(&s)
	if err == sql.ErrNoRows {
		return cache.ErrNotFound
	}
	return err
}

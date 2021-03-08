package common

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // this is specifically for PG
)

// Connect connects to a PG database.
func Connect(ctx context.Context, dsn string, versionField string, targetVersion int32, fs embed.FS) (*sql.DB, error) {
	schema := "postgres"
	if pos := strings.Index(dsn, "://"); pos > -1 {
		schema = dsn[:pos]
	}

	db, err := sql.Open(schema, dsn)
	if err != nil {
		return nil, err
	}

	if err := validateEncoding(ctx, db, "utf8"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := validateTimezone(ctx, db, "utc"); err != nil {
		_ = db.Close()
		return nil, err
	}

	version, err := schemaVersion(ctx, db, versionField)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	// Create schema if version is 0, migrate otherwise.
	if version == 0 {
		err = createSchema(ctx, db, fs)
	} else {
		err = migrateSchema(ctx, db, version, targetVersion)
	}
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

// validateEncoding makes sure database is set to specific encoding.
func validateEncoding(ctx context.Context, db *sql.DB, encoding string) error {
	var value string
	if err := db.QueryRowContext(ctx, `
		SELECT LOWER(pg_encoding_to_char(encoding))
		FROM pg_database
		WHERE datname = current_database()
	`).Scan(&value); err != nil {
		return fmt.Errorf("encoding check failed with %w", err)
	} else if strings.ToLower(value) != encoding {
		return fmt.Errorf("unexpected database encoding %q", value)
	}
	return nil
}

// validateTimezone makes sure database operates in a specific timezone.
func validateTimezone(ctx context.Context, db *sql.DB, timezone string) error {
	var value string
	if err := db.QueryRowContext(ctx, `
		SELECT current_setting('TIMEZONE') AS timezone
	`).Scan(&value); err != nil {
		return fmt.Errorf("timezone check failed with %w", err)
	} else if strings.ToLower(value) != timezone {
		return fmt.Errorf("unexpected database timezone %q", value)
	}
	return nil
}

// tableExists returns true if a table exists.
func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var value string
	err := db.QueryRowContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_name = $1
	`, table).Scan(&value)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("table check failed with %w", err)
	}
	return true, nil
}

// schemaVersion returns the stored schema version.
func schemaVersion(ctx context.Context, db *sql.DB, field string) (version int32, err error) {
	if ok, err := tableExists(ctx, db, "metainfo"); err != nil {
		return 0, err
	} else if !ok {
		return 0, nil
	}

	if err = db.QueryRowContext(ctx, `
		SELECT COALESCE(value::int, 0) AS version
		FROM metainfo
		WHERE name = $1
	`, field).Scan(&version); err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("schema check failed with %w", err)
	}
	return
}

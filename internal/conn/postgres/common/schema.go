package common

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

func createSchema(ctx context.Context, db *sql.DB, fs embed.FS) error {
	rawSQL, err := fs.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("schema creation failed with %w", err)
	}

	_, err = db.ExecContext(ctx, string(rawSQL))
	if err != nil {
		return fmt.Errorf("schema creation failed with %w", err)
	}

	return nil
}

func migrateSchema(_ context.Context, _ *sql.DB, current, target int32) error {
	if current == target {
		return nil
	}
	return fmt.Errorf("schema migration is not implemented")
}

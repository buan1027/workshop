package database

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed demo_seed.sql
var demoSeedSQL string

func ResetDemoData(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, demoSeedSQL); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

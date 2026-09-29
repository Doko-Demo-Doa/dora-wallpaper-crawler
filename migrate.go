package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS wallpapers (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		file_name varchar(2048) NOT NULL,
		url varchar(2048) NOT NULL,
		original_url varchar(2048) NOT NULL,
		is_mobile smallint NOT NULL,
		created_at date NOT NULL,
		updated_at date
	)`)
	if err != nil {
		return fmt.Errorf("migrate wallpapers: %w", err)
	}
	return nil
}

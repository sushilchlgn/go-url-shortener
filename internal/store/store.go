// Package store handles the Postgres connection and queries for the
// URL shortener. It expects a DATABASE_URL env var (works with Netlify DB,
// Neon, Supabase, or any standard Postgres connection string).
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // pure-Go postgres driver, no cgo
)

var schema = `
CREATE TABLE IF NOT EXISTS links (
	code TEXT PRIMARY KEY,
	url TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	clicks INT NOT NULL DEFAULT 0
);`

// Open connects to Postgres using DATABASE_URL and ensures the table exists.
func Open(ctx context.Context) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("ensure schema: %w", err)
	}
	return db, nil
}

// Insert saves a new short code -> url mapping. Returns an error on
// primary-key collision so the caller can retry with a new code.
func Insert(ctx context.Context, db *sql.DB, code, url string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO links (code, url) VALUES ($1, $2)`, code, url)
	return err
}

// Resolve looks up the target URL for a code and increments its click count.
func Resolve(ctx context.Context, db *sql.DB, code string) (string, error) {
	var url string
	err := db.QueryRowContext(ctx,
		`SELECT url FROM links WHERE code = $1`, code).Scan(&url)
	if err != nil {
		return "", err
	}
	_, _ = db.ExecContext(ctx,
		`UPDATE links SET clicks = clicks + 1 WHERE code = $1`, code)
	return url, nil
}

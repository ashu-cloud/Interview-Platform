package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("failed to parse DATABASE_URL: %v", err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err)
	}
	fmt.Println("Connected to database successfully.")

	// Create schema_migrations table if not exists
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT false
		);
	`)
	if err != nil {
		log.Fatalf("failed to create schema_migrations table: %v", err)
	}

	// Read migration versions
	var currentVersion int64 = 0
	_ = pool.QueryRow(ctx, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&currentVersion)
	fmt.Printf("Current schema version in database: %d\n", currentVersion)

	migrationFiles := []struct {
		version int64
		file    string
	}{
		{1, "migrations/000001_init_schema.up.sql"},
		{2, "migrations/000002_rebrand.up.sql"},
		{3, "migrations/000003_foundation.up.sql"},
		{4, "migrations/000004_evaluation_grounding.up.sql"},
	}

	for _, m := range migrationFiles {
		if m.version <= currentVersion {
			fmt.Printf("Skipping migration %d (%s) - already applied\n", m.version, filepath.Base(m.file))
			continue
		}

		sqlBytes, err := os.ReadFile(m.file)
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", m.file, err)
		}

		sql := string(sqlBytes)
		fmt.Printf("Applying migration %d (%s)...\n", m.version, filepath.Base(m.file))

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("failed to begin transaction: %v", err)
		}

		if _, err := tx.Exec(ctx, sql); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("failed executing migration %d (%s): %v", m.version, m.file, err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES ($1, false) ON CONFLICT (version) DO UPDATE SET dirty = false", m.version)
		if err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("failed to update schema_migrations for version %d: %v", m.version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("failed to commit migration %d: %v", m.version, err)
		}
		fmt.Printf("Successfully applied migration %d (%s)\n", m.version, filepath.Base(m.file))
	}

	fmt.Println("All database migrations applied successfully!")
}

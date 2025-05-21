package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	ID   int
	Up   string
	Down string
}

// RunMigrations runs all pending migrations
func RunMigrations(ctx context.Context, db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied := make(map[int]bool)
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Read migration files
	files, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		// Extract version number from filename (e.g., 0001_initial.up.sql -> 1)
		versionStr := strings.Split(name, "_")[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("invalid migration filename format: %s", name)
		}

		// Skip already applied migrations
		if applied[version] {
			continue
		}

		// Read up migration
		upSQL, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("failed to read up migration %s: %w", name, err)
		}

		// Read down migration if it exists
		downName := strings.Replace(name, ".up.sql", ".down.sql", 1)
		downSQL, _ := fs.ReadFile(migrationsFS, downName) // Ignore error if down migration doesn't exist

		migrations = append(migrations, Migration{
			ID:   version,
			Up:   string(upSQL),
			Down: string(downSQL),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	// Apply migrations in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range migrations {
		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.ID, err)
		}

		if _, err := tx.ExecContext(ctx, 
			"INSERT INTO schema_migrations (version) VALUES ($1)", migration.ID); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.ID, err)
		}

		fmt.Printf("Applied migration %d\n", migration.ID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migrations: %w", err)
	}

	return nil
}

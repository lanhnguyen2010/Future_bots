package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ErrNilDB is returned when a nil database handle is provided.
var ErrNilDB = errors.New("db: database handle is nil")

// ErrNilFS is returned when a nil filesystem is provided.
var ErrNilFS = errors.New("db: migrations filesystem is nil")

// Run applies all *.up.sql files located under dir within filesystem to the provided database handle.
// Migrations are executed in ascending version order based on the numeric prefix in the filename
// (e.g. 0001_init.up.sql). Executed migrations are tracked in the schema_migrations table using the
// same schema as the golang-migrate CLI so operators can mix automated and manual workflows.
func Run(ctx context.Context, database *sql.DB, filesystem fs.FS, dir string) error {
	if database == nil {
		return ErrNilDB
	}
	if filesystem == nil {
		return ErrNilFS
	}

	if dir == "" {
		dir = "."
	}

	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("db: ping connection: %w", err)
	}

	if err := ensureSchemaTable(ctx, database); err != nil {
		return err
	}

	applied, err := fetchApplied(ctx, database)
	if err != nil {
		return err
	}

	migrations, err := loadMigrations(filesystem, dir)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; ok {
			continue
		}

		if err := applyMigration(ctx, database, migration); err != nil {
			return err
		}
	}

	return nil
}

// RunFromDSN opens a database connection using the provided driver and DSN before executing migrations.
func RunFromDSN(ctx context.Context, driverName, dsn string, filesystem fs.FS, dir string) error {
	if driverName == "" {
		return errors.New("db: driver name must not be empty")
	}
	if dsn == "" {
		return errors.New("db: dsn must not be empty")
	}

	database, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("db: open connection: %w", err)
	}
	defer database.Close()

	return Run(ctx, database, filesystem, dir)
}

type migrationFile struct {
	Version int
	Name    string
	SQL     string
}

func ensureSchemaTable(ctx context.Context, database *sql.DB) error {
	const query = `CREATE TABLE IF NOT EXISTS schema_migrations (
        version BIGINT PRIMARY KEY,
        dirty BOOLEAN NOT NULL DEFAULT FALSE,
        applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`
	if _, err := database.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("db: ensure schema_migrations table: %w", err)
	}
	return nil
}

func fetchApplied(ctx context.Context, database *sql.DB) (map[int]struct{}, error) {
	rows, err := database.QueryContext(ctx, "SELECT version FROM schema_migrations WHERE dirty = FALSE")
	if err != nil {
		return nil, fmt.Errorf("db: fetch applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]struct{})
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("db: scan applied migration: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: iterate applied migrations: %w", err)
	}

	return applied, nil
}

func loadMigrations(filesystem fs.FS, dir string) ([]migrationFile, error) {
	entries := make([]migrationFile, 0)
	err := fs.WalkDir(filesystem, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			return nil
		}

		version, err := parseVersion(name)
		if err != nil {
			return fmt.Errorf("db: parse migration version for %s: %w", path, err)
		}

		contents, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return fmt.Errorf("db: read migration %s: %w", path, err)
		}

		entries = append(entries, migrationFile{
			Version: version,
			Name:    filepath.Base(path),
			SQL:     string(contents),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Version == entries[j].Version {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Version < entries[j].Version
	})

	return entries, nil
}

func parseVersion(name string) (int, error) {
	trimmed := strings.TrimSuffix(name, ".up.sql")
	idx := strings.IndexFunc(trimmed, func(r rune) bool { return r < '0' || r > '9' })
	if idx == -1 {
		idx = len(trimmed)
	}
	if idx == 0 {
		return 0, errors.New("missing numeric prefix")
	}
	value, err := strconv.Atoi(trimmed[:idx])
	if err != nil {
		return 0, err
	}
	return value, nil
}

func applyMigration(ctx context.Context, database *sql.DB, migration migrationFile) error {
	tx, err := database.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("db: begin migration %s: %w", migration.Name, err)
	}

	started := time.Now().UTC()

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, dirty, applied_at)
VALUES ($1, TRUE, $2)
ON CONFLICT (version) DO UPDATE SET dirty = TRUE, applied_at = EXCLUDED.applied_at`, migration.Version, started); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("db: mark migration dirty %s: %w", migration.Name, err)
	}

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("db: execute migration %s: %w", migration.Name, err)
	}

	finished := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE schema_migrations SET dirty = FALSE, applied_at = $2 WHERE version = $1`, migration.Version, finished); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("db: record migration %s: %w", migration.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit migration %s: %w", migration.Name, err)
	}

	return nil
}

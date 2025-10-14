package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestRunAppliesMigrationsFromFilesystem(t *testing.T) {
	driver := &stubDriver{}
	name := registerDriver(t, driver)

	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	defer database.Close()

	fs := fstest.MapFS{
		"migrations/0001_init.up.sql":     {Data: []byte("CREATE TABLE bots(id TEXT PRIMARY KEY);")},
		"migrations/0001_init.down.sql":   {Data: []byte("DROP TABLE bots;")},
		"migrations/0002_add_name.up.sql": {Data: []byte("ALTER TABLE bots ADD COLUMN name TEXT;")},
	}

	if err := Run(context.Background(), database, fs, "migrations"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(driver.execs) == 0 || driver.execs[0] != "CREATE TABLE schema_migrations" {
		t.Fatalf("expected schema table creation, got %v", driver.execs)
	}

	if len(driver.txExecs) != 6 { // insert dirty + migration + update for each file
		t.Fatalf("expected 6 transactional execs, got %d (%v)", len(driver.txExecs), driver.txExecs)
	}

	if driver.commits != 2 {
		t.Fatalf("expected 2 commits, got %d", driver.commits)
	}
}

func TestRunSkipsAppliedMigrations(t *testing.T) {
	driver := &stubDriver{applied: []int{1}}
	name := registerDriver(t, driver)

	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	defer database.Close()

	fs := fstest.MapFS{
		"migrations/0001_init.up.sql": {Data: []byte("CREATE TABLE bots(id TEXT PRIMARY KEY);")},
	}

	if err := Run(context.Background(), database, fs, "migrations"); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(driver.txExecs) != 0 {
		t.Fatalf("expected no transactional execs, got %v", driver.txExecs)
	}
}

func TestRunFromDSNRejectsEmptyInputs(t *testing.T) {
	ctx := context.Background()
	if err := RunFromDSN(ctx, "", "dsn", fstest.MapFS{}, "."); err == nil {
		t.Fatal("expected error for empty driver")
	}
	if err := RunFromDSN(ctx, "pgx", "", fstest.MapFS{}, "."); err == nil {
		t.Fatal("expected error for empty dsn")
	}
}

func TestRunRequiresFilesystem(t *testing.T) {
	driver := &stubDriver{}
	name := registerDriver(t, driver)

	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	defer database.Close()

	if err := Run(context.Background(), database, nil, "migrations"); !errors.Is(err, ErrNilFS) {
		t.Fatalf("expected ErrNilFS, got %v", err)
	}
}

func TestRunRequiresNumericPrefix(t *testing.T) {
	driver := &stubDriver{}
	name := registerDriver(t, driver)

	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}
	defer database.Close()

	fs := fstest.MapFS{
		"migrations/init.up.sql": {Data: []byte("SELECT 1;")},
	}

	if err := Run(context.Background(), database, fs, "migrations"); err == nil {
		t.Fatal("expected error for missing numeric prefix")
	}
}

func TestRunRejectsNilDB(t *testing.T) {
	if err := Run(context.Background(), nil, fstest.MapFS{}, "migrations"); !errors.Is(err, ErrNilDB) {
		t.Fatalf("expected ErrNilDB, got %v", err)
	}
}

type stubDriver struct {
	applied   []int
	execs     []string
	txExecs   []string
	queries   []string
	commits   int
	rollbacks int
	failOn    string
	failErr   error
}

type stubConn struct {
	driver *stubDriver
	inTx   bool
}

type stubTx struct {
	conn *stubConn
}

type stubRows struct {
	values []int
	idx    int
}

func registerDriver(t *testing.T, d *stubDriver) string {
	name := fmt.Sprintf("stub-%s-%d", t.Name(), time.Now().UnixNano())
	sql.Register(name, d)
	return name
}

func (d *stubDriver) Open(string) (driver.Conn, error) {
	return &stubConn{driver: d}, nil
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c *stubConn) Close() error {
	return nil
}

func (c *stubConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *stubConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if c.inTx {
		return nil, errors.New("transaction already active")
	}
	c.inTx = true
	return &stubTx{conn: c}, nil
}

func (c *stubConn) Ping(ctx context.Context) error {
	return nil
}

func (c *stubConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	normalized := normalizeQuery(query)
	if c.inTx {
		c.driver.txExecs = append(c.driver.txExecs, normalized)
	} else {
		c.driver.execs = append(c.driver.execs, normalized)
	}

	if c.driver.failOn != "" && strings.Contains(normalized, c.driver.failOn) {
		return nil, c.driver.failErr
	}

	return driver.RowsAffected(1), nil
}

func (c *stubConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	named := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: arg}
	}
	return c.ExecContext(context.Background(), query, named)
}

func (c *stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	normalized := normalizeQuery(query)
	c.driver.queries = append(c.driver.queries, normalized)
	if strings.Contains(strings.ToLower(normalized), "select version from schema_migrations") {
		values := append([]int(nil), c.driver.applied...)
		return &stubRows{values: values, idx: -1}, nil
	}
	return nil, fmt.Errorf("unsupported query: %s", query)
}

func (c *stubConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	named := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: arg}
	}
	return c.QueryContext(context.Background(), query, named)
}

func (tx *stubTx) Commit() error {
	if !tx.conn.inTx {
		return errors.New("commit with no active transaction")
	}
	tx.conn.inTx = false
	tx.conn.driver.commits++
	return nil
}

func (tx *stubTx) Rollback() error {
	if !tx.conn.inTx {
		return errors.New("rollback with no active transaction")
	}
	tx.conn.inTx = false
	tx.conn.driver.rollbacks++
	return nil
}

func (r *stubRows) Columns() []string {
	return []string{"version"}
}

func (r *stubRows) Close() error {
	return nil
}

func (r *stubRows) Next(dest []driver.Value) error {
	r.idx++
	if r.idx >= len(r.values) {
		return io.EOF
	}
	dest[0] = int64(r.values[r.idx])
	return nil
}

func normalizeQuery(query string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))
	for strings.Contains(trimmed, "  ") {
		trimmed = strings.ReplaceAll(trimmed, "  ", " ")
	}
	switch {
	case strings.HasPrefix(trimmed, "CREATE TABLE IF NOT EXISTS schema_migrations"):
		return "CREATE TABLE schema_migrations"
	case strings.HasPrefix(trimmed, "INSERT INTO schema_migrations"):
		return "INSERT schema_migrations"
	case strings.HasPrefix(trimmed, "UPDATE schema_migrations SET dirty = FALSE"):
		return "UPDATE schema_migrations"
	default:
		return trimmed
	}
}

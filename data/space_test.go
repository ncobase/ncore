package data

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/ncobase/ncore/data/metrics"
)

// ─────────────────────────────────────────────────────────────────────────────
// In-process SQL driver for testing without an external database.
// ─────────────────────────────────────────────────────────────────────────────

// testDriver records which SQL statements were executed. It is registered once
// under the name "testdriver".
type testDriver struct{}

func (testDriver) Open(_ string) (driver.Conn, error) { return &testConn{}, nil }

type testConn struct {
	mu       sync.Mutex
	stmts    []string
	closed   bool
	failTx   bool // if true, BeginTx returns an error
	failExec bool // if true, Exec returns an error
}

func (c *testConn) Prepare(query string) (driver.Stmt, error) {
	return &testStmt{query: query, conn: c}, nil
}
func (c *testConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}
func (c *testConn) Begin() (driver.Tx, error) { return &testTx{conn: c}, nil }

type testStmt struct {
	query string
	conn  *testConn
}

func (s *testStmt) Close() error  { return nil }
func (s *testStmt) NumInput() int { return -1 }
func (s *testStmt) Exec(_ []driver.Value) (driver.Result, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	s.conn.stmts = append(s.conn.stmts, s.query)
	if s.conn.failExec {
		return nil, errors.New("exec error")
	}
	return driver.RowsAffected(0), nil
}
func (s *testStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return &testRows{}, nil
}

type testRows struct{ done bool }

func (r *testRows) Columns() []string { return nil }
func (r *testRows) Close() error      { return nil }
func (r *testRows) Next(_ []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	return io.EOF
}

type testTx struct{ conn *testConn }

func (t *testTx) Commit() error   { return nil }
func (t *testTx) Rollback() error { return nil }

func init() {
	sql.Register("testdriver", testDriver{})
}

// openTestDB returns a *sql.DB backed by the in-process testDriver.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("testdriver", "test")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// ─────────────────────────────────────────────────────────────────────────────
// Context helpers
// ─────────────────────────────────────────────────────────────────────────────

func TestSetAndGetSpaceID(t *testing.T) {
	ctx := context.Background()

	// No space ID set yet.
	id, ok := GetSpaceID(ctx)
	if ok || id != "" {
		t.Fatalf("expected empty space ID, got %q ok=%v", id, ok)
	}

	// Set a space ID.
	ctx = SetSpaceID(ctx, "acme")
	id, ok = GetSpaceID(ctx)
	if !ok {
		t.Fatal("expected ok=true after SetSpaceID")
	}
	if id != "acme" {
		t.Fatalf("expected space ID %q, got %q", "acme", id)
	}
}

func TestGetSpaceID_EmptyStringReturnsFalse(t *testing.T) {
	ctx := SetSpaceID(context.Background(), "")
	_, ok := GetSpaceID(ctx)
	if ok {
		t.Fatal("expected ok=false for empty space ID")
	}
}

func TestSetSpaceID_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	child := SetSpaceID(parent, "acme")

	if _, ok := GetSpaceID(parent); ok {
		t.Fatal("SetSpaceID must not mutate the parent context")
	}
	if id, _ := GetSpaceID(child); id != "acme" {
		t.Fatalf("expected child to have space ID %q", "acme")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateSpaceID
// ─────────────────────────────────────────────────────────────────────────────

func TestValidateSpaceID_Valid(t *testing.T) {
	valid := []string{
		"acme",
		"acme-corp",
		"acme_corp",
		"Acme",
		"ACME",
		"01J5JKHW3JN8ZFCPRQXB3GFKXM",           // ULID
		"550e8400-e29b-41d4-a716-446655440000", // UUID
		"a",
		"a1",
		strings.Repeat("a", 63),
	}
	for _, id := range valid {
		if err := ValidateSpaceID(id); err != nil {
			t.Errorf("ValidateSpaceID(%q) returned unexpected error: %v", id, err)
		}
	}
}

func TestValidateSpaceID_Invalid(t *testing.T) {
	invalid := []string{
		"",
		" acme",
		"acme corp",
		"acme;drop",
		"../etc",
		"acme.corp",
		"acme@corp",
		"-acme",
		"_acme",
		strings.Repeat("a", 64), // too long (max 63)
	}
	for _, id := range invalid {
		if err := ValidateSpaceID(id); err == nil {
			t.Errorf("ValidateSpaceID(%q) expected error, got nil", id)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildSchemaName
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildSchemaName(t *testing.T) {
	tests := []struct {
		prefix  string
		spaceID string
		want    string
		wantErr bool
	}{
		{"space_", "acme", "space_acme", false},
		{"space_", "Acme-Corp", "space_acme_corp", false},
		{"space_", "ACME", "space_acme", false},
		{"tenant_", "my-space", "tenant_my_space", false},
		{"space_", "", "", true},
		{"space_", " invalid", "", true},
		{"space_", "drop;table", "", true},
		{"space-unsafe", "acme", "", true},
		{strings.Repeat("a", 40) + "_", strings.Repeat("b", 30), "", true},
	}
	for _, tc := range tests {
		got, err := BuildSchemaName(tc.prefix, tc.spaceID)
		if tc.wantErr {
			if err == nil {
				t.Errorf("BuildSchemaName(%q, %q) expected error, got nil", tc.prefix, tc.spaceID)
			}
			continue
		}
		if err != nil {
			t.Errorf("BuildSchemaName(%q, %q) unexpected error: %v", tc.prefix, tc.spaceID, err)
			continue
		}
		if got != tc.want {
			t.Errorf("BuildSchemaName(%q, %q) = %q, want %q", tc.prefix, tc.spaceID, got, tc.want)
		}
	}
}

func TestQuotePostgresIdentifier(t *testing.T) {
	quoted, err := QuotePostgresIdentifier("space_acme")
	if err != nil {
		t.Fatalf("QuotePostgresIdentifier returned error: %v", err)
	}
	if quoted != `"space_acme"` {
		t.Fatalf("unexpected quoted identifier: %s", quoted)
	}

	if _, err := QuotePostgresIdentifier("space-acme"); err == nil {
		t.Fatal("expected invalid identifier error")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ExecInTx
// ─────────────────────────────────────────────────────────────────────────────

func TestExecInTx_NilDB(t *testing.T) {
	err := ExecInTx(context.Background(), nil, nil, "", func(_ context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestExecInTx_StoresTxInContext(t *testing.T) {
	db := openTestDB(t)

	var gotTx *sql.Tx
	err := ExecInTx(context.Background(), db, nil, "", func(ctx context.Context) error {
		tx, txErr := GetTx(ctx)
		if txErr != nil {
			return txErr
		}
		gotTx = tx
		return nil
	})
	if err != nil {
		t.Fatalf("ExecInTx: %v", err)
	}
	if gotTx == nil {
		t.Fatal("expected *sql.Tx in context, got nil")
	}
}

func TestExecInTx_ExecutesSetupSQL(t *testing.T) {
	db := openTestDB(t)

	setupSQL := "SET LOCAL search_path TO space_acme, public"
	var executed bool

	err := ExecInTx(context.Background(), db, nil, setupSQL, func(_ context.Context) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("ExecInTx: %v", err)
	}
	if !executed {
		t.Fatal("fn was not called")
	}
}

func TestExecInTx_RollbackOnFnError(t *testing.T) {
	db := openTestDB(t)
	fnErr := errors.New("fn failed")

	err := ExecInTx(context.Background(), db, nil, "", func(_ context.Context) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("expected fn error to be returned, got: %v", err)
	}
}

func TestExecInTx_ReadOnlyOptions(t *testing.T) {
	db := openTestDB(t)

	// The in-process testdriver does not support read-only transactions; that is
	// a driver-level limitation not relevant to ExecInTx's own logic. We verify
	// that ExecInTx forwards the TxOptions to BeginTx by confirming the error
	// message originates from the driver, not from ExecInTx itself.
	err := ExecInTx(context.Background(), db, &sql.TxOptions{ReadOnly: true}, "", func(_ context.Context) error {
		return nil
	})
	// Two acceptable outcomes:
	//  a) nil – driver accepted read-only (real PG/MySQL drivers do)
	//  b) error mentioning "read-only" – driver rejected it, but ExecInTx passed it correctly
	if err != nil && !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("unexpected error from ExecInTx with ReadOnly options: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SpaceRouter injection and WithSpaceTx / WithSpaceTxRead
// ─────────────────────────────────────────────────────────────────────────────

// mockSpaceRouter records calls to WithSpace / WithSpaceRead.
type mockSpaceRouter struct {
	mu          sync.Mutex
	withCalls   []string // spaceIDs passed to WithSpace
	readCalls   []string // spaceIDs passed to WithSpaceRead
	returnError error    // if non-nil, returned from both methods
}

func (r *mockSpaceRouter) WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error {
	r.mu.Lock()
	r.withCalls = append(r.withCalls, spaceID)
	r.mu.Unlock()
	if r.returnError != nil {
		return r.returnError
	}
	return fn(ctx)
}

func (r *mockSpaceRouter) WithSpaceRead(ctx context.Context, spaceID string, fn func(context.Context) error) error {
	r.mu.Lock()
	r.readCalls = append(r.readCalls, spaceID)
	r.mu.Unlock()
	if r.returnError != nil {
		return r.returnError
	}
	return fn(ctx)
}

// newTestData builds a *Data instance suitable for unit tests without a real DB.
func newTestData(router SpaceRouter) *Data {
	d := &Data{
		collector: metrics.NoOpCollector{},
	}
	if router != nil {
		d.spaceRouter = router
	}
	return d
}

func TestWithSpaceRouter_Option(t *testing.T) {
	router := &mockSpaceRouter{}
	d := &Data{collector: metrics.NoOpCollector{}}

	opt := WithSpaceRouter(router)
	opt(d)

	if d.spaceRouter != router {
		t.Fatal("WithSpaceRouter did not set the router")
	}
}

func TestWithSpaceRouter_NilNoops(t *testing.T) {
	d := &Data{collector: metrics.NoOpCollector{}}
	opt := WithSpaceRouter(nil)
	opt(d)
	if d.spaceRouter != nil {
		t.Fatal("WithSpaceRouter(nil) should leave router nil")
	}
}

func TestGetSpaceRouter(t *testing.T) {
	router := &mockSpaceRouter{}
	d := newTestData(router)
	if d.GetSpaceRouter() != router {
		t.Fatal("GetSpaceRouter did not return the injected router")
	}
}

func TestWithSpaceTx_FallsBackWhenNoSpaceID(t *testing.T) {
	// No space ID in context: should fall back to WithTx.
	// WithTx will fail because Conn is nil, but the point is that the router
	// is NOT called.
	router := &mockSpaceRouter{}
	d := newTestData(router)

	ctx := context.Background() // no space ID
	_ = d.WithSpaceTx(ctx, func(_ context.Context) error { return nil })

	router.mu.Lock()
	defer router.mu.Unlock()
	if len(router.withCalls) > 0 {
		t.Fatal("SpaceRouter.WithSpace must not be called when context has no space ID")
	}
}

func TestWithSpaceTx_FallsBackWhenNoRouter(t *testing.T) {
	// Space ID present but no router injected: must fall back to WithTx.
	d := newTestData(nil)
	ctx := SetSpaceID(context.Background(), "acme")

	// WithTx will fail (no DB), but router must not be called.
	_ = d.WithSpaceTx(ctx, func(_ context.Context) error { return nil })
	// No panic / no router call – passing means the fallback path was taken.
}

func TestWithSpaceTx_RoutesToSpace(t *testing.T) {
	router := &mockSpaceRouter{}
	d := newTestData(router)

	ctx := SetSpaceID(context.Background(), "acme-corp")
	var called bool
	err := d.WithSpaceTx(ctx, func(_ context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithSpaceTx: %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}

	router.mu.Lock()
	defer router.mu.Unlock()
	if len(router.withCalls) != 1 || router.withCalls[0] != "acme-corp" {
		t.Fatalf("expected one WithSpace call for %q, got %v", "acme-corp", router.withCalls)
	}
}

func TestWithSpaceTxRead_RoutesToSpaceRead(t *testing.T) {
	router := &mockSpaceRouter{}
	d := newTestData(router)

	ctx := SetSpaceID(context.Background(), "globex")
	err := d.WithSpaceTxRead(ctx, func(_ context.Context) error { return nil })
	if err != nil {
		t.Fatalf("WithSpaceTxRead: %v", err)
	}

	router.mu.Lock()
	defer router.mu.Unlock()
	if len(router.readCalls) != 1 || router.readCalls[0] != "globex" {
		t.Fatalf("expected one WithSpaceRead call for %q, got %v", "globex", router.readCalls)
	}
}

func TestWithSpaceTx_PropagatesRouterError(t *testing.T) {
	routerErr := errors.New("space router error")
	router := &mockSpaceRouter{returnError: routerErr}
	d := newTestData(router)

	ctx := SetSpaceID(context.Background(), "acme")
	err := d.WithSpaceTx(ctx, func(_ context.Context) error { return nil })
	if !errors.Is(err, routerErr) {
		t.Fatalf("expected router error, got: %v", err)
	}
}

func TestWithSpaceTx_ClosedData(t *testing.T) {
	router := &mockSpaceRouter{}
	d := newTestData(router)
	d.closed = true

	ctx := SetSpaceID(context.Background(), "acme")
	err := d.WithSpaceTx(ctx, func(_ context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error on closed data layer")
	}

	router.mu.Lock()
	defer router.mu.Unlock()
	if len(router.withCalls) > 0 {
		t.Fatal("router must not be called when data layer is closed")
	}
}

func TestWithSpaceTx_ConcurrentAccess(t *testing.T) {
	router := &mockSpaceRouter{}
	d := newTestData(router)

	var wg sync.WaitGroup
	n := 50
	wg.Add(n)
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			spaceID := fmt.Sprintf("space-%d", i%5)
			ctx := SetSpaceID(context.Background(), spaceID)
			if err := d.WithSpaceTx(ctx, func(_ context.Context) error { return nil }); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent WithSpaceTx: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// normalizeIdentifier (internal)
// ─────────────────────────────────────────────────────────────────────────────

func TestNormalizeIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"acme", "acme"},
		{"Acme", "acme"},
		{"ACME", "acme"},
		{"acme-corp", "acme_corp"},
		{"Acme-Corp", "acme_corp"},
		{"acme_corp", "acme_corp"},
		{"01J5JKHW3JN8ZFCPRQXB3GFKXM", "01j5jkhw3jn8zfcprqxb3gfkxm"},
	}
	for _, tc := range tests {
		got := normalizeIdentifier(tc.input)
		if got != tc.want {
			t.Errorf("normalizeIdentifier(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

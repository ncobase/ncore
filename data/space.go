// Package data provides Space-level database isolation for multi-tenant applications.
//
// # Space Isolation Overview
//
// Space isolation allows each logical space (tenant) to operate within its own
// database namespace, providing data separation at the database level rather than
// relying solely on application-layer checks.
//
// The isolation is implemented through the [SpaceRouter] interface. ncore provides
// the plumbing and contract; the routing strategy is application-defined and injected
// via [WithSpaceRouter].
//
// # Isolation Strategies
//
// PostgreSQL – schema-per-space:
//
//	Each Space gets its own PostgreSQL schema. The search_path is set per-transaction
//	using SET LOCAL, which is safe with connection pools because the setting is
//	automatically reverted when the transaction ends.
//
//	  BEGIN;
//	  SET LOCAL search_path TO "space_acme", "public";
//	  -- all queries run in the space_acme schema
//	  COMMIT; -- connection returned to pool with original search_path restored
//
// MySQL – database-per-space:
//
//	Each Space gets its own MySQL database. Application-managed per-space connection
//	pools are used, typically with LRU eviction to cap total pool count.
//
// # Security
//
// Schema and database names are constructed from space identifiers. Because these
// identifiers cannot be parameterized via SQL placeholders, [ValidateSpaceID]
// enforces a strict allowlist regex to prevent injection attacks. Always call it
// before incorporating a space identifier into any SQL string.
//
// # Backward Compatibility
//
// All existing APIs ([Data.WithTx], [Data.GetMasterDB], etc.) are unchanged.
// If no [SpaceRouter] is injected, [Data.WithSpaceTx] and [Data.WithSpaceTxRead]
// transparently fall back to the ordinary transaction methods.
//
// # Usage
//
//	// 1. Implement SpaceRouter in your application layer.
//	router := myapp.NewPostgresSchemaRouter(masterDB, slaveDB)
//
//	// 2. Initialize Data and inject the router.
//	d, cleanup, err := data.New(cfg)
//	data.WithSpaceRouter(router)(d)
//
//	// 3. Set space ID in HTTP middleware.
//	ctx = data.SetSpaceID(ctx, space.Slug)
//
//	// 4. Execute within the space namespace in service / repository.
//	err = d.WithSpaceTx(ctx, func(ctx context.Context) error {
//	    tx, _ := data.GetTx(ctx) // *sql.Tx scoped to the space's schema or database
//	    return repo.Create(ctx, entity)
//	})
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Context helpers
// ─────────────────────────────────────────────────────────────────────────────

// ContextKeySpaceID is the context key used to store and retrieve the current
// Space identifier. Use [SetSpaceID] and [GetSpaceID] instead of reading this
// key directly.
const ContextKeySpaceID ContextKey = "space_id"

// SetSpaceID stores the Space identifier in ctx and returns the new context.
// It is typically called from an HTTP middleware after the space has been
// validated for the authenticated user.
//
//	ctx = data.SetSpaceID(r.Context(), space.Slug)
func SetSpaceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ContextKeySpaceID, id)
}

// GetSpaceID retrieves the Space identifier previously set by [SetSpaceID].
// The boolean return value is false when no space ID is present or when it is
// the empty string.
func GetSpaceID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ContextKeySpaceID).(string)
	return id, ok && id != ""
}

// ─────────────────────────────────────────────────────────────────────────────
// SpaceRouter interface
// ─────────────────────────────────────────────────────────────────────────────

// SpaceRouter is the core abstraction for Space-level database routing.
//
// Application code implements this interface and injects it into [Data] via
// [WithSpaceRouter]. ncore ships no concrete implementation so that no extra
// dependencies are imposed on projects that do not need multi-tenancy.
//
// # Implementation contract
//
//   - Space context MUST be scoped to the transaction (e.g. SET LOCAL in
//     PostgreSQL). Never use a session-level SET, which would leak the context
//     to the next caller that borrows the same connection from the pool.
//   - The context passed to fn MUST carry the active *sql.Tx so that callers
//     can retrieve it with [GetTx] and integrate it with their ORM.
//   - The implementation owns the full Begin / Commit / Rollback lifecycle of
//     the transaction.
//
// See the package-level [ExecInTx] helper, which correctly implements this
// contract and can be composed into custom router implementations.
type SpaceRouter interface {
	// WithSpace executes fn inside a read-write transaction scoped to the
	// namespace of the given space. spaceID is the application-defined
	// identifier (e.g. a slug or UUID) that the router maps to a schema or
	// database name.
	WithSpace(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error

	// WithSpaceRead executes fn inside a read-only transaction scoped to the
	// namespace of the given space. Implementations may direct the query to a
	// read replica for this variant.
	WithSpaceRead(ctx context.Context, spaceID string, fn func(ctx context.Context) error) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Data integration
// ─────────────────────────────────────────────────────────────────────────────

// WithSpaceRouter returns an [Option] that injects a [SpaceRouter] into [Data].
// Apply it after constructing Data for projects that require Space-level isolation:
//
//	d, cleanup, err := data.New(cfg)
//	data.WithSpaceRouter(myRouter)(d)
//
// Projects that do not need multi-tenancy simply omit this option; all
// Space-aware methods fall back to ordinary transactions automatically.
func WithSpaceRouter(router SpaceRouter) Option {
	return func(d *Data) {
		if router != nil {
			d.spaceRouter = router
		}
	}
}

// WithSpaceTx executes fn inside a read-write transaction scoped to the Space
// identified in ctx.
//
// Routing behaviour:
//   - ctx contains a Space ID AND a [SpaceRouter] has been injected →
//     delegates to [SpaceRouter.WithSpace]; the transaction operates in the
//     Space's dedicated schema or database.
//   - Otherwise → falls back to [Data.WithTx] (identical to the pre-isolation
//     behaviour, preserving backward compatibility).
//
// The *sql.Tx is available inside fn via [GetTx] regardless of which path is
// taken.
func (d *Data) WithSpaceTx(ctx context.Context, fn func(ctx context.Context) error) error {
	spaceID, hasSpace := GetSpaceID(ctx)

	d.mu.RLock()
	closed := d.closed
	router := d.spaceRouter
	collector := d.collector
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		collector.DBTransaction(err)
		return err
	}

	if !hasSpace || router == nil {
		return d.WithTx(ctx, fn)
	}

	start := time.Now()
	err := router.WithSpace(ctx, spaceID, fn)
	duration := time.Since(start)

	collector.DBQuery(duration, err)
	collector.DBTransaction(err)
	return err
}

// WithSpaceTxRead executes fn inside a read-only transaction scoped to the
// Space identified in ctx. It follows the same routing logic as [WithSpaceTx]
// but uses [SpaceRouter.WithSpaceRead] and falls back to [Data.WithTxRead].
func (d *Data) WithSpaceTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
	spaceID, hasSpace := GetSpaceID(ctx)

	d.mu.RLock()
	closed := d.closed
	router := d.spaceRouter
	collector := d.collector
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		collector.DBTransaction(err)
		return err
	}

	if !hasSpace || router == nil {
		return d.WithTxRead(ctx, fn)
	}

	start := time.Now()
	err := router.WithSpaceRead(ctx, spaceID, fn)
	duration := time.Since(start)

	collector.DBQuery(duration, err)
	collector.DBTransaction(err)
	return err
}

// GetSpaceRouter returns the injected [SpaceRouter], or nil if none was set.
// Useful for applications that need direct access to the router for lifecycle
// operations (e.g. schema provisioning).
func (d *Data) GetSpaceRouter() SpaceRouter {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.spaceRouter
}

// ─────────────────────────────────────────────────────────────────────────────
// Security: identifier validation
// ─────────────────────────────────────────────────────────────────────────────

// spaceIDPattern is the allowlist for space identifiers used in schema / db
// names. It accepts slugs (lowercase, digits, hyphens, underscores) as well as
// uppercase UUIDs / ULIDs. Applications must call [ValidateSpaceID] before
// incorporating any identifier into a schema or database name.
//
// Accepted examples:  "acme", "acme-corp", "acme_corp", "01J5JKHW3JN8ZFCPRQXB3GFKXM"
// Rejected examples:  "", " acme", "acme corp", "acme;drop table", "../etc"
const maxDatabaseIdentifierLength = 63

var (
	spaceIDPattern    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`)
	identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)
)

// ValidateSpaceID checks that id is safe for use as, or as a component of, a
// PostgreSQL schema name or MySQL database name.
//
// Because schema / database names cannot be parameterized in SQL, they must be
// sanitized before string interpolation. ValidateSpaceID enforces the allowlist
// regex [a-zA-Z0-9][a-zA-Z0-9_-]{0,62}.
//
// Callers should additionally convert the identifier to lowercase and replace
// hyphens with underscores before building the final schema or database name,
// since some databases treat these characters specially.
//
//	if err := data.ValidateSpaceID(slug); err != nil {
//	    return err
//	}
//	schema := "space_" + strings.ToLower(strings.ReplaceAll(slug, "-", "_"))
func ValidateSpaceID(id string) error {
	if id == "" {
		return errors.New("data: space id must not be empty")
	}
	if !spaceIDPattern.MatchString(id) {
		return fmt.Errorf(
			"data: space id %q is not a valid identifier; "+
				"must match ^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$",
			id,
		)
	}
	return nil
}

// ValidateIdentifier checks that an identifier component controlled by the
// application, such as a schema prefix or global schema name, is safe for SQL
// identifier use.
func ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("data: identifier must not be empty")
	}
	if !identifierPattern.MatchString(identifier) {
		return fmt.Errorf(
			"data: identifier %q is not valid; must match ^[a-zA-Z_][a-zA-Z0-9_]{0,62}$",
			identifier,
		)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Transaction utility
// ─────────────────────────────────────────────────────────────────────────────

// ExecInTx is a low-level helper intended for use inside [SpaceRouter]
// implementations. It:
//
//  1. Begins a transaction on db using opts (pass nil for read-write defaults).
//  2. If setupSQL is non-empty, executes it immediately after BEGIN. This is
//     the hook for statements like:
//     "SET LOCAL search_path TO space_acme, public"
//  3. Stores the *sql.Tx in ctx under [ContextKeyTransaction] so that callers
//     of fn can retrieve it with [GetTx].
//  4. Calls fn with the enriched context.
//  5. Commits on success or rolls back on error.
//
// The setupSQL string must NOT be constructed from untrusted user input.
// Always derive schema / database names through [ValidateSpaceID] first.
//
// Example PostgreSQL schema router usage:
//
//	func (r *PostgresSchemaRouter) WithSpace(ctx context.Context, spaceID string, fn func(context.Context) error) error {
//	    schema, err := r.schemaName(spaceID) // internally calls ValidateSpaceID
//	    if err != nil {
//	        return err
//	    }
//	    schemaIdent, err := data.QuotePostgresIdentifier(schema)
//	    if err != nil {
//	        return err
//	    }
//	    globalIdent, err := data.QuotePostgresIdentifier(r.globalSchema)
//	    if err != nil {
//	        return err
//	    }
//	    setup := fmt.Sprintf("SET LOCAL search_path TO %s, %s", schemaIdent, globalIdent)
//	    return data.ExecInTx(ctx, r.masterDB, nil, setup, fn)
//	}
func ExecInTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, setupSQL string, fn func(ctx context.Context) error) error {
	if db == nil {
		return errors.New("data: ExecInTx received nil *sql.DB")
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("data: begin transaction: %w", err)
	}

	if setupSQL != "" {
		if _, err = tx.ExecContext(ctx, setupSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("data: execute setup SQL: %w", err)
		}
	}

	// Store tx in context so downstream code (e.g. repositories) can access it
	// via GetTx(ctx) or integrate it with their ORM.
	txCtx := context.WithValue(ctx, ContextKeyTransaction, tx)

	fnErr := fn(txCtx)
	if fnErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("data: fn error: %w; rollback error: %v", fnErr, rbErr)
		}
		return fnErr
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("data: commit transaction: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Schema name helper
// ─────────────────────────────────────────────────────────────────────────────

// BuildSchemaName constructs a safe schema or database name by combining a
// validated prefix with a sanitized space identifier.
//
//   - prefix must be a safe identifier component (e.g. "space_").
//   - spaceID is validated via [ValidateSpaceID], then lowercased, and hyphens
//     are replaced with underscores to satisfy both PostgreSQL schema naming
//     rules and MySQL database naming conventions.
//
// Returns an error if prefix, spaceID, or the final name fails validation.
//
// Example:
//
//	name, err := data.BuildSchemaName("space_", "Acme-Corp")
//	// name == "space_acme_corp"
func BuildSchemaName(prefix, spaceID string) (string, error) {
	if err := ValidateIdentifier(prefix); err != nil {
		return "", err
	}
	if err := ValidateSpaceID(spaceID); err != nil {
		return "", err
	}
	// Normalize: lowercase + replace hyphens with underscores.
	normalized := normalizeIdentifier(spaceID)
	name := prefix + normalized
	if len(name) > maxDatabaseIdentifierLength {
		return "", fmt.Errorf("data: schema name %q exceeds %d bytes", name, maxDatabaseIdentifierLength)
	}
	return name, nil
}

// QuotePostgresIdentifier quotes a PostgreSQL identifier after validating it.
func QuotePostgresIdentifier(identifier string) (string, error) {
	if err := ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`, nil
}

// normalizeIdentifier converts an identifier to a form safe for use as a
// PostgreSQL schema name or MySQL database name component:
//   - lowercased
//   - hyphens replaced with underscores
func normalizeIdentifier(id string) string {
	out := make([]byte, len(id))
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A' // to lowercase
		} else if c == '-' {
			c = '_'
		}
		out[i] = c
	}
	return string(out)
}

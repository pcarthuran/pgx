// Package pgx is a PostgreSQL driver and toolkit for Go.
//
// The pgx driver is a low-level, high performance interface that exposes
// PostgreSQL-specific features such as LISTEN/NOTIFY and COPY. It also
// includes an adapter for the standard database/sql interface.
//
// # Connecting
//
// A connection can be established with [Connect] or by creating a connection
// pool with [pgxpool.New].
//
// # Query Interface
//
// pgx implements Query, QueryRow, and Exec with a similar interface to
// database/sql. However, pgx provides additional query modes via [QueryExecMode].
package pgx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrNoRows is returned when a query returns no rows.
var ErrNoRows = errors.New("no rows in result set")

// ErrTxClosed is returned when a transaction is already closed.
var ErrTxClosed = errors.New("tx is closed")

// ErrTxCommitRollback is returned when a transaction that has already been
// rolled back is committed.
var ErrTxCommitRollback = errors.New("commit unexpectedly resulted in rollback")

// Identifier is a PostgreSQL identifier (e.g. table name, column name).
// It is quoted appropriately when used in SQL queries.
type Identifier []string

// Sanitize returns a sanitized string safe for SQL interpolation.
func (ident Identifier) Sanitize() string {
	var str string
	for i, part := range ident {
		if i > 0 {
			str += "."
		}
		str += `"` + sanitizeIdentPart(part) + `"`
	}
	return str
}

func sanitizeIdentPart(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			result = append(result, '"', '"')
		} else {
			result = append(result, s[i])
		}
	}
	return string(result)
}

// Row is the result of calling QueryRow to select a single row.
type Row struct {
	rows Rows
	err  error
}

// Scan reads the values from the current row into dest values positionally.
// Returns ErrNoRows if no rows were found.
func (r *Row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.rows == nil {
		return ErrNoRows
	}
	defer r.rows.Close()

	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return ErrNoRows
	}

	return r.rows.Scan(dest...)
}

// Rows is the result of a query. Its cursor starts before the first row of the
// result set. Use Next to advance from row to row.
type Rows interface {
	// Close closes the rows, making the connection ready for use again. It is safe
	// to call Close after rows is already closed.
	Close()

	// Err returns any error that occurred while reading.
	Err() error

	// CommandTag returns the command tag from the final command in the query.
	CommandTag() pgconn.CommandTag

	// FieldDescriptions returns the field descriptions of the columns.
	FieldDescriptions() []pgconn.FieldDescription

	// Next prepares the next row for reading. It returns true if there is another
	// row and false if no more rows are available or an error occurred.
	Next() bool

	// Scan reads the values from the current row into dest values positionally.
	Scan(dest ...any) error

	// Values returns the decoded row values.
	Values() ([]any, error)

	// RawValues returns the raw bytes of the row values.
	RawValues() [][]byte

	// Conn returns the underlying *Conn on which the query was executed.
	Conn() *Conn
}

// CollectRows iterates through rows, calling fn for each row, and collecting
// the results into a slice of T.
func CollectRows[T any](rows Rows, fn func(row CollectableRow) (T, error)) ([]T, error) {
	defer rows.Close()

	var slice []T
	for rows.Next() {
		value, err := fn(rows)
		if err != nil {
			return nil, err
		}
		slice = append(slice, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return slice, nil
}

// CollectableRow is the interface for a row that can be collected.
type CollectableRow interface {
	Scan(dest ...any) error
	Values() ([]any, error)
	FieldDescriptions() []pgconn.FieldDescription
}

// RowToStructByName returns a T scanned from row. T must be a struct. T must
// have the same number of named public fields as row has fields. The row and
// struct fields will be matched by name, ignoring case.
func RowToStructByName[T any](row CollectableRow) (T, error) {
	var value T
	err := row.Scan(&value)
	if err != nil {
		return value, fmt.Errorf("pgx: RowToStructByName: %w", err)
	}
	return value, nil
}

// TypeMap returns a pgtype.Map that can be used to encode and decode PostgreSQL
// types. It is safe to use concurrently.
func TypeMap() *pgtype.Map {
	return pgtype.NewMap()
}

// connect is the internal implementation of Connect. It exists to allow
// testing with a custom context.
func connect(ctx context.Context, config *Config) (*Conn, error) {
	if config == nil {
		return nil, errors.New("config must not be nil")
	}
	return newConn(ctx, config)
}

package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

const (
	dbname               = "codb"
	output_parsing_table = "cmd_output_parsing_table"
)

type DB struct {
	db *sql.DB
}

// NewMemDB create a new sqlite db in memory.
func NewMemDB() (*DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// CreateTable create a table with given statement, the get() method will returns the create statement
// and the table name.
func (d *DB) CreateTable(ctx context.Context, get func() (string, string)) (Table, error) {
	stmt, name := get()
	_, err := d.db.ExecContext(ctx, stmt)
	if err != nil {
		return Table{}, fmt.Errorf("%w: create table %s", err, name)
	}
	return Table{
		name: name,
		db:   d.db,
	}, nil
}

// Close close the db.
func (d *DB) Close() error {
	return d.db.Close()
}

// Table represents a table in sqlite db.
type Table struct {
	name string
	db   *sql.DB
}

// Insert insert a row into the table.
func (t *Table) Insert(ctx context.Context, columns []string, values ...any) error {
	if len(columns) != len(values) {
		return fmt.Errorf("columns and values length mismatch: %d != %d", len(columns), len(values))
	}
	var (
		stmt  strings.Builder
		flags strings.Builder
	)
	stmt.WriteString("INSERT INTO ")
	stmt.WriteString(t.name)
	stmt.WriteString(" (")
	for i, c := range columns {
		if i > 0 {
			stmt.WriteString(",")
			flags.WriteString(",")
		}
		stmt.WriteString(c)
		flags.WriteString("?")
	}
	stmt.WriteString(") VALUES (")
	stmt.WriteString(flags.String())
	stmt.WriteString(");")

	_, err := t.db.ExecContext(ctx, stmt.String(), values...)
	return err
}

// Delete delete a row from the table, the caller need to build the where clause.
func (t *Table) Delete(ctx context.Context, where string) error {
	var stmt strings.Builder
	stmt.WriteString("DELETE FROM ")
	stmt.WriteString(t.name)
	stmt.WriteString(" WHERE ")
	stmt.WriteString(where)
	stmt.WriteString(";")

	_, err := t.db.ExecContext(ctx, stmt.String())
	return err
}

// Query query the table with given where clause, and returns the result.
func (t *Table) Query(ctx context.Context, columns []string, where string) ([][]string, error) {
	if len(columns) == 0 {
		return nil, nil
	}
	var stmt strings.Builder
	stmt.WriteString("SELECT ")
	stmt.WriteString(strings.Join(columns, ","))
	stmt.WriteString(" FROM ")
	stmt.WriteString(t.name)
	if where != "" {
		stmt.WriteString(" WHERE ")
		stmt.WriteString(where)
	}
	stmt.WriteString(";")

	rows, err := t.db.QueryContext(ctx, stmt.String())
	if err != nil {
		return nil, fmt.Errorf("%w: query table %s", err, t.name)
	}
	rs := make([][]string, 0)
	for rows.Next() {
		var ps []any
		values := make([]string, len(columns))
		for i := range values {
			ps = append(ps, &values[i])
		}
		if err := rows.Scan(ps...); err != nil {
			return nil, fmt.Errorf("%w: scan row", err)
		}
		rs = append(rs, values)
	}
	return rs, nil
}

// StatementCreateOutputParsingTable returns a statement to create a table and the table name.
func StatementCreateOutputParsingTable() (string, string) {
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		flow_id 	TEXT NOT NULL,
		node_seq    INT  NOT NULL,
		node_name   TEXT NOT NULL,
		c0 		TEXT,
		c1 		TEXT,
		c2 		TEXT,
		c3 		TEXT,
		c4 		TEXT,
		c5 		TEXT,
		c6 		TEXT,
		c7 		TEXT,
		c8 		TEXT,
		c9 		TEXT,
		c10 	TEXT,
		c11		TEXT,
		c12		TEXT,
		c13		TEXT,
		c14		TEXT,
		c15		TEXT,
		c16		TEXT,
		c17		TEXT,
		c18		TEXT,
		c19		TEXT,
		c20 	TEXT
	 );`, output_parsing_table)
	return stmt, output_parsing_table
}

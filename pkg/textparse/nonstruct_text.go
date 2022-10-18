package textparse

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cofxlabs/cofx/pkg/sqlite"
)

type NonStructText struct {
	// name is the name of the text that be parsed, you can use a file name or other string.
	name string
	// split indicates a string to split the text, it's a seprator.
	split string
	// extractFields indicates the number of fields to extract from the text, .e.g [0, 1].
	extractFields []int

	tbl sqlite.Table
}

func New(name string, split string, extract []int) (*NonStructText, error) {
	mdb, err := sqlite.NewMemDB()
	if err != nil {
		return nil, err
	}
	tb, err := mdb.CreateTable(context.Background(), stmNonStructTextTableCreate)
	if err != nil {
		return nil, err
	}
	return &NonStructText{
		tbl:           tb,
		name:          name,
		split:         split,
		extractFields: extract,
	}, nil
}

// Clear delete all data of the non struct text table.
func (t *NonStructText) Clear(ctx context.Context) error {
	return t.tbl.Delete(ctx, stmNonStructTextTableDeleteWhere(t.name))
}

// ParseFile parses a file and writes the extracted fields into the database.
func (t *NonStructText) ParseFile(ctx context.Context, filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("%w: file '%s'", err, filepath)
	}
	defer f.Close()

	buf := bufio.NewScanner(f)
	for {
		if !buf.Scan() {
			break
		}
		line := buf.Text()
		if err := t.ParseLine(ctx, line); err != nil {
			return err
		}
	}
	return nil
}

// ParseLine parses a line of text and writes the extracted fields into the database.
func (t *NonStructText) ParseLine(ctx context.Context, line string) error {
	columns, values := newNonStructTextInsertSlice(t.name, line)

	// extract fields into db
	var fields []string
	if t.split == "" {
		fields = strings.Fields(line)
	} else {
		fields = strings.Split(line, t.split)
	}
	for i, f := range t.extractFields {
		if f < len(fields) {
			v := strings.TrimSpace(fields[f])
			values = append(values, v)
		} else {
			values = append(values, "")
		}
		columns = append(columns, fmt.Sprintf("c%d", i))
	}

	return t.tbl.Insert(ctx, columns, values...)
}

// Query get the column values that wanted by the where condition.
func (t *NonStructText) Query(ctx context.Context, columns []string, where ...string) ([][]string, error) {
	if len(where) == 0 {
		return t.tbl.Query(ctx, columns, "")
	} else {
		return t.tbl.Query(ctx, columns, where[0])
	}
}

// String get the specified column value as string to return.
func (t *NonStructText) String(ctx context.Context, column string, where string) (string, error) {
	rows, err := t.Query(ctx, []string{column}, where)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", ErrNotfound
	}
	r := rows[0]
	if len(r) == 0 {
		return "", ErrNotfound
	}
	return r[0], nil
}

// Int get the specified column value as int to return.
func (t *NonStructText) Int(ctx context.Context, column string, where string) (int, error) {
	s, err := t.String(ctx, column, where)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Row get the result of the query as one row.
func (t *NonStructText) Row(ctx context.Context, column []string, where string) ([]string, error) {
	rows, err := t.Query(ctx, column, where)
	if err != nil {
		return nil, err
	}
	if len(rows) == 1 {
		return rows[0], nil
	} else if len(rows) > 1 {
		return nil, ErrTooMany
	} else {
		return nil, ErrNotfound
	}
}

type Rows [][]string

// Row2Slice converts the row that be specified by 'n' in the result to a slice.
func (rs Rows) Row2Slice(n int) []string {
	if len(rs) == 0 || n >= len(rs) {
		return nil
	}
	return rs[n]
}

// Column2Slice converts the column that be specified by 'n' in the result to a slice.
func (rs Rows) Column2Slice(n int) []string {
	if len(rs) == 0 || n >= len(rs[0]) {
		return nil
	}
	var s []string
	for _, r := range rs {
		s = append(s, r[n])
	}
	return s
}

// String get the specified position value as string to return.
func (rs Rows) String(r, c int) string {
	if len(rs) == 0 || r >= len(rs) || c >= len(rs[r]) {
		return ""
	}
	return rs[r][c]
}

// Int get the specified position value as int to return.
func (rs Rows) Int(r, c int) (int, error) {
	s := rs.String(r, c)
	return strconv.Atoi(s)
}

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
func (t *NonStructText) Qeury(ctx context.Context, columns []string, where ...string) ([][]string, error) {
	if len(where) == 0 {
		return t.tbl.Query(ctx, columns, "")
	} else {
		return t.tbl.Query(ctx, columns, where[0])
	}
}

// String get the specified column value as string to return.
func (t *NonStructText) String(ctx context.Context, column string, where string) (string, error) {
	rows, err := t.Qeury(ctx, []string{column}, where)
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
	rows, err := t.Qeury(ctx, column, where)
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

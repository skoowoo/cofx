package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dbname   = "codb"
	tp_table = "textparse_table"
)

// StatementTextParseTable returns a statement to create a table and the table name.
func StatementTextParseTable() (string, string) {
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
	 );`, tp_table)
	return stmt, tp_table
}

func TestCRUD(t *testing.T) {
	db, err := NewMemDB()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	tb, err := db.CreateTable(context.Background(), StatementTextParseTable)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, tp_table, tb.name)

	// insert
	{
		columns := []string{"flow_id", "node_seq", "node_name", "c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9"}
		values := []any{"1", "1000", "test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
		if err := tb.Insert(context.Background(), columns, values...); err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	// query
	{
		columns := []string{"c0", "c1"}
		rs, err := tb.Query(context.Background(), columns, "flow_id = 1 and node_seq = 1000")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 1, len(rs))
		assert.Equal(t, []string{"0", "1"}, rs[0])
	}

	// delete
	{
		if err := tb.Delete(context.Background(), "flow_id = 1"); err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	// query
	{
		columns := []string{"c0", "c1"}
		rs, err := tb.Query(context.Background(), columns, "flow_id = 1")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 0, len(rs))
	}
}

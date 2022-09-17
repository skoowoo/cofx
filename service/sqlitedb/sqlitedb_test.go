package sqlitedb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRUD(t *testing.T) {
	db, err := NewMemDB()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	tb, err := db.CreateTable(context.Background(), StatementCreateOutputParsingTable)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, output_parsing_table, tb.name)

	// insert
	{
		columns := []string{"flow_id", "fseq", "fname", "c0", "c1", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9"}
		values := []any{"1", "1000", "test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
		if err := tb.Insert(context.Background(), columns, values...); err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	// query
	{
		columns := []string{"c0", "c1"}
		rs, err := tb.Query(context.Background(), columns, "flow_id = 1 and fseq = 1000")
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

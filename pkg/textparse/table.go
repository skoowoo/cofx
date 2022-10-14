package textparse

import "fmt"

const (
	dbname   = "codb"
	tp_table = "textparse_table"
)

// stmNonStructTextTableCreate returns a statement to create a table and the table name.
func stmNonStructTextTableCreate() (string, string) {
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text_name   TEXT NOT NULL,
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
		c20 	TEXT,
		origin 	TEXT
	 );`, tp_table)
	return stmt, tp_table
}

func stmNonStructTextTableDeleteWhere(name string) string {
	return fmt.Sprintf("text_name = '%s'", name)
}

func newNonStructTextInsertSlice(name string, origin ...string) ([]string, []interface{}) {
	if len(origin) == 0 {
		return []string{"text_name"}, []interface{}{name}
	} else {
		return []string{"text_name", "origin"}, []interface{}{name, origin[0]}
	}
}

package db

import (
	"fmt"
)

const (
	output_parsing_table = "cmd_output_parsing_table"
)

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

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func TestTableSizesCommand(t *testing.T) {
	saved := dburi
	dburi = "postgres://postgres@db/postgres?sslmode=disable"
	var buf bytes.Buffer
	db, err := sql.Open("postgres", dburi)
	if err != nil {
		t.Errorf(fmt.Sprintf("Got error %s", err))
	}
	defer db.Close()
	_, err = db.Exec("CREATE TEMP TABLE testdata (d jsonb)")
	err = tableSize(&buf)
	dburi = saved
	if err != nil {
		t.Errorf(fmt.Sprintf("Got error %s", err))
	}
	raw := []string{
		"    NAME   | TOTALSIZE  | TABLESIZE  | INDEXSIZE  ",
		"+----------+------------+------------+-----------+",
		"  testdata | 8192 bytes | 8192 bytes | 0 bytes    \n",
	}
	expected := strings.Join(raw, "\n")

	if buf.String() != expected {
		f2 := "table-size output is:\n%s\nexpected:\n%s"
		t.Errorf(f2, buf.String(), expected)
	}
}

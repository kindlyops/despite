package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestTableSizesCommand(t *testing.T) {
	saved := dburi
	dburi = "postgres://localhost/postgres?sslmode=disable"
	var buf bytes.Buffer
	// TODO set up some tables to get sizes from
	err := tableSize(&buf)
	dburi = saved
	if err != nil {
		t.Errorf(fmt.Sprintf("Got error %s", err))
	}
	raw := []string{
		"  NAME | TOTALSIZE | TABLESIZE | INDEXSIZE  ",
		"+------+-----------+-----------+-----------+\n",
	}
	expected := strings.Join(raw, "\n")

	if buf.String() != expected {
		f2 := "table-size output is:\n%q\nexpected:\n%q"
		t.Errorf(f2, buf.String(), expected)
	}
}

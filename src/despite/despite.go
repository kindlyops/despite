//usr/bin/env go run $0 "$@"; exit $?
// Copyright 2016 Kindly Ops, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/codegangsta/cli"
	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
)

var dburi string
var buildstamp = "defined in linker flags"
var githash = "defined at compile time"
var tag = ""

func outliers(output io.Writer) error {
	var (
		totalExecTime string
		propExecTime  string
		ncalls        string
		syncIoTime    string
		query         string
	)
	db, err := sql.Open("postgres", dburi)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := hasPgStatsStatement(db); err != nil {
		return cli.NewExitError(
			`pg_stat_statements extension need to be installed in the public schema first.
You can install it by running CREATE EXTENSION pg_stat_statements;
and then adding shared_preload_libraries = 'pg_stat_statements' to postgres.conf`, 1)
	}

	sql := `SELECT interval '1 millisecond' * total_time AS total_exec_time,
        to_char((total_time/sum(total_time) OVER()) * 100, 'FM90D0') || '%'  AS prop_exec_time,
        to_char(calls, 'FM999G999G999G990') AS ncalls,
        interval '1 millisecond' * (blk_read_time + blk_write_time) AS sync_io_time,
        query AS query
        FROM pg_stat_statements WHERE userid = (SELECT usesysid FROM pg_user WHERE usename = current_user LIMIT 1)
        ORDER BY total_time DESC LIMIT 10`
	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"total_exec_time", "prop_exec_time", "ncalls", "sync_io_time", "query"})
	table.SetBorder(false)
	for rows.Next() {
		err := rows.Scan(&totalExecTime, &propExecTime, &ncalls, &syncIoTime, &query)
		if err != nil {
			return err
		}
		table.Append([]string{totalExecTime, propExecTime, ncalls, syncIoTime, query})
	}
	table.Render()
	return nil
}

func hasPgStatsStatement(db *sql.DB) error {
	sql := `SELECT exists(
        SELECT 1 FROM pg_extension e LEFT JOIN pg_namespace n ON n.oid = e.extnamespace
        WHERE e.extname='pg_stat_statements' AND n.nspname = 'public'
    ) AS available`
	var enabled string
	err := db.QueryRow(sql).Scan(&enabled)
	if err != nil {
		// TODO convert this to a multi-error
		return cli.NewExitError("Error checking pg_stat_statements", 1)
	}
	fmt.Printf("Result was: '%s'\n", enabled)
	if enabled != "true" {
		return cli.NewExitError("not available", 1)
	}
	return nil
}

func tableSize(output io.Writer) error {
	var (
		tableSize string
		totalSize string
		indexSize string
		name      string
	)
	db, err := sql.Open("postgres", dburi)
	if err != nil {
		return err
	}
	defer db.Close()
	// much love for heroku data team, who originally published in pg-extras
	// https://github.com/heroku/heroku-pg-extras/blob/master/lib/heroku/command/pg.rb
	sql := `SELECT c.relname AS name,
    pg_size_pretty(pg_total_relation_size(c.oid)) as total_size,
    pg_size_pretty(pg_table_size(c.oid)) AS table_size,
    pg_size_pretty(pg_indexes_size(c.oid)) AS index_size
  FROM pg_class c
    LEFT JOIN pg_namespace n ON (n.oid = c.relnamespace)
  WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
    AND n.nspname !~ '^pg_toast'
    AND c.relkind='r'
  ORDER BY pg_total_relation_size(c.oid) DESC`
	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	table := tablewriter.NewWriter(output)
	table.SetHeader([]string{"Name", "TotalSize", "TableSize", "IndexSize"})
	table.SetBorder(false)
	for rows.Next() {
		err := rows.Scan(&name, &totalSize, &tableSize, &indexSize)
		if err != nil {
			return err
		}
		table.Append([]string{name, totalSize, tableSize, indexSize})
	}
	table.Render()
	return nil
}

func tableSizeCmd(ctx *cli.Context) error {
	err := tableSize(os.Stdout)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s", err), 1)
	}
	return nil
}

func outliersCmd(ctx *cli.Context) error {
	err := outliers(os.Stdout)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s", err), 1)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "despite"
	app.Usage = "swiss army knife for the harried operator"
	app.Version = fmt.Sprintf("%s compiled from %s on %s", tag, githash, buildstamp)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "verbose, V",
			Value:  "无",
			Usage:  "verbosity of output",
			EnvVar: "DESPITE_VERBOSITY",
		},
		cli.StringFlag{
			Name:        "dburi, D",
			Value:       "postgres://localhost/postgres?sslmode=disable",
			Usage:       "postgres://dbuser:dbpassword@hostname/dbname?sslmode=disable",
			EnvVar:      "DESPITE_DBURI",
			Destination: &dburi,
		},
		cli.IntFlag{
			Name:   "exit, e",
			Value:  0,
			Usage:  "exit with `CODE`",
			EnvVar: "DESPITE_EXIT",
		},
	}
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Elliot Murphy",
			Email: "elliot@kindlyops.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "pg:table-size",
			Aliases: []string{"table-size"},
			Usage:   "print table sizes in descending order",
			Action:  tableSizeCmd,
		},
		{
			Name:   "pg:outliers",
			Usage:  "10 queries with longest aggregate execution time",
			Action: outliersCmd,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		fmt.Println("despite is a swiss army knife for the harried operator. -h for usage")

		return cli.NewExitError("", ctx.Int("exit"))
	}

	app.Run(os.Args)
}

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
	"os"

	"github.com/codegangsta/cli"
	_ "github.com/lib/pq"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var dburi string
	app := cli.NewApp()
	app.Name = "despite"
	app.Usage = "One day this should do something"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "verbose, V",
			Value:  "无",
			Usage:  "verbosity of output",
			EnvVar: "DESPITE_VERBOSITY",
		},
		cli.StringFlag{
			Name:        "dburi, D",
			Value:       "localhost/postgres",
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
			Action: func(ctx *cli.Context) error {
				var (
					size string
					name string
				)
				fmt.Println(dburi)
				db, err := sql.Open("postgres", dburi)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), 1)
				}
				// much love for heroku data team, who originally published this
				// query in pg-extras
				// https://github.com/heroku/heroku-pg-extras/blob/master/lib/heroku/command/pg.rb
				sql := `SELECT c.relname AS name,
          pg_size_pretty(pg_table_size(c.oid)) AS size
        FROM pg_class c
        LEFT JOIN pg_namespace n ON (n.oid = c.relnamespace)
        WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
        AND n.nspname !~ '^pg_toast'
        AND c.relkind='r'
        ORDER BY pg_table_size(c.oid) DESC`
				rows, err := db.Query(sql)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), 1)
				}
				defer rows.Close()
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Size"})
				for rows.Next() {
					err := rows.Scan(&name, &size)
					if err != nil {
						return cli.NewExitError(fmt.Sprintf("%s", err), 1)
					}
					table.Append([]string{name, size})
				}
				table.Render()
				return nil
			},
		},
	}
	app.Action = func(ctx *cli.Context) error {
		fmt.Println("Wow I can't believe you ran this.")

		return cli.NewExitError("", ctx.Int("exit"))
	}

	app.Run(os.Args)
}

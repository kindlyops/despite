//usr/bin/env go run $0 $@; exit $?
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
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
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
	app.Action = func(ctx *cli.Context) error {
		fmt.Println("Wow I can't believe you ran this.")

		return cli.NewExitError("", ctx.Int("exit"))
	}

	app.Run(os.Args)
}

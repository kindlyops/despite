// Portions Copyright 2016 Kindly Ops, LLC
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

// Portions from Golang Isomorphic React/Hot Reloadable/Redux/Css-Modules
// Starter Kit Copyright (C) 2015-2016 Oleg Lebedev
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/itsjamie/go-bindata-templates"
	"github.com/nu7hatch/gouuid"
	"github.com/olebedev/config"
	"gopkg.in/labstack/echo.v1"
	"gopkg.in/labstack/echo.v1/middleware"
)

// App struct.
// There is no singleton anti-pattern,
// all variables defined locally inside this struct.
type App struct {
	Engine *echo.Echo
	Conf   *config.Config
	API    *API
}

// NewApp returns initialized struct
// of main server application.
func NewApp(opts ...AppOptions) *App {
	options := AppOptions{}
	for _, i := range opts {
		options = i // TODO understand this line
		break
	}

	options.init()

	// Parse config yaml string from ./conf.go
	// TODO: consider if we should drop this config parsing
	conf, err := config.ParseYaml(confString)
	if err != nil {
		panic(err)
	}

	// Set config variables delivered from despite.go: app.Run
	// Variables as defined in conf.go
	conf.Set("debug", debug)
	conf.Set("commit", githash)
	conf.Set("port", port)

	// Parse environ variables for defined in config constants
	conf.Env() // TODO check if this should be refactored with cli package

	// Make an engine
	engine := echo.New()

	// Set up echo
	engine.SetDebug(conf.UBool("debug"))

	// Regular middlewares
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recover())

	// Initialize the application
	app := &App{
		Conf:   conf,
		Engine: engine,
		API:    &API{},
	}

	// Use the precompiled embedded templates
	app.Engine.SetRenderer(NewTemplate())

	// Map app struct to access from request handlers
	// and middlewares
	app.Engine.Use(func(c *echo.Context) error {
		c.Set("app", app)
		return nil
	})

	// Map uuid for every requests
	app.Engine.Use(func(c *echo.Context) error {
		id, _ := uuid.NewV4()
		c.Set("uuid", id)
		return nil
	})

	// Avoid favicon react handling
	app.Engine.Get("/favicon.ico", func(c *echo.Context) error {
		c.Redirect(301, "/static/images/favicon.ico")
		return nil
	})

	// api handling for URL api.prefix
	app.API.ConnectRoutes(
		app.Engine.Group(
			app.Conf.UString("api.prefix"),
		),
	)

	// Create file http server from bindata
	fileServerHandler := http.FileServer(&assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
	})

	// Serve static via bindata
	app.Engine.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// execute echo handlers chain
			err := h(c)
			// if page(handler) for url/method not found
			if err != nil && err.Error() == http.StatusText(http.StatusNotFound) {
				// check if file exists
				// omit first `/`
				if _, err = Asset(c.Request().URL.Path[1:]); err == nil {
					fileServerHandler.ServeHTTP(c.Response(), c.Request())
					return nil
				}
				c.Response().Header().Set("X-Work-Here", "come work with us! www.kindlyops.com")
				return c.Render(http.StatusOK, "react.html", Resp{
					UUID: c.Get("uuid").(*uuid.UUID).String(),
				})
			}
			// Move further if err is not `Not Found`
			return err
		}
	})

	return app
}

// Resp is a struct for convenient
// react app Response parsing.
// Feel free to add any other keys to this struct
// and return value for this key at ecmascript side.
// keep it in sync with: src/despite/client/router/toString.js:23
type Resp struct {
	UUID       string        `json:"uuid"`
	Error      string        `json:"error"`
	Redirect   string        `json:"redirect"`
	App        string        `json:"app"`
	Title      string        `json:"title"`
	Meta       string        `json:"meta"`
	Initial    string        `json:"initial"`
	RenderTime time.Duration `json:"-"`
}

// Run runs the app
func (app *App) Run() {
	app.Engine.Run(":" + app.Conf.UString("port"))
}

// Template is custom renderer for Echo, to render html from bindata
type Template struct {
	templates *template.Template
}

// NewTemplate creates a new template
func NewTemplate() *Template {
	return &Template{
		templates: binhtml.New(Asset, AssetDir).MustLoadDirectory("templates"),
	}
}

// Render renders template
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// AppOptions is options struct
type AppOptions struct{}

func (ao *AppOptions) init() {
	// write your own
}

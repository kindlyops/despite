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
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"

	"github.com/nu7hatch/gouuid"

	"gopkg.in/labstack/echo.v1"
	"gopkg.in/olebedev/go-duktape-fetch.v2"
	"gopkg.in/olebedev/go-duktape.v2"
)

// React struct contains duktape
// pool to serve HTTP requests and
// separates some domain specific
// resources
type React struct {
	pool
	debug bool
	path  string
}

// NewReact initialize a React struct
func NewReact(filePath string, debug bool, server http.Handler) *React {
	r := &React{
		debug: debug,
		path:  filePath,
	}
	if !debug {
		r.pool = newDuktapePool(filePath, runtime.NumCPU()+1, server)
	} else {
		// Use onDemand pool to load full react
		// app each time for any http requests.
		// Useful to debug the app.
		r.pool = &onDemandPool{
			path:   filePath,
			engine: server,
		}
	}
	return r
}

// Handle handles all HTTP requests which
// have not been caught via static file
// handler or other middlewares
func (r *React) Handle(c *echo.Context) error {
	UUID := c.Get("uuid").(*uuid.UUID)
	defer func() {
		if r := recover(); r != nil {
			c.Render(http.StatusInternalServerError, "react.html", Resp{
				UUID:  UUID.String(),
				Error: r.(string),
			})
		}
	}()

	vm := r.get()

	start := time.Now()
	select {
	case re := <-vm.Handle(map[string]interface{}{
		"url":     c.Request().URL.String(),
		"headers": c.Request().Header,
		"uuid":    UUID.String(),
	}):
		re.RenderTime = time.Since(start)
		// Return vm back to pool
		r.put(vm)
		// Handle the Response
		if len(re.Redirect) == 0 && len(re.Error) == 0 {
			// if no redirection and no errors
			c.Response().Header().Set("X-React-Render-Time", fmt.Sprintf("%s", re.RenderTime))
			c.Response().Header().Set("X-Work-Here", "come work with us! www.kindlyops.com")
			return c.Render(http.StatusOK, "react.html", re)
			// if redirect
		} else if len(re.Redirect) != 0 {
			return c.Redirect(http.StatusMovedPermanently, re.Redirect)
			// if internal error
		} else if len(re.Error) != 0 {
			c.Response().Header().Set("X-React-Render-Time", fmt.Sprintf("%s", re.RenderTime))
			c.Response().Header().Set("X-Work-Here", "help us fix the bugs! www.kindlyops.com")
			return c.Render(http.StatusInternalServerError, "react.html", re)
		}
	case <-time.After(2 * time.Second):
		// release duktape context
		r.drop(vm)
		return c.Render(http.StatusInternalServerError, "react.html", Resp{
			UUID:  UUID.String(),
			Error: fmt.Sprintf("Render timeout on %s after 2 seconds", c.Request().URL.String()),
		})
	}
	return nil
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

// HTMLApp returns an application template
func (r Resp) HTMLApp() template.HTML {
	return template.HTML(r.App)
}

// HTMLTitle returns a title data
func (r Resp) HTMLTitle() template.HTML {
	return template.HTML(r.Title)
}

// HTMLMeta returns a meta data
func (r Resp) HTMLMeta() template.HTML {
	return template.HTML(r.Meta)
}

// Interface to serve React app on demand or from prepared pool
type pool interface {
	get() *ReactVM
	put(*ReactVM)
	drop(*ReactVM)
}

func newDuktapePool(filePath string, size int, engine http.Handler) *duktapePool {
	pool := &duktapePool{
		path:   filePath,
		ch:     make(chan *ReactVM, size),
		engine: engine,
	}

	go func() {
		for i := 0; i < size; i++ {
			pool.ch <- newReactVM(filePath, engine)
		}
	}()

	return pool
}

// newReactVM loads bundle.js to context
func newReactVM(filePath string, engine http.Handler) *ReactVM {
	vm := &ReactVM{
		Context: duktape.New(),
		ch:      make(chan Resp, 1),
	}

	vm.PevalString(`var console = {log:print,warn:print,error:print,info:print}`)
	fetch.PushGlobal(vm.Context, engine)
	app, err := Asset(filePath)
	if err != nil {
		panic(err)
	}

	// Reduce CGO calls
	vm.PushGlobalGoFunction("__goServerCallback__", func(ctx *duktape.Context) int {
		result := ctx.SafeToString(-1)
		vm.ch <- func() Resp {
			var re Resp
			json.Unmarshal([]byte(result), &re)
			return re
		}()
		return 0
	})

	fmt.Printf("%s loaded\n", filePath)
	if err := vm.PevalString(string(app)); err != nil {
		derr := err.(*duktape.Error)
		panic(derr.Message)
	}
	vm.PopN(vm.GetTop())
	return vm
}

// ReactVM wraps duktape.Context
type ReactVM struct {
	*duktape.Context
	ch chan Resp
}

// Handle handles http requests
func (r *ReactVM) Handle(req map[string]interface{}) <-chan Resp {
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	// Keep it in sinc with `src/despite/client/index.js:4`
	r.PevalString(`main(` + string(b) + `, __goServerCallback__)`)
	return r.ch
}

// DestroyHeap destroys the context's heap
func (r *ReactVM) DestroyHeap() {
	close(r.ch)
	r.Context.DestroyHeap()
}

// Pool's implementations

type onDemandPool struct {
	path   string
	engine http.Handler
}

func (f *onDemandPool) get() *ReactVM {
	return newReactVM(f.path, f.engine)
}

func (f onDemandPool) put(c *ReactVM) {
	c.Lock()
	c.FlushTimers()
	c.Gc(0)
	c.DestroyHeap()
}

func (f *onDemandPool) drop(c *ReactVM) {
	f.put(c)
}

type duktapePool struct {
	ch     chan *ReactVM
	path   string
	engine http.Handler
}

func (o *duktapePool) get() *ReactVM {
	return <-o.ch
}

func (o *duktapePool) put(ot *ReactVM) {
	// Drop any futured async calls
	ot.Lock()
	ot.FlushTimers()
	ot.Unlock()
	o.ch <- ot
}

func (o *duktapePool) drop(ot *ReactVM) {
	ot.Lock()
	ot.FlushTimers()
	ot.Gc(0)
	ot.DestroyHeap()
	ot = nil
	o.ch <- newReactVM(o.path, o.engine)
}

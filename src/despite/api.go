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

import "gopkg.in/labstack/echo.v1"

// The API is not yet defined
type API struct{}

// ConnectRoutes connects the routes
func (api *API) ConnectRoutes(group *echo.Group) {
	group.Get("/healthz", api.HealthzHandler)
	group.Get("/conf", api.ConfigHandler)
}

// HealthzHandler reports a healthcheck for this app
func (api *API) HealthzHandler(c *echo.Context) error {
	app := c.Get("app").(*App)

	// if you need to fake connection latency
	// <-time.After(time.Millisecond * 500)

	c.JSON(200, app.Conf.Root)
	return nil
}

// ConfigHandler exposes the config for this app
func (api *API) ConfigHandler(c *echo.Context) error {
	app := c.Get("app").(*App)
	c.JSON(200, app.Conf.Root)
	return nil
}

# DESPITE
[![CircleCI](https://circleci.com/gh/kindlyops/despite.svg?style=svg)](https://circleci.com/gh/kindlyops/despite)

Despite the pressure, self-doubt, hysteria, and rampant speculation that surrounds operational emergencies, we still need dependable tools to help us probe-sense-respond.

The first set of commands are some useful DB diagnostics ported from the
heroku pg-extras project. Unlike pg-extras, this command will connect to any
PG database, not just ones running on heroku. There are absolutely minimal
binary dependencies, not even libpq.

## building

This application is compiled inside a Docker container that has the go
toolchain installed. Using a build container guarantees that we are all using
the same toolchain to compile, and using the [gb](https://getgb.io/) build tool
ensures that we have reproducible builds without import rewriting, depending
on github uptime during compile, or setting up environment variables for paths.

To check and see if you have docker available and set up

    docker -v
    docker-compose -v
    docker info

If you don't have docker running, use the instructions at https://www.docker.com.
At the time of writing, this is working fine with docker 1.11.1-beta13.1.
Once you have docker set up:

    make        # show the available make targets
    make image  # build and upload the go toolchain container
    make build  # compile, using docker build container
    make test   # run tests (provisions postgres inside docker)

## TODO

* [x] learn how to make unit tests
* [x] convert to compile with a build image
* [ ] make an animated gif for the readme similar to https://github.com/tcnksm/ghr
* [ ] set up a homebrew tap to make this easy to install on OSX
* [ ] set up circleci to publish docker image to docker hub
* [ ] set up circleci to publish binaries to github releases

* [ ] set up bash autocompletion
* [ ] add notes on running via docker image
* [ ] add a Makefile that automatically gets the dependencies
      via homebrew or apt and go get, and has a make test target
* [ ] add docker compose support

## http server experiment

* [ ] add a serve command that runs HTTP server using
      https://github.com/gocraft/web and exposes commands
* [ ] instrument the http server with https://github.com/gocraft/health
* [ ] experiment with calling into PostgreSQL
      http://www.cybertec.at/2016/05/beating-uber-with-a-postgresql-prototype/
* [ ] publish application metrics into prometheus
      https://prometheus.io/
* [ ] visualize application metrics with grafana
      https://prometheus.io/docs/visualization/grafana/
* [ ] set up vagrant with mesos-playa and some frameworks
  * kubernetes
  * chronos
  * marathon
* [ ] set up vagrant with CoreOS and kubernetes

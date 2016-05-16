# DESPITE

## building

You will need go 1.6 or later and the gb tool installed.

    go get github.com/constabulary/gb/...

    gb build all
    bin/despite

## TODO

* [ ] learn how to make unit tests
* [ ] set up circleci to publish docker image to docker hub
* [ ] set up circleci to publish binaries to github releases
* [ ] set up a homebrew tap to make this easy to install on OSX
* [ ] set up bash autocompletion
* [ ] add notes on running via docker image
* [ ] turn on slack integrations
* [ ] add a Makefile that automatically gets the dependencies
      via homebrew or apt and go get, and has a make test target

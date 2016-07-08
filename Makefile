.DEFAULT_GOAL  := help
.PHONY         : help test clean
docker         := $(shell command -v docker 2> /dev/null)
docker-compose := $(shell command -v docker-compose 2> /dev/null)
xgo            := $(shell command -v xgo 2> /dev/null)
GOOS           ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
CC             ?= o64-clang
CXX            ?= o64-clang++
BINDATA         = src/despite/bindata.go
BINDATA_FLAGS   = -pkg=main -prefix=src/despite/data
BUNDLE          = src/despite/data/static/build/bundle.js
APP             = $(shell find src/client -type f)
NODE_BIN        = $(shell npm bin)
THIS_FILE_PATH :=$(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
THIS_DIR       :=$(shell cd $(dir $(THIS_FILE_PATH));pwd)
THIS_MAKEFILE  :=$(notdir $(THIS_FILE_PATH))
GOPATH          = $(THIS_DIR)
XGO_TARGETS     = linux/amd64,linux/arm-7,darwin-10.9/*,windows-6.0/*

clean:
	@git clean -x -f
	@rm -f $(BINDATA)
	@rm -f src/despite/data/static/build/*

check-deps: ## Check if we have required dependencies
ifndef xgo
	@echo "I couldn't find the xgo command, install with go get github.com/karalabe/xgo"
endif
ifndef docker
	@echo "I couldn't find the docker command, install from www.docker.com"
endif
ifndef docker-compose
	@echo "I couldn't find the docker-compose command, install from www.docker.com"
endif
	@docker info >/dev/null

# the stuff to the right of the pipe symbol is order-only prerequisites
test: | check-deps ## Run the tests
# run go container, and execute tests inside that container
	@docker-compose run -e GOPATH=$(GOPATH) -w /code build-go make inner-test

# this target is hidden, only meant to be invoked inside the build container
$(BINDATA):
	go-bindata $(BINDATA_FLAGS) -o=$@ src/despite/data/...

$(BUNDLE): $(APP)
	@docker-compose run -w /code build-node make inner-bundle

# this target is hidden, only meant to be invoked inside the build container
inner-bundle:
	@npm install
	@$(NODE_BIN)/webpack --progress --colors --bail

# this target is hidden, only meant to be invoked inside the build container
inner-test:
	go env
	go test -v despite

# the stuff to the right of the pipe symbol is order-only prerequisites
xbuild: $(BUNDLE) $(BINDATA) | check-deps ## cross-compile using xgo in docker
	xgo --targets=$(XGO_TARGETS) -ldflags "-X main.tag=$(CIRCLE_TAG) -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse HEAD`" $(GOPATH)/src/despite

build: XGO_TARGETS=darwin-10.9/amd64
build: xbuild ## build darwin/amd64 only (faster for local dev)

build-container: | check-deps ## build & upload our go & npm build containers
	docker build -t kindlyops/golang go-build-image
	docker push kindlyops/golang
	cd npm-build-image && docker build -t kindlyops/node .
	docker push kindlyops/node

shasums:
	@sudo sh -c 'sha256sum bin/* > bin/SHA256_SUMS.txt'

inner-prerelease:
	@ghr -r despite --username $(GITHUB_USER) --token $(GITHUB_TOKEN) --replace --prerelease --debug pre-release $(CIRCLE_ARTIFACTS)

inner-release:
	@ghr -r despite --username $(GITHUB_USER) --token $(GITHUB_TOKEN) --debug $(CIRCLE_TAG) $(CIRCLE_ARTIFACTS)

prerelease: shasums | check-deps
	@docker-compose run -e GITHUB_TOKEN=$(GITHUB_TOKEN) -e GITHUB_USER=$(GITHUB_USER) -e CIRCLE_ARTIFACTS=$(CIRCLE_ARTIFACTS) -w /code build-go make inner-prerelease

release: shasums | check-deps
	@docker-compose run -e GITHUB_TOKEN=$(GITHUB_TOKEN) -e GITHUB_USER=$(GITHUB_USER) -e CIRCLE_TAG=$(CIRCLE_TAG) -e CIRCLE_ARTIFACTS=$(CIRCLE_ARTIFACTS) -w /code build-go make inner-release

homebrew: | check-deps
	@git clone git@github.com:kindlyops/homebrew-tap.git
	@erb version=$(CIRCLE_TAG) packaging-templates/despite.rb.erb > homebrew-tap/despite.rb
	@git config --global user.name "CircleCI"
	@git config --global user.email "statik@users.noreply.github.com"
	@cd homebrew-tap && git commit -am "Releasing $(CIRCLE_TAG)" && git push origin master


# cleverness from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show the help for this makefile
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

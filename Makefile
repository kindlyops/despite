.DEFAULT_GOAL := help
.PHONY: help test
docker := $(shell command -v docker 2> /dev/null)
docker-compose := $(shell command -v docker-compose 2> /dev/null)
GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')

check-deps: ## Check if we have required dependencies
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
	@docker-compose run -w /code build make inner-test

# this target is hidden, only meant to be invoked inside the build container
inner-test:
	gb test -v

# this target is hidden, only meant to be invoked inside the build container
inner-build:
	@echo GOOS=$(GOOS) GOARCH=$(GOARCH)
	gb build -ldflags "-X main.tag=$(CIRCLE_TAG) -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse HEAD`" all;

# the stuff to the right of the pipe symbol is order-only prerequisites
build: | check-deps ## Compile using a docker build container
	@docker-compose run -e CIRCLE_TAG=$(CIRCLE_TAG) -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -w /code build make inner-build


image: | check-deps ## build & upload our go build container
	docker build -t kindlyops/golang build-image
	docker push kindlyops/golang

shasums:
	@sudo sh -c 'sha256sum bin/* > bin/SHA256_SUMS.txt'

inner-prerelease:
	@ghr -r despite --username $(GITHUB_USER) --token $(GITHUB_TOKEN) --replace --prerelease --debug pre-release bin/

inner-release:
	@ghr -r despite --username $(GITHUB_USER) --token $(GITHUB_TOKEN) --debug $(CIRCLE_TAG) bin/

prerelease: shasums | check-deps
	@docker-compose run -e GITHUB_TOKEN=$(GITHUB_TOKEN) -e GITHUB_USER=$(GITHUB_USER) -w /code build make inner-prerelease

release: shasums | check-deps
	@docker-compose run -e GITHUB_TOKEN=$(GITHUB_TOKEN) -e GITHUB_USER=$(GITHUB_USER) -e CIRCLE_TAG=$(CIRCLE_TAG) -w /code build make inner-release

homebrew: | check-deps
	@git clone git@github.com:kindlyops/homebrew-tap.git
	@erb version=$(CIRCLE_TAG) packaging-templates/despite.rb.erb > homebrew-tap/despite.rb
	@cd homebrew-tap && git commit -a -c user.name='CircleCI' -c user.email='statik@users.noreply.github.com' -m "Releasing $(CIRCLE_TAG)" && git push origin master


# cleverness from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show the help for this makefile
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

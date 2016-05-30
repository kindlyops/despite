.DEFAULT_GOAL := help
.PHONY: help test
uname := $(shell uname -s)
docker := $(shell command -v docker 2> /dev/null)
docker-compose := $(shell command -v docker-compose 2> /dev/null)

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
	gb build all;

# the stuff to the right of the pipe symbol is order-only prerequisites
build: | check-deps ## Compile using a docker build container
	@docker-compose run -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -w /code build make inner-build


image: | check-deps ## build & upload our go build container
	docker build -t kindlyops/golang build-image
	docker push kindlyops/golang

# cleverness from http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show the help for this makefile
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

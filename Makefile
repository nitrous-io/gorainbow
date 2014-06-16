image:
	@docker build -t registry.nitrous.io/gorainbow .
.PHONY: image

shell: image
	@./script/shell
.PHONY: shell

setup:
	@rm -rf _vendor
	@./script/install-deps
.PHONY: setup

run-local:
	@go install github.com/nitrous-io/gorainbow/gorainbow
	@gorainbow -c config.json
.PHONY: run-local

test: image
	@./script/test
.PHONY: test

build: image
	@./script/build
.PHONY: build

release: test build
	@./script/release

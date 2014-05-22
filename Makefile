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

run:
	@./script/install
	@compass
.PHONY: setup

test: image
	@./script/test
.PHONY: test

build: image
	@./script/build
.PHONY: build

release: build
	@./script/release

.PHONY: \
	all \
	build-linux \
	build-osx \
	package \
	govet \
	golint \
	gofmt \
	lint

export GO15VENDOREXPERIMENT=1

GOPATH ?= $(GOPATH:)
GOFLAGS ?= $(GOFLAGS:)
DIST := dist
VERSION := $(shell git describe --tags)
GO := go

GO_LINKER_FLAGS ?= -ldflags "-X main.Version=${VERSION}"

all: build-linux build-osx

build-linux:
	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(DIST)/linux/mattermail/mattermail $(GOFLAGS) $(GO_LINKER_FLAGS) *.go

build-osx:
	@echo Build OSX amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(DIST)/osx/mattermail/mattermail $(GOFLAGS) $(GO_LINKER_FLAGS) *.go

package:
	@echo Create Linux package
	cp config.json $(DIST)/linux/mattermail/
	tar -C $(DIST)/linux -czf $(DIST)/mattermail-$(VERSION).linux.am64.tar.gz mattermail

	@echo Create OSX package
	cp config.json $(DIST)/osx/mattermail/
	tar -C $(DIST)/osx -czf $(DIST)/mattermail-$(VERSION).osx.am64.tar.gz mattermail

govet:
	@echo GOVET
	$(shell go vet ./*.go)

golint:
	@echo GOLINT
	@golint ./*.go

gofmt:
	@echo GOFMT
	$(eval GOFMT_OUTPUT := $(shell gofmt -d -s *.go 2>&1))
	@echo "$(GOFMT_OUTPUT)"
	@if [ ! "$(GOFMT_OUTPUT)" ]; then \
		echo "gofmt sucess"; \
	else \
		echo "gofmt failure"; \
		exit 1; \
	fi

lint: govet golint gofmt

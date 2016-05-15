.PHONY: all build-linux build-osx package

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

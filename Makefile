.PHONY: all build-linux build-osx package

GOPATH ?= $(GOPATH:)
GOFLAGS ?= $(GOFLAGS:)
DIST=dist

GO=$(GOPATH)/bin/godep go

all: build-linux build-osx

build-linux:
	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 $(GO) install $(GOFLAGS) ./...

build-osx:
	@echo Build OSX amd64
	env GOOS=darwin GOARCH=amd64 $(GO) install $(GOFLAGS) ./...

package:
	@echo Clear dist folder
	rm -fr $(DIST)

	@echo Create Linux package
	mkdir -p $(DIST)/linux/mattermail
	cp $(GOPATH)/bin/mattermail $(DIST)/linux/mattermail/
	cp config.json $(DIST)/linux/mattermail/
	tar -C $(DIST)/linux -czf $(DIST)/mattermail.linux.am64.tar.gz mattermail

	@echo Create OSX package
	mkdir -p $(DIST)/osx/mattermail
	cp $(GOPATH)/bin/darwin_amd64/mattermail $(DIST)/osx/mattermail/
	cp config.json $(DIST)/osx/mattermail/
	tar -C $(DIST)/osx -czf $(DIST)/mattermail.osx.am64.tar.gz mattermail

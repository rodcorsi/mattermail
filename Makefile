.PHONY: \
	all \
	build-linux \
	build-osx \
	package \
	govet \
	golint \
	gofmt \
	format \
	lint \
	test \
	cover

GOPATH ?= $(GOPATH:)
GOFLAGS ?= $(GOFLAGS:)
DIST := dist
VERSION := $(shell git describe --tags)
GO := go

GO_LINKER_FLAGS ?= -ldflags "-X github.com/rodcorsi/mattermail/cmd.Version=${VERSION}"

all: build-linux build-osx build-windows

build-linux:
	@echo Build Linux amd64
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(DIST)/linux/mattermail/mattermail $(GOFLAGS) $(GO_LINKER_FLAGS) *.go

build-osx:
	@echo Build OSX amd64
	env GOOS=darwin GOARCH=amd64 $(GO) build -o $(DIST)/osx/mattermail/mattermail $(GOFLAGS) $(GO_LINKER_FLAGS) *.go

build-windows:
	@echo Build Windows amd64
	env GOOS=windows GOARCH=amd64 $(GO) build -o $(DIST)/windows/mattermail/mattermail $(GOFLAGS) $(GO_LINKER_FLAGS) *.go

package:
	@echo Create Linux package
	cp config.json $(DIST)/linux/mattermail/
	tar -C $(DIST)/linux -czf $(DIST)/mattermail-$(VERSION).linux.am64.tar.gz mattermail

	@echo Create OSX package
	cp config.json $(DIST)/osx/mattermail/
	tar -C $(DIST)/osx -czf $(DIST)/mattermail-$(VERSION).osx.am64.tar.gz mattermail

	@echo Create Windows package
	cp config.json $(DIST)/windows/mattermail/
	tar -C $(DIST)/windows -czf $(DIST)/mattermail-$(VERSION).windows.am64.tar.gz mattermail

govet:
	@echo GOVET
	$(eval PKGS := $(shell go list ./... | grep -v /vendor/))
	@$(GO) vet $(PKGS)

golint:
	@echo GOLINT
	$(eval PKGS := $(shell go list ./... | grep -v /vendor/))
	@for pkg in $(PKGS) ; do \
		golint -set_exit_status $$pkg; \
	done

gofmt:
	@echo GOFMT
	@mkdir -p $(DIST)
	@find ./ -type f -name "*.go" -not -path "./vendor/*" -exec gofmt -s -d {} \; | tee $(DIST)/format.diff
	@test ! -s $(DIST)/format.diff || { echo "ERROR: the source code has not been formatted - please use 'make format' or 'gofmt'"; exit 1; }

format:
	@find ./ -type f -name "*.go" -not -path "./vendor/*" -exec gofmt -w {} \;

lint: govet golint gofmt

test:
	@echo Running tests
	$(eval PKGS := $(shell go list ./... | grep -v /vendor/))
	$(eval PKGS_DELIM := $(shell echo $(PKGS) | sed -e 's/ /,/g'))
	$(GO) list -f '{{if or (len .TestGoFiles) (len .XTestGoFiles)}}$(GO) test -run=$(TESTS) -test.v -test.timeout=120s -covermode=count -coverprofile={{.Name}}_{{len .Imports}}_{{len .Deps}}.coverprofile -coverpkg $(PKGS_DELIM) {{.ImportPath}}{{end}}' $(PKGS) | xargs -I {} bash -c {}
	gocovmerge `ls *.coverprofile` > cover.out
	rm *.coverprofile

cover:
	@echo Opening coverage info on browser. If this failed run make test first

	$(GO) tool cover -html=cover.out
	$(GO) tool cover -func=cover.out

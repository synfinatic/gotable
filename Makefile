DIST_DIR ?= dist/
GOOS ?= $(shell uname -s | tr "[:upper:]" "[:lower:]")
ARCH ?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
GOARCH             := amd64
else
GOARCH             := $(ARCH)  # no idea if this works for other platforms....
endif
BUILDINFOSDET ?=
PROGRAM_ARGS ?=

PROJECT_VERSION           := 0.0.1
DOCKER_REPO               := synfinatic
PROJECT_NAME              := gotable
PROJECT_TAG               := $(shell git describe --tags 2>/dev/null $(git rev-list --tags --max-count=1))
ifeq ($(PROJECT_TAG),)
PROJECT_TAG               := NO-TAG
endif
PROJECT_COMMIT            := $(shell git rev-parse HEAD)
ifeq ($(PROJECT_COMMIT),)
PROJECT_COMMIT            := NO-CommitID
endif
PROJECT_DELTA             := $(shell DELTA_LINES=$$(git diff | wc -l); if [ $${DELTA_LINES} -ne 0 ]; then echo $${DELTA_LINES} ; else echo "''" ; fi)
VERSION_PKG               := $(shell echo $(PROJECT_VERSION) | sed 's/^v//g')
LICENSE                   := BSD 3-Clause
URL                       := https://github.com/$(DOCKER_REPO)/$(PROJECT_NAME)
DESCRIPTION               := Simple ASCII tables for Go structs
BUILDINFOS                := $(shell date +%FT%T%z)$(BUILDINFOSDET)
HOSTNAME                  := $(shell hostname)
LDFLAGS                   := -X "main.Version=$(PROJECT_VERSION)" -X "main.Delta=$(PROJECT_DELTA)" -X "main.Buildinfos=$(BUILDINFOS)" -X "main.Tag=$(PROJECT_TAG)" -X "main.CommitID=$(PROJECT_COMMIT)"

clean: ## Clean Go cache
	go clean -i -r -cache -modcache

go-get:  ## Get our go modules
	go get -v all

.PHONY: unittest
unittest: ## Run go unit tests
	go test .

.PHONY: vet
vet: ## Run `go vet` on the code
	@echo checking code is vetted...
	go vet .

test: vet unittest ## Run all tests

.PHONY: fmt
fmt: ## Format Go code
	@go fmt .

.PHONY: test-fmt
test-fmt: fmt ## Test to make sure code if formatted correctly
	@if test `git diff . | wc -l` -gt 0; then \
	    echo "Code changes detected when running 'go fmt':" ; \
	    git diff -Xfiles ; \
	    exit -1 ; \
	fi

.PHONY: test-tidy
test-tidy:  ## Test to make sure go.mod is tidy
	@go mod tidy
	@if test `git diff go.mod | wc -l` -gt 0; then \
	    echo "Need to run 'go mod tidy' to clean up go.mod" ; \
	    exit -1 ; \
	fi

precheck: test test-fmt test-tidy  ## Run all tests that happen in a PR

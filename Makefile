#!/usr/bin/make -f

###############################################################################
###                           Module & Versioning                           ###
###############################################################################

VERSION ?= $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

###############################################################################
###                          Directories & Binaries                         ###
###############################################################################

BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
BINARY := sirrmeshd

###############################################################################
###                              Repo Info                                  ###
###############################################################################

HTTPS_GIT := https://github.com/sirrchat/sirrmesh.git
DOCKER := $(shell which docker)

export GO111MODULE = on

###############################################################################
###                            Build Settings                               ###
###############################################################################

MAIN_PKG := ./cmd/sirrmeshd

# process build tags
build_tags = netgo
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

# process linker flags
ldflags = -X github.com/mail-chat-chain/sirrmeshd/config.Version=$(VERSION)

ifeq (,$(findstring nostrip,$(BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

ifeq (,$(findstring nostrip,$(BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# check if no optimization option is passed (for remote debugging)
ifneq (,$(findstring nooptimization,$(BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

###############################################################################
###                        Build & Install                                  ###
###############################################################################

# Build into $(BUILDDIR)
build: go.sum $(BUILDDIR)/
	@echo "Building sirrmeshd to $(BUILDDIR)/$(BINARY) ..."
	@CGO_ENABLED="1" go build $(BUILD_FLAGS) -o $(BUILDDIR)/$(BINARY) $(MAIN_PKG)

# Cross-compile for Linux AMD64
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

# Cross-compile for Linux ARM64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build

# Cross-compile for macOS AMD64
build-darwin:
	GOOS=darwin GOARCH=amd64 $(MAKE) build

# Cross-compile for macOS ARM64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(MAKE) build

# Install into $(BINDIR)
install: go.sum
	@echo "Installing sirrmeshd to $(BINDIR) ..."
	@CGO_ENABLED="1" go install $(BUILD_FLAGS) $(MAIN_PKG)

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

# Default & all target
.PHONY: all build build-linux build-linux-arm64 build-darwin build-darwin-arm64 install
all: build

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	@echo "Ensure dependencies have not been modified ..."
	@go mod verify
	@go mod tidy

vulncheck:
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

###############################################################################
###                           Tests                                         ###
###############################################################################

PACKAGES := $(shell go list ./... | grep -v '/tests/')
TEST_PACKAGES := ./...

test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test -tags=test -mod=readonly -timeout=15m $(PACKAGES)

test-race:
	@echo "Running tests with race detection..."
	@go test -tags=test -mod=readonly -race -timeout=15m $(PACKAGES)

test-cover:
	@echo "Running tests with coverage..."
	@go test -tags=test -mod=readonly -timeout=15m -coverprofile=coverage.txt -covermode=atomic $(PACKAGES)
	@go tool cover -func=coverage.txt

.PHONY: test test-unit test-race test-cover

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=golangci-lint
golangci_version=v2.2.2

lint: lint-go

lint-go:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m

lint-fix:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m --fix

.PHONY: lint lint-fix lint-go

###############################################################################
###                              Formatting                                 ###
###############################################################################

format: format-go format-shell

format-go:
	@echo "Formatting Go files..."
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofumpt -w -l

format-shell:
	@echo "Formatting shell files..."
	@shfmt -l -w . 2>/dev/null || true

.PHONY: format format-go format-shell

###############################################################################
###                                Clean                                    ###
###############################################################################

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILDDIR)
	@rm -f coverage.txt

.PHONY: clean

###############################################################################
###                                Releasing                                ###
###############################################################################

PACKAGE_NAME := github.com/sirrchat/sirrmesh
GOLANG_CROSS_VERSION = v1.22

release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		--clean --skip validate --skip publish --snapshot

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --skip validate

.PHONY: release-dry-run release

###############################################################################
###                                 Help                                    ###
###############################################################################

help:
	@echo "Available targets:"
	@echo "  build           - Build the sirrmeshd binary"
	@echo "  build-linux     - Cross-compile for Linux AMD64"
	@echo "  build-linux-arm64 - Cross-compile for Linux ARM64"
	@echo "  build-darwin    - Cross-compile for macOS AMD64"
	@echo "  build-darwin-arm64 - Cross-compile for macOS ARM64"
	@echo "  install         - Install sirrmeshd to GOPATH/bin"
	@echo "  test            - Run unit tests"
	@echo "  test-race       - Run tests with race detection"
	@echo "  test-cover      - Run tests with coverage"
	@echo "  lint            - Run linter"
	@echo "  lint-fix        - Run linter and fix issues"
	@echo "  format          - Format code"
	@echo "  clean           - Clean build artifacts"
	@echo "  vulncheck       - Check for vulnerabilities"
	@echo "  help            - Show this help"

.PHONY: help

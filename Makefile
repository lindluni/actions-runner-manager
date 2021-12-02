#
# SPDX-License-Identifier: Apache-2.0
#

GOTOOLS = counterfeiter gofumpt goimports golint staticcheck
BUILD_DIR ?= build
GOTOOLS_BINDIR ?= $(shell go env GOBIN)

go.fqp.counterfeiter := github.com/maxbrunsfeld/counterfeiter/v6
go.fqp.gofumpt       := mvdan.cc/gofumpt
go.fqp.goimports     := golang.org/x/tools/cmd/goimports
go.fqp.golint        := golang.org/x/lint/golint
go.fqp.staticcheck   := honnef.co/go/tools/cmd/staticcheck

.PHONY: lint
lint: tools
	./scripts/linter.sh

.PHONY: profile
profile:
	go test -coverprofile=c.out ./...
	go tool cover -html=c.out

.PHONY: clean
clean:
	rm -rf c.out

.PHONY: tests
tests: unit-tests

.PHONY: tools
tools: $(patsubst %,$(GOTOOLS_BINDIR)/%, $(GOTOOLS))

.PHONY: unit-tests
unit-tests:
	go test -cover ./...

gotool.%:
	$(eval TOOL = ${subst gotool.,,${@}})
	@echo "Building ${go.fqp.${TOOL}} -> $(TOOL)"
	@cd tools && GO111MODULE=on GOBIN=$(abspath $(GOTOOLS_BINDIR)) go install ${go.fqp.${TOOL}}

$(GOTOOLS_BINDIR)/%:
	$(eval TOOL = ${subst $(GOTOOLS_BINDIR)/,,${@}})
	@$(MAKE) gotool.$(TOOL)
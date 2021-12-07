#
# SPDX-License-Identifier: Apache-2.0
#

GOTOOLS = counterfeiter gofumpt goimports golint staticcheck swagger swag
BUILD_DIR ?= build
GOTOOLS_BINDIR ?= $(shell go env GOBIN)

go.fqp.counterfeiter := github.com/maxbrunsfeld/counterfeiter/v6
go.fqp.gofumpt       := mvdan.cc/gofumpt
go.fqp.goimports     := golang.org/x/tools/cmd/goimports
go.fqp.golint        := golang.org/x/lint/golint
go.fqp.staticcheck   := honnef.co/go/tools/cmd/staticcheck
go.fqp.swagger       := github.com/go-swagger/go-swagger/cmd/swagger
go.fqp.swag   		 := github.com/swaggo/swag/cmd/swag

.PHONY: clean
clean:
	rm -rf c.out dist

.PHONY: docker
docker:
	docker build -t ghcr.io/lindluni/actions-runner-manager .

.PHONY: integration-tests
integration-tests:
	go test -cover ./integration/...

.PHONY: lint
lint: tools
	./scripts/linter.sh

.PHONY: mocks
mocks: tools
	go generate ./...

.PHONY: profile-unit-tests
profile:
	go test -coverprofile=c.out ./pkg/...
	go tool cover -html=c.out

.PHONY: release
release:
	go build -o dist/actions-runner-manager ./pkg

.PHONY: swagger
swagger:
	swag fmt --dir pkg
	swag init --parseDependency --parseInternal --parseDepth 1 --dir pkg

.PHONY: tests
tests: unit-tests integration-tests

.PHONY: tools
tools: $(patsubst %,$(GOTOOLS_BINDIR)/%, $(GOTOOLS))

.PHONY: unit-tests
unit-tests:
	go test -p 4 -cover ./pkg/...

gotool.%:
	$(eval TOOL = ${subst gotool.,,${@}})
	@echo "Building ${go.fqp.${TOOL}} -> $(TOOL)"
	@cd tools && GO111MODULE=on GOBIN=$(abspath $(GOTOOLS_BINDIR)) go install ${go.fqp.${TOOL}}

$(GOTOOLS_BINDIR)/%:
	$(eval TOOL = ${subst $(GOTOOLS_BINDIR)/,,${@}})
	@$(MAKE) gotool.$(TOOL)
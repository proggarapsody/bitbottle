.PHONY: build test test-scripts lint clean setup gen

BIN := bitbottle
GO  := go

# Pinned oapi-codegen version — update here and re-run `make gen` when bumping.
OAPI_CODEGEN_VERSION := v2.4.1

build:
	$(GO) build -o $(BIN) ./cmd/bitbottle

test:
	$(GO) test ./... -race

test-script:
	$(GO) test ./test/script/... -race

test-scripts:
	@set -e; for t in auto-iter/scripts/*_test.sh; do echo "--- $$t ---"; bash "$$t" || exit 1; done

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)

setup:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/pre-push
	@echo "Git hooks active. pre-commit: gofmt. pre-push: golangci-lint."

# gen regenerates spec-derived Go types from the OpenAPI specs.
# Requires oapi-codegen $(OAPI_CODEGEN_VERSION) on PATH.
# Install: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)
gen:
	@command -v oapi-codegen >/dev/null 2>&1 || \
		(echo "Installing oapi-codegen $(OAPI_CODEGEN_VERSION)..." && \
		 go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION))
	oapi-codegen -config api/cloud/gen/oapi-codegen.yaml -o api/cloud/gen/types.go api/cloud/gen/openapi.yaml
	oapi-codegen -config api/server/gen/oapi-codegen.yaml -o api/server/gen/types.go api/server/gen/openapi.yaml

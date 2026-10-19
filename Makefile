.PHONY: build test test-script test-scripts test-acceptance coverage coverage-report qa-matrix qa-matrix-check lint clean setup gen record-cassettes

BIN := bitbottle
GO  := go
GO_TEST_FLAGS := -race -shuffle=on -count=1
GO_COVERPROFILE ?= coverage.out
QA_MATRIX_DIR ?= .qa-matrix

# Pinned oapi-codegen version — update here and re-run `make gen` when bumping.
OAPI_CODEGEN_VERSION := v2.4.1

build:
	$(GO) build -o $(BIN) ./cmd/bitbottle

test:
	$(GO) test $(GO_TEST_FLAGS) ./...

coverage:
	$(GO) test $(GO_TEST_FLAGS) -coverprofile="$(GO_COVERPROFILE)" ./...

coverage-report:
	$(GO) tool cover -func="$(GO_COVERPROFILE)"

qa-matrix:
	python3 scripts/qa_matrix.py --run --packages ./... --coverprofile "$(GO_COVERPROFILE)" --output "$(QA_MATRIX_DIR)"

qa-matrix-check:
	python3 scripts/qa_matrix.py --check --output "$(QA_MATRIX_DIR)"

test-script:
	$(GO) test $(GO_TEST_FLAGS) ./test/script/...

test-scripts:
	@set -e; for t in auto-iter/scripts/*_test.sh scripts/run-acceptance_test.sh; do echo "--- $$t ---"; bash "$$t" || exit 1; done
	python3 -m unittest discover -s scripts -p 'qa_matrix_test.py'

test-acceptance:
	bash scripts/run-acceptance.sh

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)

setup:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/pre-push
	@echo "Git hooks active. pre-commit: gofmt. pre-push: golangci-lint."

# record-cassettes: Re-record VCR cassettes against the real Bitbucket backend.
# Requires: BITBOTTLE_TOKEN, BITBOTTLE_HOST, BITBOTTLE_PROJECT, BITBOTTLE_REPO env vars.
# Run once with real credentials; commit the resulting YAML files.
record-cassettes:
	BITBOTTLE_RECORD=1 $(GO) test ./... -run Integration -count=1 -v

# gen regenerates spec-derived Go types from the OpenAPI specs.
# Requires oapi-codegen $(OAPI_CODEGEN_VERSION) on PATH.
# Install: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)
gen:
	@command -v oapi-codegen >/dev/null 2>&1 || \
		(echo "Installing oapi-codegen $(OAPI_CODEGEN_VERSION)..." && \
		 go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION))
	oapi-codegen -config api/cloud/gen/oapi-codegen.yaml -o api/cloud/gen/types.go api/cloud/gen/openapi.yaml
	oapi-codegen -config api/server/gen/oapi-codegen.yaml -o api/server/gen/types.go api/server/gen/openapi.yaml

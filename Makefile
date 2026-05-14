.PHONY: build test test-scripts lint clean setup

BIN := bitbottle
GO  := go

build:
	$(GO) build -o $(BIN) ./cmd/bitbottle

test:
	$(GO) test ./... -race

test-scripts:
	@set -e; for t in scripts/auto-iter/*_test.sh; do echo "--- $$t ---"; bash "$$t" || exit 1; done

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)

setup:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/pre-push
	@echo "Git hooks active. pre-commit: gofmt. pre-push: golangci-lint."

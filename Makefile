.PHONY: build test lint clean setup doc-graph

BIN := bitbottle
GO  := go

build:
	$(GO) build -o $(BIN) ./cmd/bitbottle

test:
	$(GO) test ./... -race

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)

setup:
	git config core.hooksPath .githooks

doc-graph:
	@./scripts/doc-graph.sh

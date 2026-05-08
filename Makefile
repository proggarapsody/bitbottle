.PHONY: build test lint clean setup

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

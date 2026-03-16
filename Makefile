APP_NAME=go-skylight
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

vet:
	go vet ./...

build:
	go build $(LDFLAGS) -o $(APP_NAME)

lint:
	golangci-lint run ./...

test:
	go test ./... -v

build-trigger:
	go build -o alpaca-trigger ./cmd/alpaca-trigger/

clean:
	rm -f $(APP_NAME) alpaca-trigger

.PHONY: vet build build-trigger lint test clean

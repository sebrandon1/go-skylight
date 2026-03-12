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

clean:
	rm -f $(APP_NAME)

.PHONY: vet build lint test clean

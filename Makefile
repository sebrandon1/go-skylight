APP_NAME=go-skylight

vet:
	go vet ./...

build:
	go build -o $(APP_NAME)

lint:
	golangci-lint run ./...

test:
	go test ./... -v

clean:
	rm -f $(APP_NAME)

.PHONY: vet build lint test clean

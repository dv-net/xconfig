lint:
	golangci-lint run

test:
	go test -v ./...

all: lint test

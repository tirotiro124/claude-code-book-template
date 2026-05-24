VERSION ?= dev
BINARY  = task

.PHONY: build test lint fmt bench clean

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) .

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -f $(BINARY) coverage.out

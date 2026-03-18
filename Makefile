BINARY  := gmt
VERSION := 0.2.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)

.PHONY: build test vet lint clean

build: lint
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	golangci-lint run ./...

clean:
	rm -f $(BINARY)

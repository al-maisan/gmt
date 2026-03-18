BINARY  := gmt
VERSION := 0.2.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.appVersion=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)

.PHONY: all build test vet lint fmt install clean

all: test build

build: lint
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	golangci-lint run ./...

fmt:
	gofmt -w .

install: lint
	go install -ldflags "$(LDFLAGS)"

clean:
	rm -f $(BINARY)

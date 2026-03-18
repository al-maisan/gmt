BINARY  := gmt
VERSION := 0.2.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)

.PHONY: build test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

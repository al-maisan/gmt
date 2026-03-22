BINARY  := gmt-mail
VERSION := 0.2.2
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.appVersion=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)

RPMBUILD_DIR := /tmp/gmt-rpmbuild
TARBALL      := $(RPMBUILD_DIR)/SOURCES/gmt-$(VERSION).tar.gz

.PHONY: all build test vet lint fmt install clean srpm

all: test build

build: lint
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

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

srpm: vendor
	mkdir -p $(RPMBUILD_DIR)/{SOURCES,SPECS,SRPMS}
	tar czf $(TARBALL) --transform 's,^\.,gmt-$(VERSION),' \
		--exclude='./.git' --exclude='./gmt-mail.spec' \
		--exclude='./ai' --exclude='./.claude' \
		--exclude='./$(BINARY)' --exclude='./ppa' .
	rpmbuild -bs gmt-mail.spec --define "_topdir $(RPMBUILD_DIR)"
	@echo "SRPM: $$(ls $(RPMBUILD_DIR)/SRPMS/gmt-mail-$(VERSION)-*.src.rpm)"

vendor:
	go mod vendor

clean:
	rm -f $(BINARY)
	rm -rf $(RPMBUILD_DIR)

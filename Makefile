BINARY  := gmt-mail
VERSION := 0.5.1
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.appVersion=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)
TAG     := v$(VERSION)

RPMBUILD_DIR := /tmp/gmt-rpmbuild
TARBALL      := $(RPMBUILD_DIR)/SOURCES/gmt-$(VERSION).tar.gz

.PHONY: all build test vet lint fmt install clean srpm vendor tag release copr ppa

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

vendor:
	go mod vendor

tag:
	@if git rev-parse $(TAG) >/dev/null 2>&1; then \
		echo "Tag $(TAG) already exists"; \
	else \
		git tag -s $(TAG) -m "$(TAG)"; \
		git push origin $(TAG); \
		echo "Created and pushed tag $(TAG)"; \
	fi

release: tag
	gh release create $(TAG) --title "$(TAG)" --generate-notes

srpm: vendor rpm-changelog
	mkdir -p $(RPMBUILD_DIR)/{SOURCES,SPECS,SRPMS}
	tar czf $(TARBALL) --transform 's,^\.,gmt-$(VERSION),' \
		--exclude='./.git' --exclude='./gmt-mail.spec' \
		--exclude='./ai' --exclude='./.claude' \
		--exclude='./$(BINARY)' --exclude='./ppa' .
	rpmbuild -bs gmt-mail.spec --define "_topdir $(RPMBUILD_DIR)"
	@echo "SRPM: $$(ls $(RPMBUILD_DIR)/SRPMS/gmt-mail-$(VERSION)-*.src.rpm)"

rpm-changelog:
	@./scripts/gen-rpm-changelog.sh > /tmp/rpm-changelog.tmp
	@sed -i '/^%changelog/,$$d' gmt-mail.spec
	@echo '%changelog' >> gmt-mail.spec
	@cat /tmp/rpm-changelog.tmp >> gmt-mail.spec
	@rm -f /tmp/rpm-changelog.tmp

copr: srpm
	copr-cli build gmt-mail $$(ls $(RPMBUILD_DIR)/SRPMS/gmt-mail-$(VERSION)-*.src.rpm)

ppa: tag
	ppa/build-ppa.sh

publish: release copr ppa
	@echo "Published $(TAG) to GitHub, COPR, and PPA"

clean:
	rm -f $(BINARY)
	rm -rf $(RPMBUILD_DIR)

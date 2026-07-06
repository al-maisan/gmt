BINARY  := gmt-mail
# VERSION is derived from the latest git tag (the single source of truth for a
# release). Cut a release by creating the tag, e.g.:
#   git tag -s v0.7.0 -m "v0.7.0" && git push origin v0.7.0
# Override for a one-shot cut before the tag exists: make release VERSION=0.7.0
GIT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
VERSION     := $(patsubst v%,%,$(or $(GIT_VERSION),v0.0.0))
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.appVersion=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildDate=$(DATE)
TAG     := v$(VERSION)

RPMBUILD_DIR := /tmp/gmt-rpmbuild
TARBALL      := $(RPMBUILD_DIR)/SOURCES/gmt-$(VERSION).tar.gz
SPEC_TEMPLATE := gmt-mail.spec
SPEC_OUT      := $(RPMBUILD_DIR)/SPECS/gmt-mail.spec

.PHONY: all build test vet lint fmt install clean srpm vendor tag release copr ppa gen-spec

all: test lint build

build:
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
		echo "tag: $(TAG) exists locally"; \
	else \
		git tag -s $(TAG) -m "$(TAG)"; \
		echo "tag: created $(TAG)"; \
	fi
	@if git ls-remote --tags origin $(TAG) | grep -q $(TAG); then \
		echo "tag: $(TAG) exists on remote"; \
	else \
		git push origin $(TAG); \
		echo "tag: pushed $(TAG)"; \
	fi

PREV_RELEASE := $(shell gh release list --limit 1 --json tagName --jq '.[0].tagName' 2>/dev/null)

release: tag
	@if gh release view $(TAG) >/dev/null 2>&1; then \
		echo "release: $(TAG) exists on GitHub"; \
	elif [ -n "$(PREV_RELEASE)" ]; then \
		gh release create $(TAG) --title "$(TAG)" --generate-notes --notes-start-tag $(PREV_RELEASE); \
		echo "release: created $(TAG) (diff from $(PREV_RELEASE))"; \
	else \
		gh release create $(TAG) --title "$(TAG)" --generate-notes; \
		echo "release: created $(TAG) (first release)"; \
	fi

srpm: vendor gen-spec
	mkdir -p $(RPMBUILD_DIR)/{SOURCES,SPECS,SRPMS}
	tar czf $(TARBALL) --transform 's,^\.,gmt-$(VERSION),' \
		--exclude='./.git' --exclude='./gmt-mail.spec' \
		--exclude='./ai' --exclude='./.claude' \
		--exclude='./$(BINARY)' --exclude='./ppa' .
	rpmbuild -bs $(SPEC_OUT) --define "_topdir $(RPMBUILD_DIR)"
	@echo "SRPM: $$(ls $(RPMBUILD_DIR)/SRPMS/gmt-mail-$(VERSION)-*.src.rpm)"

# Generate the final spec into the build dir from the tracked template,
# substituting @VERSION@ and splicing in the git-derived %changelog. The tracked
# gmt-mail.spec is never modified, so it can never drift or dirty the tree.
gen-spec:
	@mkdir -p $(RPMBUILD_DIR)/SPECS
	@./scripts/gen-rpm-changelog.sh > /tmp/rpm-changelog.tmp
	@awk -v ver='$(VERSION)' -v clog=/tmp/rpm-changelog.tmp \
		'{ gsub(/@VERSION@/, ver) } /^@CHANGELOG@$$/ { while ((getline l < clog) > 0) print l; next } { print }' \
		$(SPEC_TEMPLATE) > $(SPEC_OUT)
	@rm -f /tmp/rpm-changelog.tmp
	@echo "spec: generated $(SPEC_OUT) (version $(VERSION))"

copr: srpm
	copr-cli build gmt-mail $$(ls $(RPMBUILD_DIR)/SRPMS/gmt-mail-$(VERSION)-*.src.rpm)

ppa: tag vendor
	ppa/build-ppa.sh

publish: release copr ppa
	@echo "Published $(TAG) to GitHub, COPR, and PPA"

clean:
	rm -f $(BINARY)
	rm -rf $(RPMBUILD_DIR)

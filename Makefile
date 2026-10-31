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

# Run gh with any ambient GITHUB_TOKEN removed so it uses the keyring login;
# a narrow-scoped GITHUB_TOKEN in the environment cannot create releases.
GH := env -u GITHUB_TOKEN gh

# Per-user build dir: a fixed /tmp path could be pre-created by another user on a
# shared host to tamper with the spec/tarball before rpmbuild and upload.
BUILD_ROOT   := $(or $(XDG_CACHE_HOME),$(HOME)/.cache)/gmt-mail
RPMBUILD_DIR := $(BUILD_ROOT)/rpmbuild
TARBALL      := $(RPMBUILD_DIR)/SOURCES/gmt-$(VERSION).tar.gz
SPEC_TEMPLATE := gmt-mail.spec
SPEC_OUT      := $(RPMBUILD_DIR)/SPECS/gmt-mail.spec
# gmt-mail.1 carries @VERSION@/@DATE@; the RPM spec and debian/rules fill them in from
# the package version and changelog date. `make man` renders it for local viewing.
MAN_OUT  := $(BUILD_ROOT)/man/gmt-mail.1
MAN_DATE := $(or $(shell git log -1 --format=%cs $(TAG) 2>/dev/null),$(DATE))

.PHONY: all build test vet lint fmt install clean srpm vendor tag release copr ppa test-ppa test-copr gen-spec man test-man

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

PREV_RELEASE := $(shell $(GH) release list --limit 1 --json tagName --jq '.[0].tagName' 2>/dev/null)

release: tag
	@if $(GH) release view $(TAG) >/dev/null 2>&1; then \
		echo "release: $(TAG) exists on GitHub"; \
	elif [ -n "$(PREV_RELEASE)" ]; then \
		$(GH) release create $(TAG) --title "$(TAG)" --generate-notes --notes-start-tag $(PREV_RELEASE) \
			&& echo "release: created $(TAG) (diff from $(PREV_RELEASE))"; \
	else \
		$(GH) release create $(TAG) --title "$(TAG)" --generate-notes \
			&& echo "release: created $(TAG) (first release)"; \
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
	@./scripts/gen-rpm-changelog.sh > $(RPMBUILD_DIR)/rpm-changelog.tmp
	@awk -v ver='$(VERSION)' -v clog=$(RPMBUILD_DIR)/rpm-changelog.tmp \
		'{ gsub(/@VERSION@/, ver) } /^@CHANGELOG@$$/ { while ((getline l < clog) > 0) print l; next } { print }' \
		$(SPEC_TEMPLATE) > $(SPEC_OUT)
	@rm -f $(RPMBUILD_DIR)/rpm-changelog.tmp
	@echo "spec: generated $(SPEC_OUT) (version $(VERSION))"

man:
	@mkdir -p $(dir $(MAN_OUT))
	@sed -e 's/@VERSION@/$(VERSION)/g' -e 's/@DATE@/$(MAN_DATE)/g' gmt-mail.1 > $(MAN_OUT)
	@echo "man: rendered $(MAN_OUT) (version $(VERSION), $(MAN_DATE)); view with: man -l $(MAN_OUT)"

test-man: man gen-spec
	scripts/man_test.sh $(MAN_OUT) $(SPEC_OUT) $(VERSION)

# Builds for every released, still-maintained Fedora that doesn't have this version
# yet (builds the SRPM only then). COPR_FORCE=1 rebuilds; COPR_CHROOTS="..." overrides.
copr:
	scripts/copr-build.sh $(VERSION) $(RPMBUILD_DIR)/SRPMS

test-copr:
	scripts/copr-build_test.sh
	scripts/gen-rpm-changelog_test.sh

# Uploads only releases that don't have this version yet; force a rebuild of an
# already-uploaded version with: make ppa PPA_REV=2
ppa: tag vendor
	ppa/build-ppa.sh $(PPA_REV)

test-ppa:
	ppa/build-ppa_test.sh

publish: release copr ppa
	@echo "Published $(TAG) to GitHub, COPR, and PPA"

clean:
	rm -f $(BINARY)
	rm -rf $(BUILD_ROOT)

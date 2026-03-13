.PHONY: all build install lint test ci docker.build docker.run docker.lint docker.test docker.ci jira.server release release.publish clean

##############
# Build vars #
##############

# https://git-scm.com/docs/git-stash#Documentation/git-stash.txt-create
#
# If uncommitted changes exist, then 'git stash create' will create a "stash
# entry" and print its object name; otherwise 'git stash create' will do
# nothing and print the empty string. In either case, 'git stash create'
# returns success.
#
# 'git rev-parse HEAD` (on success) prints the sha1sum of the current HEAD.
#
# Invoke both commands and take the first 40-xdigit string.
GIT_COMMIT ?= $(shell { git stash create; git rev-parse HEAD; } | grep -Exm1 '[[:xdigit:]]{40}')

# https://reproducible-builds.org/docs/source-date-epoch/
export SOURCE_DATE_EPOCH ?= $(shell git show -s --format="%ct" $(GIT_COMMIT))

VERSION ?= $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
VERSION_PKG = github.com/ankitpokhrel/jira-cli/internal/version
export LDFLAGS += -X $(VERSION_PKG).GitCommit=$(GIT_COMMIT)
export LDFLAGS += -X $(VERSION_PKG).SourceDateEpoch=$(SOURCE_DATE_EPOCH)
export LDFLAGS += -X $(VERSION_PKG).Version=$(VERSION)
export LDFLAGS += -s
export LDFLAGS += -w

export CGO_ENABLED ?= 0
export GOCACHE ?= $(CURDIR)/.gocache

DOCKER_GOLANG_IMAGE       ?= golang:1.25-alpine3.23
DOCKER_GOLANG_IMAGE_TEST  ?= golang:1.25
DOCKER_LINT_IMAGE         ?= golangci/golangci-lint:v2.6.2-alpine
DOCKER_RUN_DEV  = docker run --rm -v $(CURDIR):/app -w /app -e CGO_ENABLED -e GOCACHE=/app/.gocache $(DOCKER_GOLANG_IMAGE)
DOCKER_RUN_TEST = docker run --rm -v $(CURDIR):/app -w /app -e GOCACHE=/app/.gocache $(DOCKER_GOLANG_IMAGE_TEST)

all: build

build:
	$(DOCKER_RUN_DEV) sh -c "go mod vendor -v && go build -ldflags='$(LDFLAGS)' ./..."

install:
	go install -ldflags='$(LDFLAGS)' ./...

lint:
	docker run --rm -v $(CURDIR):/app -w /app $(DOCKER_LINT_IMAGE) golangci-lint run ./...

test:
	$(DOCKER_RUN_TEST) sh -c "CGO_ENABLED=1 go test -race ./..."

ci: lint test

docker.lint: lint
docker.test: test
docker.ci: ci

docker.build:
	docker build -t jira-cli:latest .

docker.run:
	docker run --rm jira-cli:latest $(ARGS)

jira.server:
	docker compose up -d

###########
# Release #
###########

RELEASE_VERSION ?= $(error RELEASE_VERSION is required — run: make release RELEASE_VERSION=vX.Y.Z)

release: ci
	@test -z "$$(git status --porcelain)" || { echo "ERROR: working tree is dirty"; exit 1; }
	@test "$$(git branch --show-current)" = "main" || { echo "ERROR: must be on main branch"; exit 1; }
	git tag -a $(RELEASE_VERSION) -m "Release $(RELEASE_VERSION)"
	git push origin $(RELEASE_VERSION)
	@echo "Tag $(RELEASE_VERSION) pushed — GitHub Actions release workflow triggered."

release.publish:
	gh release edit $(RELEASE_VERSION) --draft=false
	@echo "Release $(RELEASE_VERSION) published."

clean:
	$(DOCKER_RUN_DEV) go clean -x ./...

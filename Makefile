.PHONY: all deps build install lint test ci docker.build docker.run docker.lint docker.test docker.ci jira.server clean distclean

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

all: build

deps:
	go mod vendor -v

build: deps
	go build -ldflags='$(LDFLAGS)' ./...

install:
	go install -ldflags='$(LDFLAGS)' ./...

lint:
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b "$$(go env GOPATH)/bin" v2.6.2 ; \
	fi
	golangci-lint run ./...

test:
	@go clean -testcache
	CGO_ENABLED=1 go test -race ./...

ci: lint test

################
# Docker build #
################

docker.build:
	docker build -t jira-cli:latest .

docker.run:
	docker run --rm jira-cli:latest $(ARGS)

DOCKER_GOLANG_IMAGE ?= golang:1.25-alpine3.23
DOCKER_RUN_DEV = docker run --rm -v $(CURDIR):/app -w /app -e CGO_ENABLED -e GOCACHE=/app/.gocache $(DOCKER_GOLANG_IMAGE)

docker.lint:
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v2.6.2-alpine golangci-lint run ./...

docker.test:
	$(DOCKER_RUN_DEV) sh -c "apk add -U --no-cache build-base && CGO_ENABLED=1 go test -race ./..."

docker.ci: docker.lint docker.test

jira.server:
	docker compose up -d

clean:
	go clean -x ./...

distclean:
	go clean -x -cache -testcache -modcache ./...

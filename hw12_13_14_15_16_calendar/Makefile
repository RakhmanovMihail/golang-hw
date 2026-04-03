BIN_CALENDAR := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_STORER := "./bin/calendar_storer"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
export GOROOT := $(shell go env GOROOT)
export GOPATH := $(shell go env GOPATH)

build: build-calendar build-scheduler build-storer

build-calendar:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

build-storer:
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/calendar_storer

run: build-calendar
	$(BIN_CALENDAR) -config ./configs/config.toml

run-scheduler: build-scheduler
	$(BIN_SCHEDULER) -config ./configs/scheduler_config.toml

run-storer: build-storer
	$(BIN_STORER) -config ./configs/storer_config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build-calendar
	$(BIN_CALENDAR) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.11.4

lint: install-lint-deps
	golangci-lint run ./...

generate:
	oapi-codegen -generate types,chi-server,spec -package internalhttp -o internal/server/http/generated.go api/openapi.yaml

.PHONY: build build-calendar build-scheduler build-storer run run-scheduler run-storer build-img run-img version test lint generate

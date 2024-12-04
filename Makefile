GO=go
GO_VET_OPTS=-v
GO_TEST_OPTS=-v -race

GO_FMT=gofmt
GO_FMT_OPTS=-s -l

SIZE ?= 100m
FILE ?= ./data
PHONY: gen-file
gen-file:
	head -c $(SIZE) /dev/urandom > $(FILE)

.PHONY: fmt
fmt:
	$(GO_FMT) $(GO_FMT_OPTS) .

.PHONY: vet
vet:
	$(GO) vet $(GO_VET_OPTS) ./...

.PHONY: mod
mod:
	$(GO) mod tidy

.PHONY: build
build:
	$(GO) build $(GO_BUILD_OPT) -o ./bin/tcp-half-close-tester .

DOCKER=docker
DOCKER_FILE=./Dockerfile
DOCKER_REPO=ghcr.io/terassyi
DOCKER_TAG=dev
DOCKER_CONTEXT=.

.PHONY: docker
docker:
	$(DOCKER) image build -f $(DOCKER_FILE) -t $(DOCKER_REPO)/tcp-keepalive:$(DOCKER_TAG) $(DOCKER_CONTEXT)

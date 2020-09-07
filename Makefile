  
SHELL          := /bin/bash
BIN             = simple-job-queue
LDFLAGS        := "-s -w -X \"main.ldBuildDate=${RELEASE_TIME}\" -X main.ldGitCommit=${RELEASE_VER}"
LINTER          = https://install.goreleaser.com/github.com/golangci/golangci-lint.sh
LINTER_VERSION  = v1.24.0

all: lint test build

build:
	@echo "Building the binary"
	go build -a -o ./${BIN} -ldflags ${LDFLAGS}

test:
	go test -race ./...

lint:
	@echo "Linting"
	curl -sfL $(LINTER) | sh -s -- -b $(shell go env GOPATH)/bin $(LINTER_VERSION)
	# See https://github.com/golangci/golangci-lint/issues/843 for --exclude-use-default=false use
	$(shell go env GOPATH)/bin/golangci-lint --fix=true --exclude-use-default=false -v run


## For utility only, this will tidy, drop files out of git, remove, and then re-add
revendor:
	@echo "Re Vendoring"
	go mod tidy
	go mod vendor
	git add vendor

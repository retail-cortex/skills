# Root Makefile for skill-builder delegating to Bazel

.PHONY: all build build-cli build-binaries test test-go clean

all: test build

build: build-cli

build-cli:
	bazel build //:skm

build-binaries:
	bazel build //:skm-binaries

test: test-go

test-go:
	bazel test //cli/... //clients/go/...

clean:
	bazel clean

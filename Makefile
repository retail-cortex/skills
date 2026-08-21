# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all build test lint fmt clean server cli docs test-e2e

GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")

LDFLAGS := -X github.com/retail-cortex/castor/internal/commands.Version=$(VERSION) \
           -X github.com/retail-cortex/castor/internal/commands.GitCommit=$(GIT_COMMIT) \
           -X github.com/retail-cortex/castor/internal/commands.BuildDate=$(BUILD_DATE)

all: build test

build:
	bazel build //...

test:
	bazel test //...

test-e2e:
	bazel test //:test-e2e

lint:
	golangci-lint run ./cmd/... ./pkg/... ./internal/... ./clients/go/...

fmt:
	gofmt -w -s .
	ruff format .
	ruff check --fix .

server:
	bazel run //cmd/castor_server

cli:
	bazel run //cmd/cstr -- $(ARGS)

docs:
	bazel run //docs:serve

clean:
	bazel clean

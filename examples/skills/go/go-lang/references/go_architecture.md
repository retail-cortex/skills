# Go Enterprise Service Architecture & TDD Guide

## 1. Table-Driven TDD with testify/assert

Write structured, parameterized test suites in Go using `testify/assert`:

```go
package server_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		inputPort   string
		expectedErr bool
	}{
		{name: "Valid Port", inputPort: "8080", expectedErr: false},
		{name: "Invalid Port", inputPort: "", expectedErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.inputPort)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

## 2. Automated Mock Generation with Mockery

Isolate database and gRPC calls by generating interface mocks:

```bash
# Generate mock for Database interface
mockery --name=Database --dir=internal/database --output=internal/database/mocks
```

## 3. GitHub Actions CI/CD (`.github/workflows/go.yml`)

```yaml
name: Go CI/CD

on: [push, pull_request]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26.x'
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
      - name: Run TDD Suite with 85% Coverage
        run: |
          go test -v -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

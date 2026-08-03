# Go Enterprise Scaffolding & Configuration Architecture

## Directory Scaffolding Script

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Module Initialization
go mod init github.com/enterprise/service

# 2. Directory Layout
mkdir -p cmd/server internal/server internal/database pkg/utils api/proto configs/terraform .github/workflows

# 3. Create baseline files
touch cmd/server/main.go
touch internal/server/server.go
touch configs/.env.toml configs/.env.local.toml
touch configs/terraform/main.tf configs/terraform/variables.tf
touch BUILD.bazel
```

## modenv Integration (`cmd/server/main.go`)

Load configuration relative to `configs/` using `modenv`:

```go
package main

import (
	"log"
	"os"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type Config struct {
	Server struct {
		RestPort string `toml:"rest_port"`
		GrpcPort string `toml:"grpc_port"`
	} `toml:"server"`
	Database struct {
		Host     string `toml:"host"`
		Password string `toml:"password"`
	} `toml:"database"`
}

func loadAppConfig() *Config {
	os.Setenv("MODENV_PREFIX", "configs")
	var cfg Config
	clone, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Fatal config error: %v", err)
	}
	return clone.(*Config)
}
```

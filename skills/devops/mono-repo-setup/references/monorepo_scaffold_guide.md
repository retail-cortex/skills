# Enterprise Monorepo Scaffolding & Configuration Architecture

## Directory Initialization Commands

Run the following bash script to initialize a new enterprise monorepo workspace:

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Directory Tree
mkdir -p apps libs api/proto configs/terraform build/ci .github/workflows docs

# 2. Bazel Base Files (Java 25 LTS, Python 3.13, Go 1.26+, Node 22+)
cat << 'EOF' > .bazelrc
build --cxxopt=-std=c++17
build --enable_bzlmod
test --test_output=errors
coverage --combined_report=lcov
EOF

echo "9.0.0" > .bazelversion
```

## Configuration Invariants (`configs/`)

1. **`configs/terraform/`**: Contains GCS remote backend configuration and GCP module definitions for GKE, Cloud Run, BigQuery, and AlloyDB.
2. **`configs/.env.toml`**: Base configuration parsed by `modenv` or language TOML readers.
3. **`configs/.env.local.toml`**: Local developer secrets with XOR encryption (`xor:...`), strictly excluded in `.gitignore`.

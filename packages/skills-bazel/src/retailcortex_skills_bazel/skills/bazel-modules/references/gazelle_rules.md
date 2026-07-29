# Gazelle Build File Generation Rules

## Overview

Gazelle is a build file generator for Go, Python, and Protocol Buffers in Bazel. It automatically parses import statements and writes appropriate `BUILD.bazel` rules.

## Root BUILD.bazel Directives

Define Gazelle binaries and plugins at the repository root:

```starlark
load("@bazel_gazelle//:def.bzl", "gazelle", "gazelle_binary")

# Root Gazelle rule for Go
# gazelle:prefix github.com/enterprise/monorepo
# gazelle:proto disable_global
gazelle(name = "gazelle")

# Gazelle rule for Python
# gazelle:python_root .
# gazelle:python_manifest //:gazelle_python.yaml
gazelle(
    name = "gazelle_python",
    gazelle = ":gazelle_binary",
)

gazelle_binary(
    name = "gazelle_binary",
    languages = [
        "@bazel_gazelle//language/go",
        "@bazel_gazelle//language/proto",
        "@rules_python_gazelle_plugin//python",
    ],
)
```

## Running Gazelle

Run Gazelle commands via Bazel:
```bash
# Update Go targets
bazel run //:gazelle

# Update Python targets
bazel run //:gazelle_python

# Fix and format Go dependency imports
bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies -prune
```

# Bzlmod Enterprise Architecture & BCR Integration

## Overview

Bzlmod is the external dependency subsystem for Bazel. It replaces the legacy `WORKSPACE` model with a transitive, modular dependency graph defined in `MODULE.bazel`.

## Core Directives & Latest Toolchains

### 1. Direct Module Dependencies (`bazel_dep`)
```starlark
bazel_dep(name = "bazel_skylib", version = "1.8.2")
bazel_dep(name = "rules_go", version = "0.61.1", repo_name = "io_bazel_rules_go")
bazel_dep(name = "gazelle", version = "0.44.0", repo_name = "bazel_gazelle")
bazel_dep(name = "rules_python", version = "1.7.0")
bazel_dep(name = "rules_java", version = "9.1.0")
bazel_dep(name = "rules_proto_grpc", version = "5.3.1")
bazel_dep(name = "aspect_rules_js", version = "2.3.8")
bazel_dep(name = "aspect_rules_ts", version = "3.6.0")
```

### 2. Hermetic Toolchains

#### Java 25 (LTS) Toolchain
```starlark
# Java 25 LTS hermetic toolchain setup
rules_java = use_extension("@rules_java//java:extensions.bzl", "rules_java")
```

#### Go 1.26+ SDK Extension
```starlark
go_sdk = use_extension("@rules_go//go:extensions.bzl", "go_sdk")
go_sdk.download(version = "1.26.3")

go_deps = use_extension("@bazel_gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
use_repo(go_deps, "com_github_gin_gonic_gin", "io_gorm_gorm")
```

#### Python 3.13+ Hermetic Toolchain
```starlark
python = use_extension("@rules_python//python/extensions:python.bzl", "python")
python.toolchain(
    configure_coverage_tool = True,
    is_default = True,
    python_version = "3.13",
)
use_repo(python, "python_3_13")
```

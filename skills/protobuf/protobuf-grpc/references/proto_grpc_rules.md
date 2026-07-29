# Bazel rules_proto_grpc Multi-Language Compilation

## BUILD.bazel Target Definitions

Compile `.proto` files into Go, Python, and TypeScript stubs simultaneously:

```starlark
load("@rules_proto//proto:defs.bzl", "proto_library")
load("@rules_proto_grpc_go//:defs.bzl", "go_grpc_library")
load("@rules_proto_grpc_python//:defs.bzl", "python_grpc_library")

proto_library(
    name = "customer_proto",
    srcs = ["customer.proto"],
    visibility = ["//visibility:public"],
)

go_grpc_library(
    name = "customer_go_grpc",
    importpath = "github.com/enterprise/service/api/customer",
    protos = [":customer_proto"],
    visibility = ["//visibility:public"],
)

python_grpc_library(
    name = "customer_python_grpc",
    protos = [":customer_proto"],
    visibility = ["//visibility:public"],
)
```

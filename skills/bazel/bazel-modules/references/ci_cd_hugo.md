# Bazel CI/CD & Hugo GitHub Pages Documentation

## GitHub Actions Workflow (`.github/workflows/bazel.yml`)

```yaml
name: Bazel CI/CD

on:
  push:
    branches: [main]
    tags: ['v*.*.*']
  pull_request:
    branches: [main]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Bazel
        uses: bazel-contrib/setup-bazel@v0.8.5
        with:
          bazelisk-version: '1.20.0'

      - name: Test All Targets (TDD)
        run: bazel test //... --test_output=errors

      - name: Collect Code Coverage
        run: bazel coverage //... --combined_report=lcov

      - name: Build Hugo Documentation
        run: bazel build //docs:hugo_site

      - name: Deploy to GitHub Pages
        if: startsWith(github.ref, 'refs/tags/v')
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./bazel-bin/docs/hugo_site
```

## Hugo Rules in MODULE.bazel

```starlark
bazel_dep(name = "rules_hugo", version = "0.3.0")
hugo_deps = use_extension("@rules_hugo//hugo:extensions.bzl", "hugo_deps")
hugo_deps.http_archive(
    name = "hugo_book",
    strip_prefix = "hugo-book-11.0.0",
    url = "https://github.com/alex-shpak/hugo-book/archive/v11.0.0.tar.gz",
)
use_repo(hugo_deps, "hugo", "hugo_book")
```

---
title: "Skill Dependencies"
weight: 40
---

# Skill Dependencies & Installation

The dependencies workflow resolves, downloads, verifies, and compiles skills for agent consumption in local workspaces.

---

## Adding Dependencies (`skm add`)

Fetch and install skills into `.skills/` (or `-d <path>`), generating a `.manifest.lock` with SHA-256 digests:

```bash
# Add registered enterprise skill from central server
skm add skm://skills/sk-9b1deb4d

# Add from GitHub repository
skm add github://retail-cortex/skills@main/packages/skills-python

# Add from Go module cache
skm add mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go

# Add from Java Maven artifact
skm add maven://com.retailcortex.skills:skills-java:1.0.0

# Add from local filesystem directory
skm add file:///path/to/my-skill -d ./skills --force
```

---

## Polyglot URI Resolution

`skm add` supports universal polyglot URI resolution:

| Scheme | Example | Description |
| :--- | :--- | :--- |
| **`skm://`** | `skm://skills/sk-9b1deb4d` | Resolves registered enterprise skills from central `skills-service`. |
| **`github://`** | `github://owner/repo[@ref][/path]` | Fetches git trees via `git clone` or GitHub API zipballs. |
| **`mod://`** | `mod://module_path[@version][/path]` | Resolves via `$GOPATH/pkg/mod` or `go mod download`. |
| **`maven://`** | `maven://groupId:artifactId:version` | Resolves from `~/.m2/repository` or `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves workspace packages within local Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves directly from local filesystem paths. |

---

## Integrity Verification (`skm verify`)

Cryptographically audit installed skills against `.manifest.lock` to detect tampering or drift:

```bash
skm verify -d ./skills
skm verify -d ./skills --json  # Structured JSON report
```

---

## Manifest Compilation (`skm compile`)

Compile skills into a pre-compiled `skills_manifest.json` for zero-I/O cold starts in serverless or container runtimes:

```bash
skm compile -d ./skills -o ./skills_manifest.json
```

---

## Discovery & Search (`skm list` & `skm search`)

List or search installed skills:

```bash
skm list -d .skills
skm search python -d .skills
```

# Retail Cortex DevOps & Cloud Skills (`retailcortex-skills-devops`)

[![PyPI Version](https://img.shields.io/pypi/v/retailcortex-skills-devops.svg)](https://pypi.org/project/retailcortex-skills-devops/)
[![Python Version](https://img.shields.io/badge/python-3.13%2B-blue.svg)](https://www.python.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Retail Cortex enterprise AI agent skills for Docker, Terraform GCP, NX monorepos, and OpenTelemetry.

## Bundled Skills

- `docker-containers`: Multi-stage Docker builds, distroless images, and container hardening.
- `terraform-gcp`: Infrastructure as Code with Terraform, GCS remote state, GKE & Cloud Run.
- `nx-monorepo`: NX monorepo management and affected build caching.
- `mono-repo-setup`: Monorepo structure best practices.
- `opentelemetry-google`: Google Cloud OpenTelemetry instrumentation and trace flushing.

## Installation

```bash
pip install retailcortex-skills-devops
```

Or using `uv`:

```bash
uv add retailcortex-skills-devops
```

## License

Apache License 2.0. See LICENSE for details.

# Castor: Google Colab Interactive Tutorial

[![Open In Colab](https://colab.research.google.com/assets/colab-badge.svg)](https://colab.research.google.com/github/retail-cortex/skills/blob/main/examples/colab/skills_service_adk_tutorial.ipynb)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This directory contains the customer-facing **Google Colab interactive tutorial** for the Castor Enterprise AI Agent Skills platform (`Castor Registry`, `Castor CLI (cstr)`) and Google Agent Development Kit (ADK).

---

## What This Colab Demonstrates

1. **Go Daemon & CLI Execution in Colab**:
   - Compiling the Go `castor-server` daemon and `cstr` CLI from source or downloading pre-built Linux `amd64` binaries from GitHub Releases (`curl -sSL https://github.com/retail-cortex/skills/releases/latest/download/...`).
   - Running `castor-server` as a background process listening on `http://127.0.0.1:8000` with health check verification.
2. **Enterprise Application Registration**:
   - Registering a client application with domain scoping (`urn:castor:app:retailcortex.com:checkout-agent`).
   - Domain verification status (`VERIFIED_SSO`), freemail protection against public email spoofing, and API key provisioning (`cstr_live_...`).
3. **Multi-Modal Skill Registration (Local & Remote GitHub)**:
   - Ingesting local file skills with progressive disclosure (`SKILL.md`, `references/`, `examples/`).
   - Ingesting remote GitHub repositories (`github://retail-cortex/skills@main/skills/canvas-image`).
   - Computing poly-column multi-modal vector embeddings (768d / 1408d / 3072d) for `pgvector` HNSW index storage.
4. **Cryptographic Manifest Locking (`.manifest.lock`)**:
   - Generating immutable SHA-256 digests for all installed skills.
   - Performing real-time tampering detection against simulated prompt injection attacks.
5. **JIT Dynamic Pre-Call Skill Retrieval ($k \le 3$)**:
   - Querying the central vector index to select the top $\le 3$ relevant skills on user prompt arrival.
   - Eliminating context window saturation and tool confusion.
6. **Google ADK Autonomous Agent Execution**:
   - Instantiating an autonomous AI coding agent grounded in dynamically retrieved skill instructions.
   - Executing multi-turn generation with HTTP 429 exponential backoff and randomized jitter resilience.

---

## Running in Google Colab

Click the **[Open in Colab](https://colab.research.google.com/github/retail-cortex/skills/blob/main/examples/colab/skills_service_adk_tutorial.ipynb)** badge at the top of this document or upload [`skills_service_adk_tutorial.ipynb`](skills_service_adk_tutorial.ipynb) directly to [Google Colab](https://colab.research.google.com/).

### Prerequisites
* Standard Colab Linux CPU runtime (x86_64).
* Go 1.22+ toolchain (pre-installed or auto-installed via `apt-get install -y golang-go`).
* Python 3.11 or higher.
* Optional: Google Cloud Vertex AI project and service account credentials if running against a live Vertex AI multimodal embedding endpoint.

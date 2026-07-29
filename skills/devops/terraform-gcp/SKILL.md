---
name: terraform-gcp
description: Elite Infrastructure as Code SDLC for GCP with Terraform. Covers terraform test TDD, GitHub Actions CI/CD with tflint, HTTP 429 API rate backoff, and GCS remote state security (CWE-312).
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
---

# GCP Infrastructure as Code & Terraform SDLC Skill

This skill prescribes best practices for provisioning and managing **Google Cloud Platform (GCP)** enterprise infrastructure using **Terraform**, hermetic Bazel `rules_tf` toolchains, and complete SDLC rigor.

## Prerequisites & Pre-Flight Checklist

1. Terraform 1.9.0+ or OpenTofu installed.
2. GCS bucket provisioned for remote state backend (`backend "gcs"`).
3. GCP Service Account with least-privilege deployment roles.

## HTTP 429 Rate Limit & GCP Provider Backoff

- Configure Google Terraform Provider with exponential backoff retries to survive GCP API 429 quota exhaustion during bulk provisioning.

## Security Checkpoints & CWE Invariants

- **CWE-312 (Cleartext Storage of Sensitive Information)**: State files MUST be stored in a dedicated GCS bucket with versioning and object encryption enabled. Local state is prohibited.
- **CWE-250 (Execution with Unnecessary Privileges)**: Enforce least-privilege IAM service accounts for Cloud Run and GKE workloads.
- **CWE-798 (Hardcoded Credentials)**: Never commit plain-text credentials; use GCP Secret Manager references.

## 3-Phase Execution Protocol

### Phase 1: Test-Driven Infrastructure (TDD)
Validate infrastructure definitions using native `terraform test` blocks or `terratest` in Go before deployment.

### Phase 2: Automated Linting & CI/CD Verification
Run static analysis and validate formatting in CI:
```bash
terraform fmt -recursive -check
tflint --recursive
```

### Phase 3: Plan, Review & Tagged Apply
```bash
terraform plan -out=tfplan
terraform apply tfplan
```

## Progressive Disclosure References

- **GCP Terraform Guide**: Read [`references/gcp_terraform_guide.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/terraform-gcp/references/gcp_terraform_guide.md).
- **Reference Terraform Manifest**: View [`examples/main.tf`](file:///Users/rmcguinness/Projects/skill-builder/skills/terraform-gcp/examples/main.tf).

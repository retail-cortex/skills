# Google Cloud Terraform Architecture & TDD Guide

## 1. Test-Driven Infrastructure with `terraform test`

Write native `.tftest.hcl` test blocks to assert infrastructure variables, CIDR ranges, and IAM bindings:

```hcl
run "verify_cloud_run_service_name" {
  command = plan

  assert {
    condition     = google_cloud_run_v2_service.agent_service.name == "adk-fastapi-service"
    error_message = "Cloud Run service name must be adk-fastapi-service"
  }
}
```

## 2. GitHub Actions CI/CD (`.github/workflows/terraform.yml`)

```yaml
name: GCP Terraform CI/CD

on: [push, pull_request]

jobs:
  validate-and-plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.9.5
      - run: terraform fmt -check
      - run: tflint
      - run: terraform init
      - run: terraform test
      - run: terraform plan
```

#!/usr/bin/env bash
set -euo pipefail

# Enterprise Infrastructure Deployment Script (Terraform & Google Cloud)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="$(cd "${SCRIPT_DIR}/../terraform" && pwd)"

PROJECT_ID="${GCP_PROJECT_ID:-${GOOGLE_CLOUD_PROJECT:-}}"
REGION="${GCP_REGION:-us-central1}"
TF_STATE_BUCKET="${GCS_TF_STATE_BUCKET:-}"

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT environment variable must be set."
  exit 1
fi

echo "=========================================================================="
echo "Enterprise Skill Builder: GCP Infrastructure Deployment"
echo "Project ID : ${PROJECT_ID}"
echo "Region     : ${REGION}"
echo "Terraform  : ${TF_DIR}"
echo "=========================================================================="

cd "${TF_DIR}"

echo "[1/4] Checking Terraform Formatting..."
terraform fmt -check -recursive

echo "[2/4] Initializing Terraform Modules & Backend..."
if [[ -n "${TF_STATE_BUCKET}" ]]; then
  echo "Using GCS Remote State Bucket: gs://${TF_STATE_BUCKET}/skill-builder"
  terraform init \
    -backend-config="bucket=${TF_STATE_BUCKET}" \
    -backend-config="prefix=skill-builder/deployments" \
    -upgrade
else
  echo "Warning: GCS_TF_STATE_BUCKET not set. Initializing with local/default state."
  terraform init -upgrade
fi

echo "[3/4] Validating Terraform Configuration..."
terraform validate

echo "[4/4] Generating Execution Plan..."
terraform plan \
  -var="project_id=${PROJECT_ID}" \
  -var="region=${REGION}" \
  -out=tfplan

echo ""
read -r -p "Do you want to apply this Terraform plan to project '${PROJECT_ID}'? (yes/no): " CONFIRM
if [[ "${CONFIRM}" == "yes" ]]; then
  echo "Applying Terraform Execution Plan..."
  terraform apply tfplan
  rm -f tfplan
  echo "Infrastructure provisioning complete."
else
  echo "Deployment aborted by user."
  rm -f tfplan
  exit 0
fi

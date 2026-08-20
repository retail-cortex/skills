#!/usr/bin/env bash
set -euo pipefail

# Workload Identity IAM Binding Verification Script

PROJECT_ID="${GCP_PROJECT_ID:-${GOOGLE_CLOUD_PROJECT:-}}"
REGION="${GCP_REGION:-us-central1}"
ENV="${1:-dev}"
NAMESPACE="castor"
K8S_SA_NAME="castor-registry-sa"
GCP_SA_NAME="castor-registry-${ENV}-sa"
GCP_SA_EMAIL="${GCP_SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT must be set."
  exit 1
fi

echo "=========================================================================="
echo "Configuring GKE Workload Identity IAM Bindings"
echo "Project ID  : ${PROJECT_ID}"
echo "Environment : ${ENV}"
echo "GCP SA      : ${GCP_SA_EMAIL}"
echo "K8s SA      : ${NAMESPACE}/${K8S_SA_NAME}"
echo "=========================================================================="

# 1. Allow Kubernetes ServiceAccount to impersonate Google Service Account
gcloud iam service-accounts add-iam-policy-binding "${GCP_SA_EMAIL}" \
  --project="${PROJECT_ID}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${K8S_SA_NAME}]"

# 2. Annotate Kubernetes ServiceAccount
kubectl annotate serviceaccount "${K8S_SA_NAME}" \
  --namespace="${NAMESPACE}" \
  --overwrite \
  iam.gke.io/gcp-service-account="${GCP_SA_EMAIL}"

echo "Workload Identity successfully configured for ${GCP_SA_EMAIL} -> ${NAMESPACE}/${K8S_SA_NAME}."

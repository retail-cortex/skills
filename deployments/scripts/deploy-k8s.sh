#!/usr/bin/env bash
set -euo pipefail

# Enterprise Kubernetes Deployment Script for Skill Service on GKE (dev, qa, prod)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(cd "${SCRIPT_DIR}/../kubernetes" && pwd)"

ENV="${1:-}"
PROJECT_ID="${GCP_PROJECT_ID:-${GOOGLE_CLOUD_PROJECT:-}}"
REGION="${GCP_REGION:-us-central1}"
NAMESPACE="skill-builder"

if [[ -z "${ENV}" || ( "${ENV}" != "dev" && "${ENV}" != "qa" && "${ENV}" != "prod" ) ]]; then
  echo "Usage: $0 <dev|qa|prod>"
  exit 1
fi

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT environment variable must be set."
  exit 1
fi

CLUSTER_NAME="skill-builder-gke-${ENV}"
OVERLAY_DIR="${K8S_DIR}/overlays/${ENV}"

echo "=========================================================================="
echo "Skill Service GKE Deployment"
echo "Environment : ${ENV}"
echo "Cluster     : ${CLUSTER_NAME}"
echo "Project ID  : ${PROJECT_ID}"
echo "Region      : ${REGION}"
echo "Namespace   : ${NAMESPACE}"
echo "=========================================================================="

# 1. Fetch GKE Credentials
echo "[1/5] Fetching GKE cluster credentials..."
gcloud container clusters get-credentials "${CLUSTER_NAME}" \
  --region "${REGION}" \
  --project "${PROJECT_ID}"

# 2. Ensure Target Namespace Exists
echo "[2/5] Creating namespace '${NAMESPACE}' if not present..."
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# 3. Synchronize AlloyDB DSN Secret from Secret Manager
echo "[3/5] Syncing database connection string from Google Secret Manager..."
SECRET_ID="skill-builder-alloydb-${ENV}-dsn"
DB_DSN=$(gcloud secrets versions access latest --secret="${SECRET_ID}" --project="${PROJECT_ID}")

kubectl create secret generic skill-service-db-secret \
  --namespace="${NAMESPACE}" \
  --from-literal=database_url="${DB_DSN}" \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Deploy Kustomize Overlay
echo "[4/5] Applying Kustomize manifests for '${ENV}'..."
kubectl apply -k "${OVERLAY_DIR}"

# 5. Monitor Rollout Status
echo "[5/5] Monitoring rollout progress..."
kubectl rollout status deployment/skill-service \
  --namespace="${NAMESPACE}" \
  --timeout=180s

echo "Successfully deployed skill-service to ${CLUSTER_NAME} in namespace ${NAMESPACE}."

#!/usr/bin/env bash
set -euo pipefail

# Bootstrap AlloyDB AI Databases and Extensions

PROJECT_ID="${GCP_PROJECT_ID:-${GOOGLE_CLOUD_PROJECT:-}}"
REGION="${GCP_REGION:-us-east4}"
CLUSTER_ID="${ALLOYDB_CLUSTER_ID:-castor-alloydb}"
PRIMARY_IP="${ALLOYDB_PRIMARY_IP:-}"

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT must be set."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/init-alloydb-extensions.sql"

echo "=========================================================================="
echo "Initializing AlloyDB AI Extensions (vector, alloydb_scann, google_ml)"
echo "Project ID : ${PROJECT_ID}"
echo "Region     : ${REGION}"
echo "Cluster ID : ${CLUSTER_ID}"
echo "=========================================================================="

DATABASES=("castor_dev" "castor_qa" "castor_prod" "skills_dev" "skills_qa" "skills_prod")

for DB in "${DATABASES[@]}"; do
  echo "Applying extensions to database '${DB}'..."
  if [[ -n "${PRIMARY_IP}" ]]; then
    PGPASSWORD="${PGPASSWORD:-}" psql -h "${PRIMARY_IP}" -U postgres -d "${DB}" -f "${SQL_FILE}" || echo "Notice: Ensure database ${DB} exists or psql connection is active."
  else
    echo "Info: Run the following SQL against database '${DB}':"
    cat "${SQL_FILE}"
  fi
done

echo "Extension initialization instructions completed."

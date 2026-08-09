#!/usr/bin/env bash
set -euo pipefail

# Verify Hugo site artifact output directory exists and contains index.html
SITE_DIR=$(find . -name "index.html" -path "*/docs/site/*" | head -n 1)

if [ -z "${SITE_DIR}" ]; then
  echo "Error: index.html not found in Hugo site build outputs."
  exit 1
fi

if grep -q 'href="/"' "${SITE_DIR}"; then
  echo "Error: index.html has empty stylesheet references (href=\"/\"). Theme assets missing!"
  exit 1
fi

echo "Hugo site successfully verified at ${SITE_DIR}"
exit 0


#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# One-time setup before first deploy:
#
# 1. Login Docker ke Artifact Registry:
#    gcloud auth configure-docker asia-southeast2-docker.pkg.dev
#
#    # i made my own AR remote from console cz this script is garbage
# 2. Buat AR remote repository (proxy ghcr.io):
#    gcloud artifacts repositories create ghcr-remote \
#      --repository-format docker \
#      --location asia-southeast2 \
#      --mode remote-repository \
#      --remote-docker-repo https://ghcr.io
#
# 3. Login ghcr.io (untuk push):
#    echo $GITHUB_TOKEN | docker login ghcr.io -u redsylx --password-stdin
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -f "$SCRIPT_DIR/.env" ]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
fi

: "${GCS_BUCKET_NAME:?Must be set in .env}"
: "${DISCORD_WEBHOOK_URL:?Must be set in .env}"
: "${GCP_PROJECT:?Must be set in .env}"

GHCR_IMAGE="${GHCR_IMAGE:-ghcr.io/redsylx/tldr-discord-service}"
AR_REMOTE_IMAGE="${AR_REMOTE_IMAGE:-asia-southeast2-docker.pkg.dev/$GCP_PROJECT/ghcr-remote/redsylx/tldr-discord-service}"
VERSION="${VERSION:-v0.1.0}"
REGION="${REGION:-asia-southeast2}"
SERVICE_NAME="${SERVICE_NAME:-tldr-discord-service}"

RUNTIME_VARS=(GCS_BUCKET_NAME DISCORD_WEBHOOK_URL BATCH_LINE_COUNT DISCORD_DELAY_MS)

ENV_VARS=""
first=true
for var in "${RUNTIME_VARS[@]}"; do
  if [ -n "${!var:-}" ]; then
    if [ "$first" = true ]; then
      ENV_VARS="${var}=${!var}"
      first=false
    else
      ENV_VARS="${ENV_VARS},${var}=${!var}"
    fi
  fi
done

echo "==> Deploying to Cloud Run ($REGION)"
gcloud run deploy "$SERVICE_NAME" \
  --project "$GCP_PROJECT" \
  --image "$AR_REMOTE_IMAGE:$VERSION" \
  --region "$REGION" \
  --max-instances 1 \
  --concurrency 1 \
  --memory 128Mi \
  --cpu 0.1 \
  --timeout 5m \
  --set-env-vars "$ENV_VARS"

echo "==> Done!"
#!/usr/bin/env bash
set -euo pipefail

# Rebuild and restart the single-container FastClaw deployment.
#
# Default behavior keeps the persisted data dir (~/.fastclaw) intact:
#   ./scripts/redeploy-docker.sh
#
# Optional: pull the current branch before rebuilding:
#   ./scripts/redeploy-docker.sh --pull
#
# Optional env overrides:
#   FASTCLAW_CONTAINER_NAME=fastclaw
#   FASTCLAW_IMAGE=fastclaw
#   FASTCLAW_PORT=18953
#   FASTCLAW_HOME_DIR=$HOME/.fastclaw

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

PULL_FIRST=0
if [[ "${1:-}" == "--pull" ]]; then
  PULL_FIRST=1
fi

CONTAINER_NAME="${FASTCLAW_CONTAINER_NAME:-fastclaw}"
IMAGE_NAME="${FASTCLAW_IMAGE:-fastclaw}"
PORT="${FASTCLAW_PORT:-18953}"
HOME_DIR="${FASTCLAW_HOME_DIR:-$HOME/.fastclaw}"

if [[ "$PULL_FIRST" == "1" ]]; then
  echo "==> Pulling latest code"
  git pull
fi

echo "==> Ensuring data directory exists: $HOME_DIR"
mkdir -p "$HOME_DIR"

echo "==> Removing old container if present: $CONTAINER_NAME"
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

echo "==> Building image: $IMAGE_NAME"
docker build -t "$IMAGE_NAME" .

echo "==> Starting container: $CONTAINER_NAME"
docker run -d \
  --name "$CONTAINER_NAME" \
  -p "$PORT:18953" \
  -e FASTCLAW_BIND=all \
  -e FASTCLAW_PORT=18953 \
  -v "$HOME_DIR:/data/.fastclaw" \
  "$IMAGE_NAME"

echo "==> Container status"
docker ps --filter "name=$CONTAINER_NAME"

echo "==> Recent logs"
docker logs --tail 50 "$CONTAINER_NAME"

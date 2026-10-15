#!/bin/bash
# builds openpanel/caddy-coraza for amd64 and arm64 into one manifest, and pushes it with --push
# usage: bash build.sh [tag] [--push]   e.g. bash build.sh 2.11.7 --push  (also tags :latest)
set -euo pipefail
cd "$(dirname "$0")"

IMAGE="docker.io/openpanel/caddy-coraza"
TAG="${1:-test}"
PUSH=false
[[ "${2:-}" == "--push" ]] && PUSH=true

podman manifest rm "$IMAGE:$TAG" >/dev/null 2>&1 || true
podman build --platform linux/amd64,linux/arm64 --manifest "$IMAGE:$TAG" .

# the manifest must hold both arches, and the native one must have the rate limit and waf handlers (arm64 can't run here without qemu)
podman manifest inspect "$IMAGE:$TAG" | grep -q '"architecture": "arm64"' || { echo "arm64 missing from manifest"; exit 1; }
podman manifest inspect "$IMAGE:$TAG" | grep -q '"architecture": "amd64"' || { echo "amd64 missing from manifest"; exit 1; }
modules=$(podman run --rm --platform linux/amd64 "$IMAGE:$TAG" caddy list-modules --skip-standard)
grep -q "http.handlers.rate_limit" <<< "$modules" || { echo "rate_limit module missing"; exit 1; }
grep -q "http.handlers.waf" <<< "$modules" || { echo "waf module missing"; exit 1; }
echo "$IMAGE:$TAG built for amd64 and arm64 with rate_limit and waf."

if $PUSH; then
    podman manifest push --all "$IMAGE:$TAG" "docker://$IMAGE:$TAG"
    podman manifest push --all "$IMAGE:$TAG" "docker://$IMAGE:latest"
fi

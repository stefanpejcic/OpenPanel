#!/bin/bash

# openpanel/caddy-coraza was rebuilt on caddy 2.11.9 with the newest go and patched alpine packages (brotli, curl, musl, zlib),
# servers on the WAF image pull the new :latest and recreate caddy with it.
# the new image is pulled and has to validate the current Caddyfile before the running container is touched,
# so a failed pull or a config the new caddy doesn't accept leaves the old caddy running.

IMAGE="docker.io/openpanel/caddy-coraza:latest"

if ! podman ps --format '{{.Names}}' | grep -qx caddy; then
    echo "caddy container is not running, skipping the caddy-coraza image update"
    exit 0
fi

current=$(podman inspect --format '{{.ImageName}}' caddy 2>/dev/null)
if [[ "$current" != *openpanel/caddy-coraza* ]]; then
    echo "caddy runs $current, not the WAF image, skipping the caddy-coraza image update"
    exit 0
fi

old_id=$(podman inspect --format '{{.Image}}' caddy 2>/dev/null)

echo "Pulling $IMAGE..."
if ! timeout 300 podman pull -q "$IMAGE" >/dev/null; then
    echo "Failed to pull $IMAGE, caddy keeps running the old image. Update it by hand: podman pull $IMAGE && cd /root && podman-compose down caddy && podman-compose up -d caddy"
    exit 0
fi

new_id=$(podman image inspect --format '{{.Id}}' "$IMAGE" 2>/dev/null)
if [[ -n "$old_id" && "$old_id" == "$new_id" ]]; then
    echo "caddy already runs the newest $IMAGE"
    exit 0
fi

# same mounts as the caddy service in /root/docker-compose.yml, so the waf rules and imports resolve like they do in the real container
if ! timeout 120 podman run --rm --network none \
        -v /etc/openpanel/caddy/Caddyfile:/etc/caddy/Caddyfile:ro \
        -v /etc/openpanel/caddy/:/etc/openpanel/caddy:ro \
        -v /var/log/caddy/:/var/log/caddy/ \
        "$IMAGE" caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/tmp/caddy-validate-2.0.25.log 2>&1; then
    echo "The new $IMAGE does not accept /etc/openpanel/caddy/Caddyfile, caddy keeps running the old image. Details: /tmp/caddy-validate-2.0.25.log"
    exit 0
fi
rm -f /tmp/caddy-validate-2.0.25.log

echo "Recreating caddy with the new image..."
cd /root || exit 0
podman-compose down caddy >/dev/null 2>&1
if ! podman-compose up -d caddy >/dev/null 2>&1; then
    echo "Failed to start caddy with the new image, check: cd /root && podman-compose up -d caddy"
    exit 0
fi

# caddy needs a few seconds to load the waf rules
for _ in $(seq 1 15); do
    if [[ "$(podman inspect --format '{{.State.Running}}' caddy 2>/dev/null)" == "true" ]] && podman exec caddy caddy version >/dev/null 2>&1; then
        echo "caddy now runs $(podman exec caddy caddy version | awk '{print $1}') from $IMAGE"
        # drop the old image so it doesn't keep its CVEs around on disk, ignored when something else still uses it
        [[ -n "$old_id" ]] && podman rmi "$old_id" >/dev/null 2>&1
        exit 0
    fi
    sleep 2
done
echo "caddy did not come up after the image update, check: podman logs caddy"

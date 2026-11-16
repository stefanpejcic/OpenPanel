#!/bin/bash

# AI & bot control: existing servers need the botcontrol snippets and their import, or the first block/rate-limit a user picks fails the caddy reload
# the file only defines snippets, so it's valid on caddy images without the rate_limit module too; domain configs aren't touched, the panel adds the line when a policy is set
BC_CONF=/etc/openpanel/caddy/botcontrol.conf
CADDYFILE=/etc/openpanel/caddy/Caddyfile
if [ -f "$CADDYFILE" ]; then
    if [ ! -f "$BC_CONF" ]; then
        echo "Adding $BC_CONF..."
        cat > "$BC_CONF" <<'BOTCONTROL'
# AI & bot crawler control snippets for per-domain Caddy configs
# a domain config holds one of: "# botcontrol: allow", "import botcontrol_block" or "import botcontrol_ratelimit <domain>"
# keep the list in sync with knownBots in the panel and OpenAdmin; Google-Extended and Applebot-Extended are left out, they are robots.txt tokens and never sent as a User-Agent

(ai_crawler_matcher) {
    @ai_crawlers {
        header_regexp ai_ua User-Agent (?i)(GPTBot|ChatGPT-User|OAI-SearchBot|ClaudeBot|Claude-Web|anthropic-ai|Bytespider|CCBot|PerplexityBot|Perplexity-User|Amazonbot|Meta-ExternalAgent|Meta-ExternalFetcher|FacebookBot|Diffbot|Omgilibot|cohere-ai|YouBot|DuckAssistBot|Timpibot|Scrapy)
    }
}

(botcontrol_block) {
    import ai_crawler_matcher
    respond @ai_crawlers "Forbidden" 403
}

# args[0] = domain
# keyed on the User-Agent, not the IP: crawlers spread over hundreds of addresses, so a per-IP limit never kicks in
# 10 requests per minute per crawler per domain, over that the visitor gets 429
(botcontrol_ratelimit) {
    import ai_crawler_matcher
    handle @ai_crawlers {
        rate_limit {
            zone botcontrol_{args[0]} {
                key {http.request.header.User-Agent}
                events 10
                window 1m
            }
        }
    }
}
BOTCONTROL
    fi
    if ! grep -q "^import $BC_CONF" "$CADDYFILE"; then
        echo "Adding the bot control snippets import to $CADDYFILE..."
        # written in place, caddy bind-mounts this single file and would keep the old inode after sed -i
        sed "s|^import /etc/openpanel/caddy/domains/\*|# import AI \& bot crawler control snippets\nimport $BC_CONF\n\n&|" "$CADDYFILE" > /tmp/op_2025_caddyfile \
            && grep -q "^import $BC_CONF" /tmp/op_2025_caddyfile \
            && cat /tmp/op_2025_caddyfile > "$CADDYFILE"
        rm -f /tmp/op_2025_caddyfile
        grep -q "^import $BC_CONF" "$CADDYFILE" || echo "Could not add the import, add 'import $BC_CONF' above 'import /etc/openpanel/caddy/domains/*' in $CADDYFILE"
    fi
fi

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

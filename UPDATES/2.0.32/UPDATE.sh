#!/bin/bash

# AI assistant: reasoning effort setting, empty leaves it out of requests so the model's default applies
OPENPANEL_CONFIG=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$OPENPANEL_CONFIG" ] && grep -q "^\[AI\]" "$OPENPANEL_CONFIG" && ! grep -q "^ai_reasoning_effort=" "$OPENPANEL_CONFIG"; then
    echo "Adding ai_reasoning_effort to the [AI] section of $OPENPANEL_CONFIG..."
    sed -i "/^\[AI\]/a ai_reasoning_effort=" "$OPENPANEL_CONFIG"
fi

# Golub webmail: added next to Roundcube in the mail compose, Roundcube stays the default until the admin switches with: opencli email-webmail client golub
MAIL_DIR=/usr/local/mail/openmail
MAIL_COMPOSE=$MAIL_DIR/compose.yml
MAIL_ENV=$MAIL_DIR/mailserver.env
if [ -f "$MAIL_COMPOSE" ] && ! grep -q '^  golub:' "$MAIL_COMPOSE"; then
    echo "Adding the Golub webmail service to $MAIL_COMPOSE..."
    cp -a "$MAIL_COMPOSE" "$MAIL_COMPOSE.bak-2.0.32"
    golub_block=$(mktemp)
    cat > "$golub_block" <<'BLOCK'
  # alternative webmail, only one of roundcube and golub runs since both use port 8080; switch with: opencli email-webmail client golub
  golub:
    image: ghcr.io/stefanpejcic/golub:0.3.1
    container_name: openadmin_golub
    env_file: golub.env
    restart: always
    depends_on:
      - mailserver
    environment:
      # opencli rewrites these to the hostname with implicit TLS once the mailserver has a certificate
      - GOLUB_IMAP_HOST=mailserver
      - GOLUB_IMAP_TLS=none
      - GOLUB_SMTP_HOST=mailserver
      - GOLUB_SMTP_TLS=none
      - GOLUB_TRUST_PROXY=127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
    ports:
      - "8080:8080"
    networks:
      - network

BLOCK
    # with a hostname roundcube connects over TLS to it and has it pinned to the public IP, golub does the same
    mail_host=$(grep '^OVERRIDE_HOSTNAME=' "$MAIL_ENV" 2>/dev/null | cut -d= -f2-)
    if [ -n "$mail_host" ]; then
        sed -i -e "s|GOLUB_IMAP_HOST=mailserver|GOLUB_IMAP_HOST=$mail_host|" -e "s|GOLUB_SMTP_HOST=mailserver|GOLUB_SMTP_HOST=$mail_host|" \
               -e "s|GOLUB_IMAP_TLS=none|GOLUB_IMAP_TLS=implicit|" -e "s|GOLUB_SMTP_TLS=none|GOLUB_SMTP_TLS=implicit|" "$golub_block"
        pin=$(grep -A1 'extra_hosts: # opencli domain' "$MAIL_COMPOSE" | sed -n 2p)
        if [ -n "$pin" ]; then
            sed -i "/container_name: openadmin_golub/a\\    extra_hosts: # opencli domain\\n$pin" "$golub_block"
        fi
    fi
    # goes right before postfwd, or before the top-level networks: when there is no postfwd
    if grep -q '^  postfwd:' "$MAIL_COMPOSE"; then
        anchor='^  postfwd:'
    else
        anchor='^networks:'
    fi
    awk -v anchor="$anchor" -v block="$golub_block" '
        !done && $0 ~ anchor { while ((getline l < block) > 0) print l; done = 1 }
        { print }
    ' "$MAIL_COMPOSE" > "$MAIL_COMPOSE.tmp" && mv "$MAIL_COMPOSE.tmp" "$MAIL_COMPOSE"
    rm -f "$golub_block"
fi
# servers that got the golub block from an earlier 2.0.32 beta move up to 0.3.1; only the golub container is recreated, never the mailserver
if [ -f "$MAIL_COMPOSE" ] && grep -qE 'ghcr.io/stefanpejcic/golub:0\.(1|2|3|3\.0)$' "$MAIL_COMPOSE"; then
    echo "Updating Golub webmail to 0.3.1 in $MAIL_COMPOSE..."
    sed -i -E 's#ghcr.io/stefanpejcic/golub:0\.(1|2|3|3\.0)$#ghcr.io/stefanpejcic/golub:0.3.1#' "$MAIL_COMPOSE"
    if podman ps --format '{{.Names}}' | grep -qx openadmin_golub; then
        # pull first so golub is down only for the restart itself
        podman pull -q ghcr.io/stefanpejcic/golub:0.3.1 >/dev/null
        podman stop -t 10 openadmin_golub >/dev/null
        podman rm -f openadmin_golub >/dev/null
        (cd "$MAIL_DIR" && podman-compose up -d --no-deps golub >/dev/null 2>&1)
    fi
fi
if [ -f "$MAIL_ENV" ] && ! grep -q '^WEBMAIL_CLIENT=' "$MAIL_ENV"; then
    printf '\n# webmail client for the users: roundcube or golub, change it with: opencli email-webmail client golub\nWEBMAIL_CLIENT=roundcube\n' >> "$MAIL_ENV"
fi

# the webmail feature isn't only Roundcube anymore
FEATURES_JSON="/etc/openpanel/openadmin/config/features.json"
if [ -f "$FEATURES_JSON" ] && grep -q '"Webmail (Roundcube)"' "$FEATURES_JSON"; then
    echo "Renaming the webmail feature in $FEATURES_JSON..."
    sed -i -e 's/"title": "Webmail (Roundcube)"/"title": "Webmail"/' \
        -e 's/auto-login to webmail (Roundcube) for email accounts. The Roundcube service must be configured separately./auto-login to webmail (Roundcube or Golub) for email accounts. The webmail service must be configured separately./' "$FEATURES_JSON"
fi

# docker-proxy: deny rules so nothing on the socket-proxy network can create containers, change them or copy files in or out
# the rules are a haproxy template mounted over the image's own, docker-proxy is recreated for accounts where it's running
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
PROXY_TPL=/etc/openpanel/docker/socket-proxy/haproxy.cfg.template
PROXY_MOUNT="$PROXY_TPL:/usr/local/etc/haproxy/haproxy.cfg.template:ro"
tpl_tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$tpl_tmp" "$CONFIG_RAW/docker/socket-proxy/haproxy.cfg.template" </dev/null && grep -q "containers/(create|prune)" "$tpl_tmp"; then
    mkdir -p "$(dirname "$PROXY_TPL")"
    cat "$tpl_tmp" > "$PROXY_TPL"
    chmod 644 "$PROXY_TPL"
    echo "Updated $PROXY_TPL"
else
    echo "Could not download the docker-proxy rules, docker-proxy was left as it is. Run this update again later."
fi
rm -f "$tpl_tmp"

# adds the mount after docker-proxy's last volume with that line's indent, checked with the yaml parser, prints "added", "present" or "skip"
proxy_mount() {
    python3 - "$1" "$PROXY_MOUNT" <<'PY' 2>/dev/null || echo skip
import re, sys, yaml
path, mount = sys.argv[1], sys.argv[2]
text = open(path).read()
def proxy_vols(t):
    svc = ((yaml.safe_load(t) or {}).get("services") or {}).get("docker-proxy") or {}
    return [str(v) for v in (svc.get("volumes") or [])]
try:
    vols = proxy_vols(text)
except Exception:
    print("skip"); sys.exit()
if mount in vols:
    print("present"); sys.exit()
lines = text.split("\n")
ind = lambda l: len(l) - len(l.lstrip())
skip = lambda l: not l.strip() or l.lstrip().startswith("#")
# the docker-proxy block, its volumes key, then the last item of that list
top = [i for i, l in enumerate(lines) if re.match(r"^services:\s*$", l)]
if len(top) != 1:
    print("skip"); sys.exit()
first = next((j for j in range(top[0] + 1, len(lines)) if not skip(lines[j])), None)
if first is None:
    print("skip"); sys.exit()
# depends_on in cron and backup has a docker-proxy: line too, only the one at the service indent counts
start = [i for i, l in enumerate(lines) if re.match(r"^\s+['\"]?docker-proxy['\"]?:\s*$", l) and ind(l) == ind(lines[first])]
if len(start) != 1:
    print("skip"); sys.exit()
s0 = start[0]
end = next((j for j in range(s0 + 1, len(lines)) if not skip(lines[j]) and ind(lines[j]) <= ind(lines[s0])), len(lines))
vk = [j for j in range(s0 + 1, end) if re.match(r"^\s*volumes:\s*$", lines[j])]
if len(vk) != 1:
    print("skip"); sys.exit()
items = []
for j in range(vk[0] + 1, end):
    l = lines[j]
    if skip(l):
        continue
    if not l.lstrip().startswith("-") or ind(l) < ind(lines[vk[0]]):
        break
    items.append(j)
if not items:
    print("skip"); sys.exit()
i = items[-1]
indent = re.match(r"^(\s*-\s*)", lines[i]).group(1)
lines.insert(i + 1, indent + mount)
new = "\n".join(lines)
try:
    if mount not in proxy_vols(new):
        print("skip"); sys.exit()
except Exception:
    print("skip"); sys.exit()
open(path + ".op-2032", "w").write(new)
print("added")
PY
}

patch_proxy_compose() {
    local f="$1" res
    res=$(proxy_mount "$f")
    case "$res" in
        added)
            cp -p "$f" "$f.bak-2.0.32"
            # in place so the file keeps its owner and mode
            cat "$f.op-2032" > "$f"
            rm -f "$f.op-2032"
            echo "Added the docker-proxy rules mount to $f"
            return 0 ;;
        present) return 1 ;;
        *)
            rm -f "$f.op-2032"
            echo "Could not add the docker-proxy rules mount to $f (custom docker-proxy setup?), left as it is."
            return 1 ;;
    esac
}

# podman against the account's own socket: upod <seconds> <podman args>
upod() {
    local t="$1"; shift
    CONTAINER_HOST="unix://$sock" timeout "$t" podman --remote "$@" </dev/null
}
urunning() {
    [ "$(upod 20 inspect -f '{{.State.Running}}' "$1" 2>/dev/null)" = true ]
}
# cron and backup depend on docker-proxy and podman won't replace it while they exist, so they go with it and the ones that ran come back
proxy_up() {
    upod 60 rm -f --depend docker-proxy >/dev/null 2>&1
    (cd "$home_dir" && CONTAINER_HOST="unix://$sock" timeout 120 podman-compose up -d --no-deps docker-proxy) >/dev/null 2>&1 </dev/null
    for dep in $deps_were; do
        (cd "$home_dir" && CONTAINER_HOST="unix://$sock" timeout 120 podman-compose up -d --no-deps "$dep") >/dev/null 2>&1 </dev/null
    done
}

if [ -s "$PROXY_TPL" ]; then
    patch_proxy_compose /etc/openpanel/docker/compose/1.0/docker-compose.yml >/dev/null || true
    for f in /home/*/docker-compose.yml; do
        [ -f "$f" ] || continue
        home_dir=$(dirname "$f")
        ctx=$(basename "$home_dir")
        sock="/run/user/$(stat -c '%u' "$home_dir" 2>/dev/null)/podman/podman.sock"
        proxy_runs=0
        deps_were=""
        if [ -S "$sock" ] && urunning docker-proxy; then
            proxy_runs=1
            for dep in cron backup; do
                urunning "$dep" && deps_were="$deps_were $dep"
            done
            # an idle backup container has nothing in /tmp, a backup in progress keeps its archive and file lists there
            case " $deps_were " in *" backup "*)
                if [ -n "$(upod 20 exec backup sh -c 'find /tmp -type f | head -n 1' 2>/dev/null)" ] && ! grep -qF "$PROXY_MOUNT" "$f"; then
                    echo "A backup is running for $ctx, its docker-proxy was left as it is. Run this update again when the backup is done."
                    continue
                fi ;;
            esac
        fi
        patch_proxy_compose "$f" || continue
        if [ "$proxy_runs" = 0 ]; then
            rm -f "$f.bak-2.0.32"
            continue
        fi
        proxy_up
        if urunning docker-proxy && upod 20 inspect docker-proxy --format '{{range .Mounts}}{{.Destination}} {{end}}' 2>/dev/null | grep -q haproxy.cfg.template; then
            rm -f "$f.bak-2.0.32"
            echo "Recreated docker-proxy for $ctx with the rules"
        else
            cat "$f.bak-2.0.32" > "$f"
            proxy_up
            echo "docker-proxy for $ctx didn't start with the rules, its compose file was put back. Check: podman logs docker-proxy"
            rm -f "$f.bak-2.0.32"
        fi
    done
fi

# the Docker socket for users' own containers is off unless the admin turns it on (OpenAdmin > Settings > OpenPanel, or docker_socket=yes)
if [ -f "$OPENPANEL_CONFIG" ] && ! grep -q "^docker_socket=" "$OPENPANEL_CONFIG"; then
    if grep -q "^\[DOCKER\]" "$OPENPANEL_CONFIG"; then
        sed -i "/^\[DOCKER\]/a docker_socket=no" "$OPENPANEL_CONFIG"
    else
        printf '\n[DOCKER]\ndocker_socket=no\n' >> "$OPENPANEL_CONFIG"
    fi
    echo "Added docker_socket=no to $OPENPANEL_CONFIG"
fi

# containers set up before this keep what they have, list them so the admin can look
for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    python3 - "$f" <<'PY' 2>/dev/null
import os, sys, yaml
path = sys.argv[1]
ctx = os.path.basename(os.path.dirname(path))
try:
    services = (yaml.safe_load(open(path)) or {}).get("services") or {}
except Exception:
    sys.exit()
def is_socket(v):
    src = (v.get("source", "") if isinstance(v, dict) else str(v).split(":")[0]).strip()
    return src.startswith(("/run/user/", "/var/run/user/", "/run/podman")) or "podman.sock" in src or "docker.sock" in src
for name, svc in services.items():
    svc = svc or {}
    if name != "docker-proxy" and any(is_socket(v) for v in (svc.get("volumes") or [])):
        print(f"Warning: {ctx}'s container {name} mounts the podman socket, it can control and mount anything in that account. Remove it from {path} if it isn't needed.")
    nets = svc.get("networks") or []
    if name not in ("cron", "backup", "docker-proxy") and "socket-proxy" in (list(nets) if isinstance(nets, (list, dict)) else []):
        print(f"Warning: {ctx}'s container {name} is on the socket-proxy network, only cron and backup should be. Remove it from that network in {path}.")
PY
done

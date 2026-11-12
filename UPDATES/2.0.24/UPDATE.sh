#!/bin/bash

# repairs docker-compose.yml files where a service, volume or network block was added with a shallower indentation than the rest of the file
# (e.g. the litespeed service and litespeed_conf volume with 2 spaces in a file that uses 4, so everything after them was read as part of them)
# 2.0.21 did this only for litespeed and gave up on files that have multi-line commands; this re-indents just the broken blocks,
# then the result has to load as a valid compose file and pass podman-compose config before it's saved
fix_compose_indent() {
    local f="$1" dir tmp rc
    [ -f "$f" ] || return 0
    dir=$(dirname "$f")
    tmp=$(mktemp)
    rc=0
    # || keeps a non-zero exit from stopping the update when it runs under set -e
    python3 - "$f" "$tmp" <<'PYFIX' || rc=$?
import re, sys
from collections import Counter

import yaml

src, out = sys.argv[1], sys.argv[2]
text = open(src).read()
lines = text.split('\n')

VOLUME_KEYS = {'driver', 'driver_opts', 'external', 'labels', 'name'}
NETWORK_KEYS = {'driver', 'driver_opts', 'external', 'labels', 'name', 'internal', 'attachable', 'ipam', 'enable_ipv6'}
# probe key used to learn the indent unit: it sits 2 levels deep under every member of the section
PROBES = {'services': 'container_name', 'volumes': 'driver', 'networks': 'driver'}


def indent(l):
    return len(l) - len(l.lstrip(' '))


def section(ls, name):
    start = next((i for i, l in enumerate(ls) if re.match(r'^%s:[ \t]*$' % name, l)), None)
    if start is None:
        return None
    end = len(ls)
    for k in range(start + 1, len(ls)):
        if ls[k].strip() and not ls[k].startswith((' ', '\t', '#')):
            end = k
            break
    return start, end


def unit(ls, sec):
    rng = section(ls, sec)
    if not rng:
        return None
    a, b = rng
    cnt = Counter(indent(ls[k]) for k in range(a + 1, b) if re.match(r'^ +%s:' % PROBES[sec], ls[k]))
    if not cnt:
        return None
    top = cnt.most_common(1)[0][0]
    return top // 2 if top % 2 == 0 and top > 0 else None


def valid(ls, sec=None):
    try:
        d = yaml.safe_load('\n'.join(ls))
    except Exception:
        return False
    if not isinstance(d, dict) or not isinstance(d.get('services'), dict):
        return False
    if sec in (None, 'services'):
      for v in d['services'].values():
        if not isinstance(v, dict) or not ('image' in v or 'container_name' in v or 'build' in v):
            return False
        # a service swallowed by a shallower one shows up as a nested mapping with its own image/container_name
        if any(isinstance(x, dict) and ('image' in x or 'container_name' in x) for x in v.values()):
            return False
    for name, keys in (('volumes', VOLUME_KEYS), ('networks', NETWORK_KEYS)):
        if sec not in (None, name):
            continue
        for v in (d.get(name) or {}).values():
            if v is not None and not (isinstance(v, dict) and set(v) <= keys):
                return False
    return True


def reindent(ls, i, end, s, u):
    new = list(ls)
    for k in range(i, end):
        l = new[k]
        if l.strip():
            n = indent(l)
            new[k] = ' ' * (n // s * u + n % s) + l.lstrip(' ')
    return new


def fix_one(ls, sec):
    """re-indents the first member of sec that sits shallower than the rest, returns new lines or None"""
    rng = section(ls, sec)
    u = unit(ls, sec)
    if not rng or not u:
        return None
    a, b = rng
    i = next((k for k in range(a + 1, b)
              if re.match(r'^ +[\w.-]+:[ \t]*$', ls[k]) and 0 < indent(ls[k]) < u and u % indent(ls[k]) == 0), None)
    if i is None:
        return None
    s = indent(ls[i])
    # the block ends either where indentation drops back, or at a key on the real member level;
    # try every possible end and keep the first one that gives a valid compose file
    for j in range(i + 1, b + 1):
        l = ls[j] if j < b else ''
        if j < b and not l.strip():
            continue
        if j == b or indent(l) <= s or (indent(l) == u and re.match(r'^ +[\w.-]+:[ \t]*$', l)):
            end = j
            while end > i + 1 and not ls[end - 1].strip():
                end -= 1
            new = reindent(ls, i, end, s, u)
            if valid(new, sec):
                return new
            if j == b or indent(l) <= s:
                break
    return None


if any('\t' in l[:indent(l) + 1] for l in lines if l.strip()):
    sys.exit(1)

if valid(lines):
    sys.exit(10)  # nothing to fix

before = len((yaml.safe_load(text) or {}).get('services') or {}) if isinstance(yaml.safe_load(text), dict) else 0
changed = False
for _ in range(50):
    for sec in ('services', 'volumes', 'networks'):
        new = fix_one(lines, sec)
        if new:
            lines, changed = new, True
            break
    else:
        break

if not changed or not valid(lines):
    sys.exit(1)
if len(yaml.safe_load('\n'.join(lines))['services']) < before:
    sys.exit(1)

open(out, 'w').write('\n'.join(lines))
PYFIX
    if [ "$rc" -eq 0 ] && command -v podman-compose >/dev/null 2>&1; then
        local env_args=()
        [ -f "$dir/.env" ] && env_args=(--env-file "$dir/.env")
        # config only parses and interpolates the file, it doesn't touch containers
        if ! (cd "$dir" && timeout 120 podman-compose -f "$tmp" "${env_args[@]}" config >/dev/null 2>&1); then
            rc=1
        fi
    fi
    if [ "$rc" -eq 0 ]; then
        [ -f "$f.bak-2.0.23" ] || cp -a "$f" "$f.bak-2.0.23"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
        echo "Fixed indentation in $f"
    elif [ "$rc" -ne 10 ]; then
        echo "Could not fix the indentation in $f, left as it is."
    fi
    rm -f "$tmp"
}

fix_compose_indent /etc/openpanel/docker/compose/1.0/docker-compose.yml
for f in /home/*/docker-compose.yml; do
    fix_compose_indent "$f"
done

# imunifyav module is new in 2.0.24, list it under OpenAdmin > Settings > Modules so admins can turn it on
FEATURES_JSON="/etc/openpanel/openadmin/config/features.json"
if [ -f "$FEATURES_JSON" ] && ! jq -e '.[] | select(.name == "imunifyav")' "$FEATURES_JSON" >/dev/null 2>&1; then
    echo "Adding the ImunifyAV module to $FEATURES_JSON..."
    tmp=$(mktemp)
    if jq '. + [{
        "name": "imunifyav",
        "title": "ImunifyAV",
        "description": "If enabled, users get ImunifyAV'"'"'s own interface for their account and a summary of infected files and scans. Needs ImunifyAV installed from OpenAdmin > Security > ImunifyAV.",
        "link": "/imunifyav",
        "type": "community",
        "help_link": "https://openpanel.com/docs/admin/settings/modules/#imunifyav"
    }]' "$FEATURES_JSON" > "$tmp"; then
        # write in place so the file keeps its owner and permissions
        cat "$tmp" > "$FEATURES_JSON"
    else
        echo "Failed to add the ImunifyAV module to $FEATURES_JSON"
    fi
    rm -f "$tmp"
fi

# openpanel mounts imunify's rpc socket dir once the module is on, without the dir podman won't start the container
install -d -m 755 /run/defence360agent
echo "d /run/defence360agent 0755 root root -" > /etc/tmpfiles.d/openpanel-imunify.conf

# servers that already have ImunifyAV get the module turned on, same as a fresh opencli imunify install
CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
if command -v imunify-antivirus >/dev/null 2>&1 && [ -d /etc/sysconfig/imunify360 ] && [ -f "$CONFIG_FILE" ]; then
    if ! grep -q "imunify-gui-panel" /usr/local/opencli/imunify.sh 2>/dev/null; then
        echo "opencli is older than 2.0.24, run 'opencli update --cli' and then 'opencli imunify start' to set up the ImunifyAV page for users"
    else
        newly_enabled=""
        enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE" | cut -d'=' -f2- | tr -d '"')
        if ! echo ",$enabled_modules," | grep -q ',imunifyav,'; then
            echo "ImunifyAV is installed, enabling the imunifyav module..."
            newly_enabled=1
            if [ -n "$enabled_modules" ]; then
                sed -i "s|^enabled_modules=.*|enabled_modules=\"${enabled_modules},imunifyav\"|" "$CONFIG_FILE"
            else
                sed -i 's|^enabled_modules=.*|enabled_modules="imunifyav"|' "$CONFIG_FILE"
            fi
        fi
        had_mount=""
        grep -q "/run/defence360agent" /root/docker-compose.yml 2>/dev/null && had_mount=1
        # mounts the socket into openpanel (recreates it when the mount is new) and sets up the gui service for the user panel
        timeout 300 opencli imunify start || echo "opencli imunify start failed, run it by hand to set up the ImunifyAV page for users"
        # no recreate happened, so openpanel needs a restart to load the module
        if [ -n "$newly_enabled" ] && [ -n "$had_mount" ] && podman ps --format '{{.Names}}' | grep -qx openpanel; then
            podman restart -t 10 openpanel >/dev/null 2>&1
        fi
    fi
fi

# one-click docker apps: the panel reads them from github and falls back to this copy, refresh it so new apps like Weblate also show up when github can't be reached
APPS_YML="/etc/openpanel/docker/templates/applications.yml"
mkdir -p "$(dirname "$APPS_YML")"
tmp=$(mktemp)
if wget -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/templates/applications.yml" && grep -q "^applications:" "$tmp" && grep -q "id: weblate" "$tmp"; then
    echo "Updating $APPS_YML..."
    cat "$tmp" > "$APPS_YML"
    chmod 644 "$APPS_YML"
else
    echo "Failed to download the new $APPS_YML, update it by hand from the openpanel-configuration repo"
fi
rm -f "$tmp"

# hosts file for users' containers, new users get base_hosts_file pointing at it, existing ones only with: opencli docker-hosts --apply
if [ -f /usr/local/opencli/docker/hosts.sh ]; then
    opencli docker-hosts || echo "Failed to write /etc/openpanel/docker/container_hosts, run 'opencli docker-hosts' by hand"
else
    echo "opencli is older than 2.0.24, run 'opencli update --cli' and then 'opencli docker-hosts'"
fi

# servers installed before the spool hook still run the old email-only iav_hook.sh, so scan results never reach Sentinel, rewrite it and the event rules
if command -v imunify-antivirus >/dev/null 2>&1 && [ -d /etc/sysconfig/imunify360 ]; then
    if grep -q "setup_av_hook" /usr/local/opencli/imunify.sh 2>/dev/null; then
        opencli imunify hook || echo "Failed to set up the ImunifyAV hook, run 'opencli imunify hook' by hand"
    else
        echo "opencli is older than 2.0.24, run 'opencli update --cli' and then 'opencli imunify hook'"
    fi
fi

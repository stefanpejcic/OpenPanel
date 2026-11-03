#!/bin/bash

# opencli update runs this as wget -O - URL | bash, so a command below that reads stdin (podman) would eat the rest of the script and bash would go on mid-line; the rest is saved to a file and run from there with stdin closed
if [ -z "${BASH_SOURCE[0]:-}" ] || [ ! -f "${BASH_SOURCE[0]}" ]; then
    self=$(mktemp) && cat > "$self" && OP_UPDATE_SELF="$self" exec bash "$self" "$@" </dev/null
fi
[ -n "${OP_UPDATE_SELF:-}" ] && trap 'rm -f "$OP_UPDATE_SELF"' EXIT

# shellcheck disable=SC1091
. /usr/local/opencli/lib/podman.sh

# repairs compose files where the litespeed service / litespeed_conf volume were added with a different indentation than the rest of the file (e.g. 2 spaces in a file that uses 4)
# only those two blocks are re-indented, the rest of the file is not touched; runs first so that podman-compose below never reads a broken file
fix_litespeed_indent() {
    local f="$1" tmp rc
    [ -f "$f" ] || return 0
    grep -qE "^ +litespeed:[[:space:]]*$" "$f" || return 0
    tmp=$(mktemp)
    python3 - "$f" "$tmp" <<'PYFIX'
import re, sys
from collections import Counter

src, out = sys.argv[1], sys.argv[2]
lines = open(src).read().split('\n')

SERVICE_KEYS = {
    'image', 'container_name', 'healthcheck', 'hostname', 'restart', 'ports',
    'working_dir', 'volumes', 'deploy', 'networks', 'labels', 'environment',
    'env_file', 'depends_on', 'command', 'entrypoint', 'user', 'ulimits',
    'tmpfs', 'cap_add', 'cap_drop', 'privileged', 'build', 'expose',
    'extra_hosts', 'dns', 'devices', 'sysctls', 'stop_grace_period',
    'logging', 'security_opt', 'read_only', 'tty', 'stdin_open', 'pid',
    'ipc', 'mem_limit', 'cpus', 'platform',
}
VOLUME_KEYS = {'driver', 'labels', 'driver_opts', 'external', 'name'}


def indent(l):
    return len(l) - len(l.lstrip(' '))


def section(name):
    start = next((i for i, l in enumerate(lines) if re.match(r'^%s:[ \t]*$' % name, l)), None)
    if start is None:
        return None
    end = len(lines)
    for k in range(start + 1, len(lines)):
        l = lines[k]
        if l.strip() and not l.startswith((' ', '\t', '#')):
            end = k
            break
    return start, end


def unit_of(sec, probe):
    rng = section(sec)
    if not rng:
        return None
    a, b = rng
    cnt = Counter(indent(lines[k]) for k in range(a + 1, b) if re.match(r'^ +%s:' % probe, lines[k]))
    if not cnt:
        return None
    top = cnt.most_common(1)[0][0]
    return top // 2 if top % 2 == 0 and top > 0 else None


def fix(sec, key, probe, known):
    rng = section(sec)
    u = unit_of(sec, probe)
    if not rng or not u:
        return False
    a, b = rng
    i = next((k for k in range(a + 1, b) if re.match(r'^ +%s:[ \t]*$' % re.escape(key), lines[k])), None)
    if i is None:
        return False
    s = indent(lines[i])
    if s == u:
        return False
    j = i + 1
    while j < b:
        l = lines[j]
        if not l.strip():
            j += 1
            continue
        ind = indent(l)
        if ind <= s:
            break
        if ind == 2 * s:
            m = re.match(r'^ +([\w.-]+):', l)
            if m and m.group(1) not in known:
                break
        j += 1
    end_blk = j
    while end_blk > i + 1 and not lines[end_blk - 1].strip():
        end_blk -= 1
    for k in range(i, end_blk):
        l = lines[k]
        if l.strip():
            n = indent(l)
            lines[k] = ' ' * (n // s * u + n % s) + l.lstrip(' ')
    return True


def consistent(sec, probe):
    rng = section(sec)
    u = unit_of(sec, probe)
    if not rng or not u:
        return True
    a, b = rng
    return all(
        not l.strip() or l.lstrip().startswith('#') or indent(l) % u == 0
        for l in lines[a + 1:b]
    )


if any('\t' in l[:indent(l) + 1] for l in lines if l.strip()):
    sys.exit(1)

changed = fix('services', 'litespeed', 'container_name', SERVICE_KEYS)
changed |= fix('volumes', 'litespeed_conf', 'driver', VOLUME_KEYS)

if not changed:
    sys.exit(10)

if not (consistent('services', 'container_name') and consistent('volumes', 'driver')):
    sys.exit(1)

new_text = '\n'.join(lines)

try:
    import yaml
    d = yaml.safe_load(new_text)
    svcs = d['services']
    assert 'litespeed' in svcs
    assert all(isinstance(v, dict) and ('image' in v or 'container_name' in v) for v in svcs.values())
    assert 'litespeed_conf' in d['volumes']
except ImportError:
    pass
except Exception:
    sys.exit(1)

open(out, 'w').write(new_text)
PYFIX
    rc=$?
    if [ "$rc" -eq 0 ]; then
        [ -f "$f.bak-2.0.20" ] || cp -a "$f" "$f.bak-2.0.20"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
        echo "Fixed litespeed indentation in $f"
    elif [ "$rc" -ne 10 ]; then
        echo "Could not fix litespeed indentation in $f, left as it is."
    fi
    rm -f "$tmp"
}

fix_litespeed_indent /etc/openpanel/docker/compose/1.0/docker-compose.yml
for f in /home/*/docker-compose.yml; do
    fix_litespeed_indent "$f"
done

# openresty.conf was mounted at /etc/openresty/nginx.conf, which openresty never reads, so edits in Server Settings did nothing and their syntax check tested the image's own nginx.conf
# it's now mounted over the file openresty reads, a running openresty is recreated with it only if the account's file passes nginx -t, otherwise that account is left as it is
OLD_MOUNT="./openresty.conf:/etc/openresty/nginx.conf:ro"
NEW_MOUNT="./openresty.conf:/usr/local/openresty/nginx/conf/nginx.conf:ro"
RESTY_TEST_CONF=/usr/local/openresty/nginx/conf/op-2020-test.conf

openresty_mount() {
    local f="$1" tmp
    tmp=$(mktemp)
    sed "s|^\(\s*\)- \./openresty\.conf:/etc/openresty/nginx\.conf:ro\s*$|\1- $NEW_MOUNT|" "$f" > "$tmp"
    if grep -qF -- "- $NEW_MOUNT" "$tmp" && ! grep -qF -- "- $OLD_MOUNT" "$tmp"; then
        [ -f "$f.bak-2.0.20" ] || cp -a "$f" "$f.bak-2.0.20"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
        rm -f "$tmp"
        return 0
    fi
    echo "Could not update the openresty mount in $f, left as it is."
    rm -f "$tmp"
    return 1
}

grep -qF -- "- $OLD_MOUNT" /etc/openpanel/docker/compose/1.0/docker-compose.yml 2>/dev/null \
    && openresty_mount /etc/openpanel/docker/compose/1.0/docker-compose.yml

for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    grep -qF -- "- $OLD_MOUNT" "$f" || continue
    home_dir=$(dirname "$f")
    ctx=$(basename "$home_dir")

    if ! podman_ctx_is_running "$ctx" openresty </dev/null; then
        openresty_mount "$f"
        continue
    fi

    # a copy next to the stock nginx.conf resolves the relative mime.types include the way the moved file will
    if ! out=$(podman_ctx "$ctx" exec openresty sh -c "cp /etc/openresty/nginx.conf $RESTY_TEST_CONF && nginx -t -c $RESTY_TEST_CONF; rc=\$?; rm -f $RESTY_TEST_CONF; exit \$rc" 2>&1 </dev/null); then
        echo "$ctx's openresty.conf fails nginx -t, its openresty was left as it is. Fix $home_dir/openresty.conf and run this update again:"
        echo "$out"
        continue
    fi
    openresty_mount "$f" || continue

    echo "Recreating openresty for $ctx with its openresty.conf..."
    sock=$(podman_user_socket "$ctx") || continue
    (cd "$home_dir" && CONTAINER_HOST="$sock" timeout 300 podman-compose up -d --no-deps openresty) >/dev/null 2>&1 </dev/null
    if ! podman_ctx_is_running "$ctx" openresty </dev/null; then
        echo "openresty for $ctx is not running after the change, the previous compose file is $f.bak-2.0.20"
    fi
done

# files-list.sh gives each run its own temp files, it could run in two backup containers at once and fail the backup on a clashing mv
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
tmp=$(mktemp)
if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/backups/scripts/files-list.sh" && grep -q "TMP_SUFFIX" "$tmp"; then
    echo "Updating /etc/openpanel/backups/scripts/files-list.sh..."
    # written in place, backup containers bind-mount this file
    cat "$tmp" > /etc/openpanel/backups/scripts/files-list.sh
    chmod 755 /etc/openpanel/backups/scripts/files-list.sh
else
    echo "Failed to download $CONFIG_RAW/backups/scripts/files-list.sh"
fi
rm -f "$tmp"

# logrotate failed every night (caddy-logs postrotate ran podman exec inside logrotate.service, csf's lfd file was group-writable) and syslog had no size limit, the new opencli server-logrotate fixes all three
opencli server-logrotate

# caddy logged every coraza "key not found in collection" warning to syslog, gigabytes a day: default logger at ERROR and coraza debug log at errors only
CADDYFILE=/etc/openpanel/caddy/Caddyfile
CORAZA_RULES=/etc/openpanel/caddy/coraza_rules.conf
caddy_restart=no
if [ -f "$CADDYFILE" ] && ! awk '/^{$/{g=1} g && /^}$/{exit} g && /^    log \{/{f=1} END{exit !f}' "$CADDYFILE"; then
    cp -a "$CADDYFILE" "$CADDYFILE.bak-2.0.20"
    tmp=$(mktemp)
    awk '!done && /^{$/ {print; print "    # default logger only, site access logs have their own; the WAF warns \"key not found in collection\" on most requests and filled syslog"; print "    log {"; print "        level ERROR"; print "    }"; print ""; done=1; next} {print}' "$CADDYFILE" > "$tmp"
    # written in place, the caddy container bind-mounts this file
    cat "$tmp" > "$CADDYFILE"
    rm -f "$tmp"
    caddy_restart=yes
fi
if [ -f "$CORAZA_RULES" ] && ! grep -q "^SecDebugLogLevel" "$CORAZA_RULES"; then
    sed -i 's|^# LOGS$|# LOGS\n# errors only, at the default level coraza warns "key not found in collection" on most requests and that filled syslog\nSecDebugLogLevel 1|' "$CORAZA_RULES"
    caddy_restart=yes
fi
if [ "$caddy_restart" = yes ]; then
    # a reload keeps the waf logging at the old level, only a restart picks it up
    if podman exec caddy caddy validate --config /etc/caddy/Caddyfile >/dev/null 2>&1 </dev/null; then
        echo "Restarting caddy with ERROR log level..."
        podman restart caddy >/dev/null 2>&1 </dev/null
    else
        echo "Caddyfile fails caddy validate, restoring $CADDYFILE.bak-2.0.20"
        [ -f "$CADDYFILE.bak-2.0.20" ] && cat "$CADDYFILE.bak-2.0.20" > "$CADDYFILE"
    fi
fi

# LiteSpeed Enterprise is a new webserver choice: it reads Apache style vhosts, runs php in its bundled lsphp and starts on a trial license
CONFIG_RAW="${CONFIG_RAW:-https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main}"
TEMPLATE=/etc/openpanel/docker/compose/1.0/docker-compose.yml
TEMPLATE_ENV=/etc/openpanel/docker/compose/1.0/.env

fetch() {
    local src="$1" dst="$2" mode="$3" tmp
    tmp=$(mktemp)
    if wget --timeout=15 --tries=2 -q -O "$tmp" "$CONFIG_RAW/$src" && [ -s "$tmp" ]; then
        mkdir -p "$(dirname "$dst")"
        cat "$tmp" > "$dst"
        chmod "$mode" "$dst"
        rm -f "$tmp"
        return 0
    fi
    echo "Failed to download $CONFIG_RAW/$src"
    rm -f "$tmp"
    return 1
}

echo "Adding LiteSpeed Enterprise files..."
fetch litespeed/httpd.conf /etc/openpanel/litespeed/httpd.conf 644 || exit 0
fetch litespeed/start.sh /etc/openpanel/litespeed/start.sh 755 || exit 0
fetch nginx/vhosts/1.1/docker_litespeed_domain.conf /etc/openpanel/nginx/vhosts/1.1/docker_litespeed_domain.conf 644 || exit 0
NEW_TEMPLATE=$(mktemp)
trap 'rm -f "$OP_UPDATE_SELF" "$NEW_TEMPLATE"' EXIT
fetch docker/compose/1.0/docker-compose.yml "$NEW_TEMPLATE" 644 || exit 0

# adds the litespeed service and its volume to a compose file, the ports lines are taken from the file's own openlitespeed block so a varnish setup stays as it is
# the indentation is detected from the target file (2, 4, ... spaces) and the litespeed block is re-indented to match
add_litespeed_service() {
    local f="$1" tmp
    grep -qE "^ +litespeed:[[:space:]]*$" "$f" && return 0
    tmp=$(mktemp)
    if ! python3 - "$NEW_TEMPLATE" "$f" "$tmp" <<'EOF'
import re, sys
new, cur, out = open(sys.argv[1]).read(), open(sys.argv[2]).read(), sys.argv[3]

def unit(text, section):
    # number of spaces used to indent the first key under a top-level section
    m = re.search(r'(?m)^%s:[ \t]*\n(?:[ \t]*\n)*( +)[^\s#]' % section, text)
    return len(m.group(1)) if m else 2

def block(text, name, u):
    m = re.search(r'(?ms)^ {%d}%s:[ \t]*\n.*?(?=^ {%d}\S|^\S|\Z)' % (u, re.escape(name), u), text)
    return m.group(0) if m else None

def reindent(txt, src, dst):
    res = []
    for line in txt.splitlines(True):
        if not line.strip():
            res.append(line)
            continue
        n = len(line) - len(line.lstrip(' '))
        res.append(' ' * (n // src * dst + n % src) + line.lstrip(' '))
    return ''.join(res)

new_vol_text = new.split('\nvolumes:\n', 1)[1] if '\nvolumes:\n' in new else ''
new_vol_text = 'volumes:\n' + new_vol_text
cur_svc_u, new_svc_u = unit(cur, 'services'), unit(new, 'services')
cur_vol_u, new_vol_u = unit(cur, 'volumes'), unit(new_vol_text, 'volumes')

ls = block(new, 'litespeed', new_svc_u)
if not ls:
    sys.exit(1)
ls = reindent(ls, new_svc_u, cur_svc_u)
if not ls.endswith('\n'):
    ls += '\n'

ols = block(cur, 'openlitespeed', cur_svc_u)
if ols:
    pat = r'(?m)^ {%d}ports:[ \t]*\n(?: {%d}- .*\n)+' % (2 * cur_svc_u, 3 * cur_svc_u)
    ports = re.search(pat, ols)
    if ports:
        ls = re.sub(pat, lambda _: ports.group(0), ls, count=1)
    cur = cur.replace(ols, ols + ls, 1)
else:
    cur = re.sub(r'(?m)^services:[ \t]*\n', lambda m: m.group(0) + ls, cur, count=1)

vol = block(new_vol_text, 'litespeed_conf', new_vol_u)
if vol and not re.search(r'(?m)^ +litespeed_conf:', cur):
    vol = reindent(vol, new_vol_u, cur_vol_u)
    if not vol.endswith('\n'):
        vol += '\n'
    if re.search(r'(?m)^volumes:[ \t]*\n', cur):
        cur = re.sub(r'(?m)^volumes:[ \t]*\n', lambda m: m.group(0) + vol, cur, count=1)
    else:
        cur = cur.rstrip('\n') + '\n\nvolumes:\n' + vol

# if PyYAML is available, verify the result before it is written
try:
    import yaml
    d = yaml.safe_load(cur)
    assert 'litespeed' in d['services'] and 'litespeed_conf' in d['volumes']
    assert 'networks' in d['services']['litespeed']
    assert len(d['services']) >= 2
except ImportError:
    pass
open(out, 'w').write(cur)
EOF
    then
        echo "Could not add the litespeed service to $f, left as it is."
        rm -f "$tmp"
        return 1
    fi
    if grep -qE "^ +litespeed:[[:space:]]*$" "$tmp" && grep -qE "^ +litespeed_conf:[[:space:]]*$" "$tmp"; then
        [ -f "$f.bak-2.0.20" ] || cp -a "$f" "$f.bak-2.0.20"
        # written in place so the file keeps its owner and mode
        cat "$tmp" > "$f"
    else
        echo "Could not add the litespeed service to $f, left as it is."
    fi
    rm -f "$tmp"
}

add_litespeed_env() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q "^LITESPEED_VERSION=" "$f" && return 0
    printf '\n# LITESPEED (tag is <release>-lsphpXX, the php version comes with the image)\nLITESPEED_VERSION="latest"\nLITESPEED_CPU="1.0"\nLITESPEED_RAM="1.0G"\nLITESPEED_PIDS="2500"\n' >> "$f"
}

add_litespeed_service "$TEMPLATE"
add_litespeed_env "$TEMPLATE_ENV"

for f in /home/*/docker-compose.yml; do
    [ -f "$f" ] || continue
    home_dir=$(dirname "$f")
    ctx=$(basename "$home_dir")
    [ -f "$home_dir/.env" ] || continue
    grep -qE "^ +litespeed:[[:space:]]*$" "$f" && [ -f "$home_dir/litespeed.conf" ] && continue
    add_litespeed_service "$f"
    add_litespeed_env "$home_dir/.env"
    if [ ! -f "$home_dir/litespeed.conf" ]; then
        cp /etc/openpanel/litespeed/httpd.conf "$home_dir/litespeed.conf"
        chown --reference="$home_dir/.env" "$home_dir/litespeed.conf"
    fi
    echo "LiteSpeed is now available to $ctx."
done

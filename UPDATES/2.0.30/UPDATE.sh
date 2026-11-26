#!/bin/bash

# GHSA-3mrp-68rg-c3fj: the panel Redis kept a TCP listener on 0.0.0.0:6379 with no password, so any account
# container could reach it at 172.20.0.4:6379 and rewrite the cache. Disable TCP, the panel only uses the socket.
for compose in /root/docker-compose.yml /root/compose.yml; do
    [ -f "$compose" ] || continue
    if grep -q '\["redis-server", "--unixsocket"' "$compose"; then
        echo "Closing the panel Redis TCP port in $compose..."
        cp -a "$compose" "$compose.bak-2.0.30"
        sed -i 's#\["redis-server", "--unixsocket"#["redis-server", "--port", "0", "--unixsocket"#' "$compose"
        recreate_redis=1
    fi
done
if [ "$recreate_redis" = "1" ]; then
    if cd /root && podman-compose up -d --force-recreate openpanel_redis >/dev/null 2>&1; then
        echo "Recreated openpanel_redis with the TCP port disabled (unix socket only)"
    else
        echo "Patched the compose file but could not recreate openpanel_redis, do it by hand: cd /root && podman-compose up -d --force-recreate openpanel_redis"
    fi
fi

# PR #18: in the OpenLiteSpeed lsphp block, match maxConns to PHP_LSAPI_CHILDREN and turn persistConn off,
# so lsphp doesn't open more backend connections than it has children. The per-user configs are copies of the
# host template made when the account was created, patch both. It takes effect when the OLS container restarts.
patch_ols_conf() {
    local f="$1"
    [ -f "$f" ] || return 0
    grep -q '^extprocessor lsphp' "$f" || return 0
    local tmp
    tmp=$(mktemp) || return 0
    awk '
        { lines[NR]=$0 }
        /^extprocessor lsphp[ \t]*\{/ { in_lsphp=1 }
        in_lsphp && /PHP_LSAPI_CHILDREN=[0-9]+/ {
            s=$0; sub(/.*PHP_LSAPI_CHILDREN=/,"",s); sub(/[^0-9].*/,"",s); children=s
        }
        in_lsphp && /^[ \t]*}/ { in_lsphp=0 }
        END {
            if (children=="") children=5
            inb=0
            for (i=1;i<=NR;i++){
                l=lines[i]
                if (l ~ /^extprocessor lsphp[ \t]*\{/) inb=1
                if (inb && l ~ /^[ \t]*maxConns[ \t]/)    sub(/maxConns[ \t]+[0-9]+/,    "maxConns                " children, l)
                if (inb && l ~ /^[ \t]*persistConn[ \t]/) sub(/persistConn[ \t]+[0-9]+/, "persistConn             0", l)
                if (inb && l ~ /^[ \t]*}/) inb=0
                print l
            }
        }' "$f" > "$tmp" || { rm -f "$tmp"; return 0; }
    if ! cmp -s "$f" "$tmp"; then
        cat "$tmp" > "$f"
        echo "Patched the lsphp block in $f"
    fi
    rm -f "$tmp"
}

patch_ols_conf /etc/openpanel/openlitespeed/httpd_config.conf
for uconf in /home/*/openlitespeed.conf; do
    patch_ols_conf "$uconf"
done

#!/bin/bash

# NOTE
# https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/1.0/autostart.services
# for this test we enable services per user to stimulate real usage:
# default php version, cron, memcached
FILE=/etc/openpanel/docker/compose/1.0/autostart.services
ENV=/etc/openpanel/docker/compose/1.0/.env
PHPV=$(grep -E '^DEFAULT_PHP_VERSION=' "$ENV" | cut -d= -f2- | tr -d '\r"'"'")
PHP_SVC="php-fpm-${PHPV}"

sed -i '/^varnish$/d' "$FILE"
grep -qxF "$PHP_SVC"  "$FILE" || echo "$PHP_SVC"  >> "$FILE"
grep -qxF 'cron'      "$FILE" || echo 'cron'      >> "$FILE"
grep -qxF 'memcached' "$FILE" || echo 'memcached' >> "$FILE"

echo "Services that will auto-start per user:"
cat "$FILE"
echo


# defaults
LICENSE_KEY="enterprise-2a5da40ecd2f60" # ip restricted!
opencli config update key $LICENSE_KEY

PANEL_PASSWORD="testingpassword"
PLAN="Standard plan"
# stop thresholds
MIN_DISK_MB=1024   # stop when < 1GB free on /
MIN_RAM_MB=512     # stop when < 512MB available RAM

rand_user() { echo "test$(tr -dc 'a-z0-9' </dev/urandom | head -c8)"; }
disk_free_mb() { df -BM --output=avail / | awk 'NR==2{gsub("M","");print $1}'; }
ram_free_mb()  { free -m | awk '/^Mem:/{print $7}'; }   # $7 = available

echo "=== START CREATING USERS ==="
START_DISK=$(disk_free_mb)
echo "disk_free_mb=$START_DISK ram_free_mb=$(ram_free_mb)"

CREATED=()
START=$(date +%s)
i=0

cd /home

while :; do
    DF=$(disk_free_mb); RF=$(ram_free_mb)
    if (( DF < MIN_DISK_MB )); then echo ">>> STOP: disk ${DF}MB < ${MIN_DISK_MB}MB"; break; fi
    if (( RF < MIN_RAM_MB )); then echo ">>> STOP: ram ${RF}MB < ${MIN_RAM_MB}MB"; break; fi

    i=$((i+1))
    U=$(rand_user)
    while opencli user-list 2>/dev/null | grep -qw "$U"; do U=$(rand_user); done

    T0=$(date +%s)
    opencli user-add "$U" "$PANEL_PASSWORD" "$U@test.com" "$PLAN" --debug >/tmp/ua.log 2>&1
    T1=$(date +%s)
    DUR=$((T1-T0))

    if opencli user-list 2>/dev/null | grep -qw "$U"; then
        CREATED+=("$U")
        # wait for this user's containers to finish coming up before measuring next
        UID_N=$(id -u "$U" 2>/dev/null)
        for _ in $(seq 1 60); do
            running=$(sudo -u "$U" XDG_RUNTIME_DIR=/run/user/$UID_N podman ps -q 2>/dev/null | wc -l)
            (( running > 0 )) && sleep 2 && break
            sleep 1
        done
        # wait for load to fall below 2x nproc; check every 30s, give up after 5min
        THRESH=$(( $(nproc) * 2 ))
        waited=0
        while (( $(awk '{print int($1)}' /proc/loadavg) > THRESH )); do
            (( waited >= 300 )) && { echo "  load high (>${THRESH}) after 5min, continuing anyway"; break; }
            sleep 30
            waited=$((waited+30))
        done
    fi
done

END=$(date +%s)
N=${#CREATED[@]}
END_DISK=$(disk_free_mb)
USED=$(( START_DISK - END_DISK ))
TOTAL=$((END-START))

echo "=== DONE ==="
echo "users_created=$N  total_time=${TOTAL}s"
echo "disk_free_mb=$END_DISK ram_free_mb=$(ram_free_mb)"
if (( N > 0 )); then
    echo "total_disk_used_mb=$USED avg_disk_per_user_mb=$(( USED / N ))"
    echo "avg_time_per_user=$(( TOTAL / N ))s"
fi

printf '%s\n' "${CREATED[@]}" > /root/created_test_users.txt
echo "saved list -> /root/created_test_users.txt"

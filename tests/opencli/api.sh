#!/bin/bash
################################################################################
# OpenAdmin API smoke test
# Tests every endpoint 1 by 1, prints PASS/FAIL per route, summary at the end.
#
# Usage:
#   ./api_test.sh                          # safe (read-only) tests only
#                                          #   as root ON the server itself!)
#   REBOOT_TEST=1 ./api_test.sh            # actually reboot the server at the end
#   ONLY="/api/users" ./api_test.sh        # run only routes matching a pattern
#   ONLY="settings" ./api_test.sh          # substring/glob match works too
#   ONLY="GET /api/domains" ./api_test.sh  # optionally prefix with method
#
# Config via env vars or edit below.
################################################################################

BASE_URL="${BASE_URL:-$(opencli admin | grep -oE 'https://[^ ]+' | head -n1)}"
ADMIN_USER="${ADMIN_USER:-}"
ADMIN_PASS="${ADMIN_PASS:-}"
REBOOT_TEST="${REBOOT_TEST:-0}"
ONLY="${ONLY:-}"

# Test fixtures
TEST_USER="apitest_$(date +%s)"
TEST_DOMAIN="apitest-$(date +%s).com"
TEST_PLAN="Standard plan"

PASS=0; FAIL=0; SKIP=0
FAILED_ROUTES=()

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'; NC='\033[0m'

if [ -z "$ADMIN_PASS" ]; then
    read -rsp "Admin password for $ADMIN_USER: " ADMIN_PASS
    echo
fi

# enable api first!
opencli api on
cd /root && docker compose up -d bind9
#opencli email-server install

################################################################################
# Helpers
################################################################################

# matches_only <method> <path>
# Returns 0 (run the test) when ONLY is empty or the route matches it.
# ONLY can be:
#   - a substring of the path:        ONLY="/api/users"  ONLY="settings"
#   - a glob pattern:                 ONLY="/api/domains/*"
#   - method + pattern:               ONLY="POST /api/users"
matches_only() {
    local method="$1" path="$2"
    [ -z "$ONLY" ] && return 0

    local want_method="" pattern="$ONLY"
    # If ONLY starts with an HTTP method, split it off
    case "$ONLY" in
        GET\ *|POST\ *|PUT\ *|PATCH\ *|DELETE\ *)
            want_method="${ONLY%% *}"
            pattern="${ONLY#* }"
            ;;
    esac

    if [ -n "$want_method" ] && [ "$want_method" != "$method" ]; then
        return 1
    fi

    # Match as glob, or as plain substring
    [[ "$path" == $pattern ]] && return 0
    [[ "$path" == *"$pattern"* ]] && return 0
    return 1
}

# test_api <method> <path> <expected_status(es)> [json_body]
test_api() {
    local method="$1" path="$2" expected="$3" body="$4"

    if ! matches_only "$method" "$path"; then
        SKIP=$((SKIP+1))
        return 0
    fi

    local args=(-s -o /tmp/api_test_body.json -w "%{http_code}" -X "$method"
                -H "Authorization: Bearer $TOKEN"
                --max-time 30
                "$BASE_URL$path")
    [ -n "$body" ] && args+=(-H "Content-Type: application/json" -d "$body")

    local code
    code=$(curl "${args[@]}")

    if [[ ",$expected," == *",$code,"* ]]; then
        printf "${GREEN}PASS${NC} [%s] %-55s -> %s\n" "$method" "$path" "$code"
        PASS=$((PASS+1))
    else
        printf "${RED}FAIL${NC} [%s] %-55s -> %s (expected %s)\n" "$method" "$path" "$code" "$expected"
        head -c 300 /tmp/api_test_body.json; echo
        FAIL=$((FAIL+1))
        FAILED_ROUTES+=("[$method] $path -> $code")
    fi
}

skip() {
    printf "${YELLOW}SKIP${NC} [%s] %-55s (%s)\n" "$1" "$2" "$3"
    SKIP=$((SKIP+1))
}

# manual_pass/manual_fail <method> <path> <detail>
# For tests where PASS/FAIL is decided by checking the OS, not the HTTP code.
manual_pass() {
    printf "${GREEN}PASS${NC} [%s] %-55s (%s)\n" "$1" "$2" "$3"
    PASS=$((PASS+1))
}
manual_fail() {
    printf "${RED}FAIL${NC} [%s] %-55s (%s)\n" "$1" "$2" "$3"
    FAIL=$((FAIL+1))
    FAILED_ROUTES+=("[$1] $2 - $3")
}

# raw API call, returns http code, body in /tmp/api_test_body.json
api_call() {
    local method="$1" path="$2" body="$3"
    local args=(-s -o /tmp/api_test_body.json -w "%{http_code}" -X "$method"
                -H "Authorization: Bearer $TOKEN"
                --max-time 30
                "$BASE_URL$path")
    [ -n "$body" ] && args+=(-H "Content-Type: application/json" -d "$body")
    curl "${args[@]}"
}

################################################################################
# 0. Auth
################################################################################
echo "== Authenticating =="
STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/")
echo "GET /api/ -> $STATUS_CODE"

TOKEN=$(curl -s -X POST "$BASE_URL/api/" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" \
    | grep -o '"access_token"[: ]*"[^"]*"' | sed 's/.*"\([^"]*\)"$/\1/')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Could not obtain token — aborting.${NC}"
    exit 1
fi
echo "Token obtained (${#TOKEN} chars)"
[ -n "$ONLY" ] && echo -e "${YELLOW}Filter active:${NC} only running routes matching '$ONLY'"
echo

################################################################################
# 1. Read-only GET endpoints (always safe)
################################################################################
echo "== Read-only endpoints =="
test_api GET  "/api/whoami"                       200
test_api GET  "/api/users"                        200
test_api GET  "/api/domains"                      200
test_api GET  "/api/plans"                        200
test_api GET  "/api/plans/1"                      "200,404"
test_api GET  "/api/services"                     200
test_api GET  "/api/services/status"              200
test_api GET  "/api/docker/info"                  200
test_api GET  "/api/ips"                          200
test_api GET  "/api/system"                       200
test_api GET  "/api/usage/cpu"                    200
test_api GET  "/api/usage/memory"                 200
test_api GET  "/api/usage/server"                 200
test_api GET  "/api/usage/disk"                   200
test_api GET  "/api/notifications"                200
test_api GET  "/api/dns/cluster"                  200
test_api GET  "/api/dns/zone-templates"           200
test_api GET  "/api/domains/file-templates"       200
test_api GET  "/api/emails/settings"              200
test_api GET  "/api/emails/accounts"              200
test_api GET  "/api/emails/queue"                 200
test_api GET  "/api/emails/domain-limits"         200
test_api GET  "/api/security/basic-auth"          200
test_api GET  "/api/security/blacklist-useragents" 200
test_api GET  "/api/security/firewall"            200
test_api GET  "/api/security/waf"                 200
test_api GET  "/api/security/waf/rules"           200
test_api GET  "/api/security/2fa"                 200
test_api GET  "/api/security/passkeys"            200
test_api GET  "/api/server/crons"                 200
test_api GET  "/api/server/ssh"                   200
test_api GET  "/api/server/ssh/config"            200
test_api GET  "/api/server/timezone"              200
test_api GET  "/api/server/processes?sort=memory" 200
test_api GET  "/api/server/node"                  200
test_api GET  "/api/server/reboot/status"         200
test_api GET  "/api/server/migrate"               200
test_api GET  "/api/settings/administrators"      200
test_api GET  "/api/settings/resellers"           200
test_api GET  "/api/settings/general"             200
test_api GET  "/api/settings/defaults"            200
test_api GET  "/api/settings/defaults/files"      200
test_api GET  "/api/settings/features"            200
test_api GET  "/api/settings/features/default"    200
test_api GET  "/api/settings/locales"             200
test_api GET  "/api/settings/modules"             200
test_api GET  "/api/settings/custom-code"         200
test_api GET  "/api/settings/php"                 200
test_api GET  "/api/settings/caddy/metrics"       200
test_api GET  "/api/settings/updates"             200
test_api GET  "/api/settings/updates/tags"        200
test_api GET  "/api/settings/notifications"       200
test_api GET  "/api/license"                      200
test_api GET  "/api/license/info"                 200
test_api GET  "/api/support/report"               200
test_api GET  "/api/import/cpanel"                200
test_api GET  "/api/import/backup-files"          200
test_api GET  "/api/import/transfers"             200
echo

################################################################################
# 2. Auth negative test (no token should be rejected)
################################################################################
if matches_only GET "/api/whoami"; then
    echo "== Auth checks =="
    NOAUTH=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/whoami")
    if [[ "$NOAUTH" == "401" || "$NOAUTH" == "403" ]]; then
        printf "${GREEN}PASS${NC} [GET] %-55s -> %s (rejected without token)\n" "/api/whoami (no auth)" "$NOAUTH"
        PASS=$((PASS+1))
    else
        printf "${RED}FAIL${NC} [GET] %-55s -> %s (should be 401/403!)\n" "/api/whoami (no auth)" "$NOAUTH"
        FAIL=$((FAIL+1))
        FAILED_ROUTES+=("[GET] /api/whoami without token -> $NOAUTH")
    fi
    echo
fi

################################################################################
# 3. Write tests — full lifecycle with a throwaway user/domain
################################################################################
# PREPARATIONS
#echo "== Preparing server.. =="
#opencli email-server installl && opencli email-server start
# todo: bind and ftp

echo "== Write tests (user: $TEST_USER, domain: $TEST_DOMAIN) =="

# User lifecycle
test_api POST "/api/users" "200,201" \
    "{\"email\":\"$TEST_USER@example.com\",\"username\":\"$TEST_USER\",\"password\":\"Test_$(date +%s)!\",\"plan_name\":\"$TEST_PLAN\"}"

# Domain lifecycle
test_api POST "/api/domains/new" "200,201" \
    "{\"username\":\"$TEST_USER\",\"domain\":\"$TEST_DOMAIN\",\"docroot\":\"/var/www/html/$TEST_DOMAIN\"}"
test_api GET  "/api/domains/docroot/$TEST_DOMAIN"       200
test_api GET  "/api/domains/$TEST_DOMAIN/dns"           200
test_api GET  "/api/domains/$TEST_DOMAIN/caddy"         200
test_api GET  "/api/domains/$TEST_DOMAIN/vhost/$TEST_USER" 200
test_api GET  "/api/domains/$TEST_DOMAIN/ssl"           200
test_api GET  "/api/domains/$TEST_DOMAIN/log"           "200,404"
test_api POST "/api/domains/suspend/$TEST_DOMAIN"       200
test_api POST "/api/domains/unsuspend/$TEST_DOMAIN"     200
test_api POST "/api/domains/delete/$TEST_DOMAIN"        200

# Containers for the test user
test_api GET  "/api/users/$TEST_USER/containers"        200

# Plan lifecycle
test_api POST "/api/plans" "200,201" \
    '{"name":"apitest_plan","description":"test","email_limit":"1","ftp_limit":"1","domains_limit":"1","websites_limit":"1","disk_limit":"1000","inodes_limit":"10000","db_limit":"1","cpu":"1","ram":"1","bandwidth":"10","feature_set":"default"}'
# NOTE: fetch id of apitest_plan and DELETE it manually, or add jq parsing here

# Cleanup: delete the test user (adjust to your actual user-delete route)
test_api DELETE "/api/users" "200,204,404,405" "{\"username\":\"$TEST_USER\"}"
echo

################################################################################
# 4. Destructive tests — verified against the OS, then reverted.
################################################################################

echo "== Destructive tests =="

if [ "$(id -u)" != "0" ] || ! command -v opencli >/dev/null 2>&1; then
    echo -e "${RED}Destructive tests require root on the OpenPanel server itself (opencli + /etc/shadow access) — skipping section.${NC}"
    skip "ALL" "destructive tests" "not root / not on server"
else
    # ------------------------------------------------------------------
    # 4.1 /api/server/root-password — change, verify hash changed, revert
    # ------------------------------------------------------------------
    if matches_only POST "/api/server/root-password"; then
        ORIG_HASH=$(grep '^root:' /etc/shadow | cut -d: -f2)
        NEW_ROOT_PW="Apitest_$(date +%s)!x"
        CODE=$(api_call POST "/api/server/root-password" "{\"password\":\"$NEW_ROOT_PW\"}")
        NEW_HASH=$(grep '^root:' /etc/shadow | cut -d: -f2)
        if [[ "$CODE" == "200" && "$NEW_HASH" != "$ORIG_HASH" ]]; then
            manual_pass POST "/api/server/root-password" "hash changed in /etc/shadow"
        else
            manual_fail POST "/api/server/root-password" "code=$CODE hash_changed=$([ "$NEW_HASH" != "$ORIG_HASH" ] && echo yes || echo no)"
        fi
        # Revert: put the original hash back (we never knew the plaintext)
        usermod -p "$ORIG_HASH" root
        RESTORED_HASH=$(grep '^root:' /etc/shadow | cut -d: -f2)
        if [ "$RESTORED_HASH" == "$ORIG_HASH" ]; then
            echo "     root password hash restored to original"
        else
            echo -e "${RED}     WARNING: could not restore original root hash — fix manually!${NC}"
        fi
    fi

    # ------------------------------------------------------------------
    # 4.2 /api/server/memory/drop-* — compare cache before vs after
    # ------------------------------------------------------------------
    for DROP in drop-caches drop-pagecache drop-inodes; do
        if matches_only POST "/api/server/memory/$DROP"; then
            # warm the page cache a little so there is something to drop
            find /usr/lib -type f -print0 2>/dev/null | head -z -n 500 | xargs -0 cat > /dev/null 2>&1
            CACHE_BEFORE=$(awk '/^Cached:/ {print $2}' /proc/meminfo)
            CODE=$(api_call POST "/api/server/memory/$DROP")
            sleep 1
            CACHE_AFTER=$(awk '/^Cached:/ {print $2}' /proc/meminfo)
            if [[ "$CODE" == "200" || "$CODE" == "404" ]]; then
                if [ "$CODE" == "404" ]; then
                    skip POST "/api/server/memory/$DROP" "route not present"
                elif [ "$CACHE_AFTER" -lt "$CACHE_BEFORE" ]; then
                    manual_pass POST "/api/server/memory/$DROP" "Cached: ${CACHE_BEFORE}kB -> ${CACHE_AFTER}kB"
                else
                    manual_fail POST "/api/server/memory/$DROP" "cache did not shrink (${CACHE_BEFORE}kB -> ${CACHE_AFTER}kB)"
                fi
            else
                manual_fail POST "/api/server/memory/$DROP" "code=$CODE"
            fi
        fi
    done

    # ------------------------------------------------------------------
    # 4.3 /api/settings/updates/now — safe no-op when no update available
    # ------------------------------------------------------------------
    test_api POST "/api/settings/updates/now" "200,201,202"
    
    # ------------------------------------------------------------------
    # 4.4 /api/server/processes/<pid>/kill — spawn dummy, kill via API,
    #     verify it is actually gone on the OS
    # ------------------------------------------------------------------
    if matches_only POST "/api/server/processes/<pid>/kill"; then
        sleep 300 &
        DUMMY_PID=$!
        disown "$DUMMY_PID" 2>/dev/null
        CODE=$(api_call POST "/api/server/processes/$DUMMY_PID/kill")
        sleep 2
        if kill -0 "$DUMMY_PID" 2>/dev/null; then
            manual_fail POST "/api/server/processes/$DUMMY_PID/kill" "code=$CODE but PID $DUMMY_PID still alive"
            kill -9 "$DUMMY_PID" 2>/dev/null   # clean up ourselves
        else
            if [ "$CODE" == "200" ]; then
                manual_pass POST "/api/server/processes/$DUMMY_PID/kill" "PID $DUMMY_PID confirmed dead"
            else
                manual_fail POST "/api/server/processes/$DUMMY_PID/kill" "PID gone but code=$CODE"
            fi
        fi
    fi

    # ------------------------------------------------------------------
    # 4.5 /api/security/disable-admin — panel must stop responding,
    #     then restore with `opencli admin on` and confirm it is back.
    #     Runs LAST in this section because it takes the API down.
    # ------------------------------------------------------------------
    if matches_only POST "/api/security/disable-admin"; then
        CODE=$(api_call POST "/api/security/disable-admin")
        sleep 3
        DOWN_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$BASE_URL/api/" || echo "000")
        opencli admin on >/dev/null 2>&1
        # give the panel time to come back
        UP_CODE="000"
        for i in $(seq 1 15); do
            sleep 2
            UP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$BASE_URL/api/" || echo "000")
            [ "$UP_CODE" != "000" ] && break
        done
        if [[ "$DOWN_CODE" == "000" || "$DOWN_CODE" == "502" || "$DOWN_CODE" == "503" ]] && [ "$UP_CODE" != "000" ]; then
            manual_pass POST "/api/security/disable-admin" "went down ($DOWN_CODE), restored via opencli admin on ($UP_CODE)"
        elif [ "$UP_CODE" == "000" ]; then
            manual_fail POST "/api/security/disable-admin" "PANEL STILL DOWN — run 'opencli admin on' manually!"
        else
            manual_fail POST "/api/security/disable-admin" "code=$CODE, panel still answered while disabled ($DOWN_CODE)"
        fi
    fi

fi # root check
echo


################################################################################
# 5. Reboot — only with explicit REBOOT_TEST=1, always last.
#    We just fire it; the script (and the server) will not survive to verify.
################################################################################
if [ "$REBOOT_TEST" == "1" ]; then
    echo "== Reboot test =="
    echo -e "${YELLOW}Firing /api/server/reboot in 5 seconds — Ctrl+C to abort!${NC}"
    sleep 5
    CODE=$(api_call POST "/api/server/reboot")
    echo "POST /api/server/reboot -> $CODE (server should be going down now)"
    [ "$CODE" == "200" ] && PASS=$((PASS+1)) || { FAIL=$((FAIL+1)); FAILED_ROUTES+=("[POST] /api/server/reboot -> $CODE"); }
else
    skip "POST" "/api/server/reboot" "set REBOOT_TEST=1 to enable"
fi
echo

################################################################################
# Summary
################################################################################
echo "======================================"
echo -e "  ${GREEN}PASS: $PASS${NC}   ${RED}FAIL: $FAIL${NC}   ${YELLOW}SKIP: $SKIP${NC}"
echo "======================================"
if [ ${#FAILED_ROUTES[@]} -gt 0 ]; then
    echo "Failed routes:"
    printf '  %s\n' "${FAILED_ROUTES[@]}"
    exit 1
fi
exit 0

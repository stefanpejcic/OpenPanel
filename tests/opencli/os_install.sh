#!/bin/bash

GITHUB_REPO="stefanpejcic/OpenPanel"

# NEEDED DATA:
: '
GITHUB_TOKEN

SSH_PUBLIC_KEY
SSH_PRIVATE_FILE
SSH_PRIVATE_KEY

VIRTUALIZOR_API
API_KEY
API_HASH
VPSID
VPS_IP
TEST_PASS

DOCKER_HUB_TOKEN
DISCORD_WEBHOOK
'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


# read admin logins
ENV_FILE="$SCRIPT_DIR/../openadmin/.env"

if [ -f "$ENV_FILE" ]; then
    set -a
    source "$ENV_FILE"
    set +a
else
    echo "Warning: $ENV_FILE not found"
fi

TEST_ADMIN_USER="${PANEL_USERNAME:-stefan}"
TEST_ADMIN_PASS="${PANEL_PASSWORD:-stefan}"

declare -A OS_MAP=(
  ["ubuntu-22"]=1017
  ["ubuntu-24"]=1108
  ["ubuntu-26"]=1215

  ["debian-11"]=979
  ["debian-12"]=1055
  ["debian-13"]=1188

  ["almalinux-8"]=1081
  ["almalinux-9"]=1200
  ["almalinux-10"]=1176

  ["rocky-8"]=1078
  ["rocky-10"]=1182

  ["centos-10"]=1179
)

declare -A TEST_RESULTS
declare -A TEST_TIMESTAMPS
declare -A TEST_INSTALL_TIME

#############################################################
# LOGGING SETUP
#############################################################
LOG_DIR="/root/os-logs"
RUN_TS="$(date '+%Y%m%d-%H%M%S')"
RUN_DIR="$LOG_DIR/$RUN_TS"
mkdir -p "$RUN_DIR"

MAIN_LOG="$RUN_DIR/main.log"
touch "$MAIN_LOG"

# Also keep a "latest" symlink so you don't have to hunt for the newest run dir
ln -sfn "$RUN_DIR" "$LOG_DIR/latest"

# log <message> -- writes to stdout (so cron mail/journal still gets it if configured)
# AND to the main run log file.
log() {
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
  echo "$msg" | tee -a "$MAIN_LOG"
}

# os_log <os> <message> -- per-OS log file, also echoed to main log
os_log() {
  local os="$1"; shift
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
  echo "$msg" >> "$RUN_DIR/${os}.log"
  log "[$os] $*"
}

#############################################################

/usr/sbin/csf -a 185.119.89.4 >>"$MAIN_LOG" 2>&1
/usr/sbin/csf -a "$VPS_IP" >>"$MAIN_LOG" 2>&1

echo "$SSH_PRIVATE_KEY" > "$SSH_PRIVATE_FILE"
chmod 600 "$SSH_PRIVATE_FILE"

send_discord() {
  local resp
  resp=$(/usr/bin/curl -s -o /dev/null -w "%{http_code}" -H "Content-Type: application/json" -X POST -d "{\"content\": \"$1\"}" "$DISCORD_WEBHOOK")
  log "DISCORD: sent (http $resp): $1"
}

# format_duration <seconds> -- prints "1h2m3s" / "2m3s" / "3s" depending on magnitude
format_duration() {
  local total_seconds="$1"
  local h=$((total_seconds / 3600))
  local m=$(((total_seconds % 3600) / 60))
  local s=$((total_seconds % 60))
  if [[ $h -gt 0 ]]; then
    printf '%dh%dm%ds' "$h" "$m" "$s"
  elif [[ $m -gt 0 ]]; then
    printf '%dm%ds' "$m" "$s"
  else
    printf '%ds' "$s"
  fi
}

reinstall_os() {
  local os="$1"
  local os_id="${OS_MAP[$os]}"
  if [[ -z "$os_id" ]]; then
    os_log "$os" "ERROR: Unknown OS: $os"
    send_discord "❌ Unknown OS: $os"
    return 1
  fi
  os_log "$os" "Reinstalling VPS with OS: $os (ID: $os_id)"

  local resp
  resp=$(/usr/bin/curl -sk -X POST -d "reinsos=1&newos=$os_id&newpass=$TEST_PASS&conf=$TEST_PASS" -L "$VIRTUALIZOR_API/index.php?act=ostemplate&svs=$VPSID&api=json&apikey=$API_KEY&apipass=$API_HASH")
  local rc=$?
  echo "$resp" >> "$RUN_DIR/${os}.log"

  if [[ $rc -ne 0 ]]; then
    os_log "$os" "ERROR: Failed to trigger reinstall for $os (curl exit $rc)"
    send_discord "❌ [$os] Reinstall failed (curl exit $rc)"
    return 1
  fi
  os_log "$os" "Reinstall triggered for $os"
}

wait_for_ssh() {
  local ip="$1"
  local os="$2"
  os_log "$os" "Waiting for SSH on $ip..."
  for i in {1..60}; do
    if /usr/bin/nc -z -w3 "$ip" 22 2>>"$RUN_DIR/${os}.log"; then
      os_log "$os" "SSH is up on $ip (attempt $i)"
      return 0
    fi
    os_log "$os" "SSH not ready yet on $ip, attempt $i/60..."
    sleep 10
  done
  os_log "$os" "ERROR: SSH timeout after 10 minutes on $ip"
  send_discord "❌ [$os] SSH timeout after 10 minutes on $ip"
  return 1
}

update_readme_results() {
  log "Updating README OS table with test results..."

  local api_response
  api_response=$(/usr/bin/curl -s -H "Authorization: token $GITHUB_TOKEN" -H "Accept: application/vnd.github.v3+json" "https://api.github.com/repos/$GITHUB_REPO/contents/README.md")

  local sha
  sha=$(echo "$api_response" | /usr/bin/python3 -c "import sys,json; print(json.load(sys.stdin)['sha'])" 2>>"$MAIN_LOG")

  local current_content
  current_content=$(echo "$api_response" | /usr/bin/python3 -c "
import sys, json, base64
d = json.load(sys.stdin)
print(base64.b64decode(d['content']).decode('utf-8'), end='')
" 2>>"$MAIN_LOG")

  if [[ -z "$sha" ]]; then
    log "ERROR: Could not fetch README from GitHub API. API response: $api_response"
    return 1
  fi

  local results_json="{"
  for os in "${!TEST_RESULTS[@]}"; do
    local install_time_json="null"
    if [[ "${TEST_RESULTS[$os]}" == "pass" && -n "${TEST_INSTALL_TIME[$os]:-}" ]]; then
      install_time_json="\"${TEST_INSTALL_TIME[$os]}\""
    fi
    results_json+="\"$os\": {\"status\": \"${TEST_RESULTS[$os]}\", \"ts\": \"${TEST_TIMESTAMPS[$os]}\", \"install_time\": $install_time_json},"
  done
  results_json="${results_json%,}}"

  local new_content
  new_content=$(/usr/bin/python3 << PYEOF
import re, json, base64, sys

content = """$(echo "$current_content" | sed 's/\\/\\\\/g; s/"""/\\"\\"\\"/g')"""
results = json.loads("""$results_json""")

def update_row(line):
    cells = [c.strip() for c in line.strip().strip('|').split('|')]
    if len(cells) < 5:
        return line

    os_name    = cells[0].strip().lower()
    os_version = cells[1].strip()

    name_map = {
        'ubuntu':     'ubuntu',
        'debian':     'debian',
        'almalinux':  'almalinux',
        'rockylinux': 'rocky',
        'centos':     'centos',
    }
    key_prefix = name_map.get(os_name)
    if not key_prefix:
        return line

    lookup_key = f"{key_prefix}-{os_version}"
    data = results.get(lookup_key)

    if data is None:
        return line

    badge = '✅ Pass' if data['status'] == 'pass' else '❌ Fail'
    cells[2] = data['ts']
    cells[3] = badge
    if data.get('install_time'):
        cells[4] = data['install_time']
    return '| ' + ' | '.join(cells) + ' |'

def replace_section(m):
    lines = m.group(1).split('\n')
    updated = []
    for line in lines:
        stripped = line.strip()
        if stripped.startswith('|') and not stripped.startswith('|---'):
            updated.append(update_row(line))
        else:
            updated.append(line)
    return '<!-- OS_TEST_RESULTS_START -->' + '\n'.join(updated) + '<!-- OS_TEST_RESULTS_END -->'

new = re.sub(
    r'<!-- OS_TEST_RESULTS_START -->(.*?)<!-- OS_TEST_RESULTS_END -->',
    replace_section,
    content,
    flags=re.DOTALL
)
print(new, end='')
PYEOF
)

  if [[ -z "$new_content" ]]; then
    log "ERROR: Python failed to process README content"
    return 1
  fi

  local new_content_b64
  new_content_b64=$(echo "$new_content" | base64 -w 0)

  local commit_message="ci: update OS test results [$(date '+%Y-%m-%d %H:%M UTC')]"

  local put_response
  put_response=$(/usr/bin/curl -s -X PUT \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "https://api.github.com/repos/$GITHUB_REPO/contents/README.md" \
    -d "{
      \"message\": \"$commit_message\",
      \"content\": \"$new_content_b64\",
      \"sha\": \"$sha\"
    }")
  echo "$put_response" >> "$MAIN_LOG"

  if echo "$put_response" | grep -q '"commit"'; then
    log "README updated successfully on GitHub"
  else
    local err
    err=$(echo "$put_response" | /usr/bin/python3 -c "import sys,json; print(json.load(sys.stdin).get('message','unknown'))" 2>>"$MAIN_LOG")
    log "ERROR: Failed to update README: $err"
    return 1
  fi
}

# install_openpanel <ip> <os> [attempt]
# Runs only the installer over SSH (auto-confirming the reboot prompt, and
# retrying once if the box reboots mid-install and drops the SSH session).
# Admin credentials are set at install time via --username/--password so
# Playwright can log in afterwards (Community edition allows a single admin).
install_openpanel() {
  local ip="$1"
  local os="$2"
  local attempt="${3:-1}"
  local max_attempts=2

  os_log "$os" "Starting OpenPanel installation on $ip (attempt $attempt/$max_attempts)..."

  local install_out
  install_out=$(yes | /usr/bin/ssh -i "$SSH_PRIVATE_FILE" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=30 -o LogLevel=ERROR root@"$ip" "bash <(/usr/bin/curl -sSL https://openpanel.org) --skip-firewall --skip-dns-server --skip-requirements --skip-panel-check --username=$TEST_ADMIN_USER --password=$TEST_ADMIN_PASS" 2>&1)
  local exit_code=$?
  {
    echo "----- INSTALL OUTPUT (attempt $attempt) -----"
    echo "$install_out"
    echo "----- INSTALL EXIT CODE: $exit_code -----"
  } >> "$RUN_DIR/${os}.log"
  os_log "$os" "Install command exit code: $exit_code"

  if [[ "$exit_code" -eq 255 ]]; then
    if [[ "$attempt" -ge "$max_attempts" ]]; then
      os_log "$os" "ERROR: SSH lost after reboot but max attempts ($max_attempts) reached"
      send_discord "❌ [$os] Install failed after reboot retry"
      return 1
    fi

    os_log "$os" "SSH connection dropped — likely reboot triggered. Waiting 60s before retrying..."
    send_discord "[$os] Reboot detected during install, waiting for SSH to come back..."
    sleep 60

    wait_for_ssh "$ip" "$os" || return 1

    os_log "$os" "SSH back up — retrying installation (attempt $((attempt + 1))/$max_attempts)..."
    install_openpanel "$ip" "$os" $((attempt + 1))
    return $?
  fi

  if [[ "$exit_code" -ne 0 ]]; then
    os_log "$os" "ERROR: OpenPanel installation failed on $ip (exit $exit_code)"
    send_discord "❌ [$os] OpenPanel installation failed on $ip (exit $exit_code)"
    return 1
  fi

  os_log "$os" "OpenPanel installation finished on $ip"

  local docker_login_out
  docker_login_out=$(/usr/bin/ssh -i "$SSH_PRIVATE_FILE" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=30 -o LogLevel=ERROR root@"$ip" "/usr/bin/docker login --username openpanel --password $DOCKER_HUB_TOKEN" 2>&1)
  local docker_login_rc=$?
  {
    echo "----- DOCKER LOGIN OUTPUT -----"
    echo "$docker_login_out"
    echo "----- DOCKER LOGIN EXIT CODE: $docker_login_rc -----"
  } >> "$RUN_DIR/${os}.log"
  os_log "$os" "docker login exit code: $docker_login_rc"

  return 0
}

# run_openadmin_playwright_tests <ip> <os>
# Waits for OpenAdmin (port 2087) to come up, then runs just the
# 'create user' and 'test autologin' tests from openadmin/tests/aa-users.spec.ts
# against the freshly installed panel.
run_openadmin_playwright_tests() {
  local ip="$1"
  local os="$2"

  local RETRY_INTERVAL=5
  local MAX_WAIT=300
  local elapsed=0
  local status=""
  local ADMIN_URL=""

  while true; do
    status=$(/usr/bin/curl -s -o /dev/null -w "%{http_code}" "http://${ip}:2087/login")
    if [[ "$status" != "000" ]]; then
      ADMIN_URL="http://${ip}:2087"
      os_log "$os" "OpenAdmin available over HTTP: $ADMIN_URL (http $status)"
      break
    fi

    status=$(/usr/bin/curl -k -s -o /dev/null -w "%{http_code}" "https://${ip}:2087/login")
    if [[ "$status" != "000" ]]; then
      ADMIN_URL="https://${ip}:2087"
      os_log "$os" "OpenAdmin available over HTTPS: $ADMIN_URL (http $status)"
      break
    fi

    os_log "$os" "OpenAdmin not ready yet (HTTP/HTTPS both failed, elapsed ${elapsed}s)"
    elapsed=$((elapsed + RETRY_INTERVAL))

    if [ "$elapsed" -ge "$MAX_WAIT" ]; then
      os_log "$os" "ERROR: OpenAdmin did not become responsive after ${MAX_WAIT}s"
      send_discord "❌ [$os] OpenAdmin not reachable on port 2087 after ${MAX_WAIT}s"
      return 1
    fi

    sleep "$RETRY_INTERVAL"
  done

  os_log "$os" "Running Playwright tests: aa-users.spec.ts 'create user' + 'test autologin'..."

  local pw_out
  local pw_rc
  pw_out=$(cd "$SCRIPT_DIR" && BASE_URL="$ADMIN_URL" PANEL_USERNAME="$TEST_ADMIN_USER" PANEL_PASSWORD="$TEST_ADMIN_PASS" \
    npx playwright test -c ../openadmin/playwright.config.ts --project=setup --project=tests ../openadmin/tests/aa-users.spec.ts --grep "create user|test autologin" --reporter=line 2>&1)
  pw_rc=$?
  {
    echo "----- PLAYWRIGHT OUTPUT -----"
    echo "$pw_out"
    echo "----- PLAYWRIGHT EXIT CODE: $pw_rc -----"
  } >> "$RUN_DIR/${os}.log"
  os_log "$os" "Playwright exit code: $pw_rc"

  if [[ $pw_rc -ne 0 ]]; then
    os_log "$os" "ERROR: Playwright openadmin tests failed on $ip"
    send_discord "❌ [$os] Playwright openadmin tests (create user / autologin) failed on $ip"
    return 1
  fi

  os_log "$os" "SUCCESS: openadmin create-user + autologin tests passed on $ip"
  send_discord "✅ [$os] OpenPanel installed — openadmin create-user + autologin tests passed on $ip ($ADMIN_URL)"
}

run_test_cycle() {
  local os="$1"
  os_log "$os" "========================================="
  os_log "$os" "STARTING TEST CYCLE: $os"
  os_log "$os" "========================================="

  TEST_TIMESTAMPS["$os"]=$(date '+%Y-%m-%d %H:%M UTC')

  reinstall_os "$os" || { TEST_RESULTS["$os"]="fail"; update_readme_results; return 1; }
  sleep 60
  wait_for_ssh "$VPS_IP" "$os" || { TEST_RESULTS["$os"]="fail"; update_readme_results; return 1; }

  local install_start install_end
  install_start=$(date +%s)
  install_openpanel "$VPS_IP" "$os" || { TEST_RESULTS["$os"]="fail"; update_readme_results; return 1; }
  install_end=$(date +%s)
  TEST_INSTALL_TIME["$os"]=$(format_duration $((install_end - install_start)))
  os_log "$os" "Install time: ${TEST_INSTALL_TIME[$os]}"
  send_discord "⏱️ [$os] Install time: ${TEST_INSTALL_TIME[$os]}"

  run_openadmin_playwright_tests "$VPS_IP" "$os" || { TEST_RESULTS["$os"]="fail"; update_readme_results; return 1; }

  TEST_RESULTS["$os"]="pass"
  update_readme_results
}

############################################
# Main
############################################
if [[ -n "$1" ]]; then
  if [[ -z "${OS_MAP[$1]}" ]]; then
    echo "ERROR: Unknown OS '$1'. Valid options: ${!OS_MAP[*]}"
    exit 1
  fi
  OS_LIST=("$1")
else
  OS_LIST=("${!OS_MAP[@]}")
fi

SUITE_START_TS=$(date +%s)

log "========================================="
log "VPS OS TEST SUITE STARTED"
log "Testing ${#OS_LIST[@]} operating systems"
log "Logs for this run: $RUN_DIR"
log "========================================="

send_discord "VPS OS test suite started — ${#OS_LIST[@]} OSes (logs: $RUN_DIR)"

FAILED=()
for os in "${OS_LIST[@]}"; do
  if ! run_test_cycle "$os"; then
    FAILED+=("$os")
    log "FAILED: $os added to failure list"
  fi
  log "-----------------------------------------"
done

log "========================================="

TOTAL_DURATION=$(format_duration $(( $(date +%s) - SUITE_START_TS )))

log "Install times per OS:"
for os in "${OS_LIST[@]}"; do
  if [[ -n "${TEST_INSTALL_TIME[$os]:-}" ]]; then
    log "  [$os] ${TEST_INSTALL_TIME[$os]}"
  fi
done
log "Total suite runtime: $TOTAL_DURATION"

if [[ ${#FAILED[@]} -gt 0 ]]; then
  log "COMPLETED WITH FAILURES: ${FAILED[*]}"
  send_discord "Test suite done in $TOTAL_DURATION — failures: ${FAILED[*]} (logs: $RUN_DIR)"
else
  log "ALL TESTS COMPLETED SUCCESSFULLY"
  send_discord "All OS tests passed in $TOTAL_DURATION (logs: $RUN_DIR)"
fi
log "========================================="

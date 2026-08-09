#!/bin/bash
################################################################################
# podman_install_test.sh
#
# Reinstalls a dedicated test VPS with a chosen OS (via the Virtualizor API),
# installs OpenPanel from THIS checkout's PODMAN_INSTALL.sh, creates a test
# user, waits for the user panel (port 2083) to come up, and verifies that
# opencli's autologin link actually reaches /dashboard.
#
# Runs on a self-hosted GitHub Actions runner that lives on the same
# controller box already used for manual OS testing (only that box holds the
# Virtualizor API credentials and the single reusable test VPS, so only one
# OS can be reinstalled/tested at a time). See:
#   .github/workflows/podman-install-test.yml
#
# Usage: podman_install_test.sh <os-key|all>
################################################################################
set -uo pipefail

GREEN='\033[0;32m'; YELLOW='\033[0;33m'; RED='\033[0;31m'; RESET='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
INSTALL_SCRIPT="$REPO_ROOT/PODMAN_INSTALL.sh"

for var in VIRTUALIZOR_API API_KEY API_HASH VPSID VPS_IP TEST_PASS SSH_PRIVATE_KEY TEST_ADMIN_PASS; do
  [[ -n "${!var:-}" ]] || { echo "ERROR: required env var $var is not set"; exit 1; }
done
[[ -f "$INSTALL_SCRIPT" ]] || { echo "ERROR: $INSTALL_SCRIPT not found"; exit 1; }

TEST_ADMIN_USER="${TEST_ADMIN_USER:-admin}"
TEST_USER="testinguser"
TEST_USER_PASS="testingpassword"
TEST_USER_EMAIL="testinguser@example.com"
TEST_USER_PLAN="Standard plan"
USER_PORT=2083

SSH_PRIVATE_FILE="$(mktemp)"
printf '%s\n' "$SSH_PRIVATE_KEY" > "$SSH_PRIVATE_FILE"
chmod 600 "$SSH_PRIVATE_FILE"
cleanup() { rm -f "$SSH_PRIVATE_FILE"; }
trap cleanup EXIT

SSH_OPTS=(-i "$SSH_PRIVATE_FILE" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=30 -o LogLevel=ERROR)
ssh_run() { ssh "${SSH_OPTS[@]}" "root@$VPS_IP" "$@"; }
scp_to()  { scp "${SSH_OPTS[@]}" "$1" "root@$VPS_IP:$2"; }

git config --global --add safe.directory "$REPO_ROOT" 2>/dev/null || true
git -C "$REPO_ROOT" config user.name "github-actions[bot]"
git -C "$REPO_ROOT" config user.email "github-actions[bot]@users.noreply.github.com"

# Virtualizor template IDs for the target test VPS. Keep in sync with the
# equivalent map in openpanel-tests/opencli/os_install.sh -- these IDs are
# specific to that Virtualizor account's OS templates.
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

declare -A RESULTS
declare -A FAIL_REASON
declare -A TEST_TIMESTAMPS
declare -A TEST_INSTALL_TIME

log()    { echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }
os_log() { local os="$1"; shift; log "[$os] $*"; }

send_discord() {
  [[ -n "${DISCORD_WEBHOOK:-}" ]] || return 0
  curl -s -o /dev/null -H "Content-Type: application/json" -X POST \
    -d "{\"content\": \"$1\"}" "$DISCORD_WEBHOOK"
}

format_duration() {
  local total_seconds="$1" h m s
  h=$((total_seconds / 3600)); m=$(((total_seconds % 3600) / 60)); s=$((total_seconds % 60))
  if [[ $h -gt 0 ]]; then printf '%dh%dm%ds' "$h" "$m" "$s"
  elif [[ $m -gt 0 ]]; then printf '%dm%ds' "$m" "$s"
  else printf '%ds' "$s"; fi
}

# update_readme -- rewrites the OS_TEST_RESULTS table in README.md in the
# local checkout using whatever's in RESULTS/TEST_TIMESTAMPS/TEST_INSTALL_TIME
# so far, then commits and pushes it. Safe to call repeatedly; it's a no-op
# if nothing actually changed (e.g. same result as last recorded run).
update_readme() {
  local os="$1"
  local results_json="{"
  local k
  for k in "${!RESULTS[@]}"; do
    local install_time_json="null"
    if [[ "${RESULTS[$k]}" == "pass" && -n "${TEST_INSTALL_TIME[$k]:-}" ]]; then
      install_time_json="\"${TEST_INSTALL_TIME[$k]}\""
    fi
    results_json+="\"$k\": {\"status\": \"${RESULTS[$k]}\", \"ts\": \"${TEST_TIMESTAMPS[$k]}\", \"install_time\": $install_time_json},"
  done
  results_json="${results_json%,}}"

  RESULTS_JSON="$results_json" python3 - "$REPO_ROOT/README.md" <<'PYEOF'
import sys, re, json, os

readme_path = sys.argv[1]
results = json.loads(os.environ["RESULTS_JSON"])

with open(readme_path) as f:
    content = f.read()

name_map = {
    'ubuntu':     'ubuntu',
    'debian':     'debian',
    'almalinux':  'almalinux',
    'rockylinux': 'rocky',
    'centos':     'centos',
}

def update_row(line):
    cells = [c.strip() for c in line.strip().strip('|').split('|')]
    if len(cells) < 5:
        return line
    os_name = cells[0].strip().lower()
    os_version = cells[1].strip()
    key_prefix = name_map.get(os_name)
    if not key_prefix:
        return line
    data = results.get(f"{key_prefix}-{os_version}")
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
    updated = [
        update_row(line) if line.strip().startswith('|') and not line.strip().startswith('|---') else line
        for line in lines
    ]
    return '<!-- OS_TEST_RESULTS_START -->' + '\n'.join(updated) + '<!-- OS_TEST_RESULTS_END -->'

new = re.sub(
    r'<!-- OS_TEST_RESULTS_START -->(.*?)<!-- OS_TEST_RESULTS_END -->',
    replace_section,
    content,
    flags=re.DOTALL,
)

with open(readme_path, 'w') as f:
    f.write(new)
PYEOF

  ( cd "$REPO_ROOT" && git add README.md )
  if ( cd "$REPO_ROOT" && git diff --cached --quiet ); then
    os_log "$os" "README already up to date, nothing to commit"
    return 0
  fi

  ( cd "$REPO_ROOT" \
    && git commit -q -m "ci: update OS test results [$os] [$(date -u '+%Y-%m-%d %H:%M UTC')]" \
    && git pull --rebase --quiet origin main \
    && git push --quiet origin HEAD:main
  ) && os_log "$os" "README updated and pushed" \
    || os_log "$os" "ERROR: failed to commit/push README update"
}

reinstall_os() {
  local os="$1" os_id="${OS_MAP[$1]}"
  os_log "$os" "Reinstalling VPS $VPSID with OS: $os (template $os_id)"
  local resp
  resp=$(curl -sk -X POST -d "reinsos=1&newos=$os_id&newpass=$TEST_PASS&conf=$TEST_PASS" -L \
    "$VIRTUALIZOR_API/index.php?act=ostemplate&svs=$VPSID&api=json&apikey=$API_KEY&apipass=$API_HASH")
  if [[ $? -ne 0 ]]; then
    os_log "$os" "ERROR: reinstall API call failed"
    return 1
  fi
  os_log "$os" "Reinstall triggered, response: $resp"
}

wait_for_ssh() {
  local os="$1"
  os_log "$os" "Waiting for SSH on $VPS_IP..."
  for i in {1..60}; do
    nc -z -w3 "$VPS_IP" 22 2>/dev/null && { os_log "$os" "SSH is up (attempt $i)"; return 0; }
    sleep 10
  done
  os_log "$os" "ERROR: SSH timeout after 10 minutes"
  return 1
}

# install_podman <os> [attempt] -- copies THIS checkout's PODMAN_INSTALL.sh to
# the VPS and runs it, auto-confirming prompts and retrying once if the VPS
# reboots mid-install and drops the SSH session (exit 255).
install_podman() {
  local os="$1" attempt="${2:-1}" max_attempts=2

  os_log "$os" "Copying PODMAN_INSTALL.sh to VPS (attempt $attempt/$max_attempts)..."
  if ! scp_to "$INSTALL_SCRIPT" "/root/PODMAN_INSTALL.sh"; then
    os_log "$os" "ERROR: scp of PODMAN_INSTALL.sh failed"
    return 1
  fi

  os_log "$os" "Running PODMAN_INSTALL.sh on $VPS_IP..."
  yes | ssh_run "bash /root/PODMAN_INSTALL.sh --skip-firewall --skip-dns-server --skip-requirements --skip-panel-check --selfsigned --username=$TEST_ADMIN_USER --password=$TEST_ADMIN_PASS"
  local exit_code=$?
  os_log "$os" "Install exit code: $exit_code"

  if [[ "$exit_code" -eq 255 ]]; then
    if [[ "$attempt" -ge "$max_attempts" ]]; then
      os_log "$os" "ERROR: SSH lost after reboot but max attempts reached"
      return 1
    fi
    os_log "$os" "SSH dropped -- likely reboot. Waiting 60s and retrying..."
    sleep 60
    wait_for_ssh "$os" || return 1
    install_podman "$os" $((attempt + 1))
    return $?
  fi

  [[ "$exit_code" -eq 0 ]]
}

wait_for_panel() {
  local os="$1" elapsed=0 interval=5 max_wait=300 status
  os_log "$os" "Waiting for user panel on port $USER_PORT..."
  while true; do
    status=$(curl -sk -o /dev/null -w "%{http_code}" --max-time 5 "http://$VPS_IP:$USER_PORT/login" 2>/dev/null)
    [[ "$status" != "000" ]] && { os_log "$os" "Panel responded (http $status)"; return 0; }
    status=$(curl -sk -o /dev/null -w "%{http_code}" --max-time 5 "https://$VPS_IP:$USER_PORT/login" 2>/dev/null)
    [[ "$status" != "000" ]] && { os_log "$os" "Panel responded over https (http $status)"; return 0; }

    elapsed=$((elapsed + interval))
    [[ "$elapsed" -ge "$max_wait" ]] && { os_log "$os" "ERROR: panel not up after ${max_wait}s"; return 1; }
    sleep "$interval"
  done
}

create_test_user() {
  local os="$1" out
  os_log "$os" "Creating test user via opencli user-add..."
  out=$(ssh_run "opencli user-add $TEST_USER $TEST_USER_PASS $TEST_USER_EMAIL '$TEST_USER_PLAN'" 2>&1)
  local rc=$?
  os_log "$os" "user-add output: $out"
  [[ $rc -eq 0 ]]
}

test_autologin() {
  local os="$1" login_url headers status location

  login_url=$(ssh_run "opencli user-login $TEST_USER" 2>/dev/null | tail -n1 | tr -d '\r')
  if [[ -z "$login_url" || "$login_url" != http* ]]; then
    os_log "$os" "ERROR: opencli user-login did not return a URL (got: '$login_url')"
    return 1
  fi
  os_log "$os" "Autologin URL: $login_url"

  headers=$(curl -sk -o /dev/null -D - --max-time 15 "$login_url")
  status=$(echo "$headers" | head -n1 | tr -d '\r')
  location=$(echo "$headers" | grep -i '^location:' | tr -d '\r')
  os_log "$os" "Autologin response: $status / $location"

  if [[ "$status" == *302* ]] && [[ "$location" == *"/dashboard"* ]]; then
    os_log "$os" "SUCCESS: autologin redirected to /dashboard"
    return 0
  fi

  os_log "$os" "ERROR: autologin did not redirect to /dashboard"
  return 1
}

run_test_cycle() {
  local os="$1"
  os_log "$os" "========== STARTING TEST CYCLE: $os =========="
  TEST_TIMESTAMPS[$os]=$(date -u '+%Y-%m-%d %H:%M UTC')

  fail() { FAIL_REASON[$os]="$1"; RESULTS[$os]="fail"; update_readme "$os"; return 1; }

  reinstall_os "$os" || { fail "reinstall failed"; return 1; }
  sleep 60
  wait_for_ssh "$os" || { fail "ssh timeout"; return 1; }

  local install_start install_end
  install_start=$(date +%s)
  install_podman "$os" || { fail "install failed"; return 1; }
  install_end=$(date +%s)
  TEST_INSTALL_TIME[$os]=$(format_duration $((install_end - install_start)))
  os_log "$os" "Install time: ${TEST_INSTALL_TIME[$os]}"

  wait_for_panel "$os" || { fail "panel did not come up"; return 1; }
  create_test_user "$os" || { fail "user-add failed"; return 1; }
  test_autologin "$os" || { fail "autologin did not reach /dashboard"; return 1; }

  RESULTS[$os]="pass"
  update_readme "$os"
  send_discord "✅ [$os] PODMAN_INSTALL.sh: install + user-add + autologin all passed on $VPS_IP (install time: ${TEST_INSTALL_TIME[$os]})"
}

############################################
# Main
############################################
REQUESTED="${1:-all}"

if [[ "$REQUESTED" == "all" ]]; then
  OS_LIST=("${!OS_MAP[@]}")
else
  [[ -n "${OS_MAP[$REQUESTED]:-}" ]] || { echo "ERROR: Unknown OS '$REQUESTED'. Valid: all ${!OS_MAP[*]}"; exit 1; }
  OS_LIST=("$REQUESTED")
fi

log "========================================="
log "PODMAN_INSTALL.sh test suite started -- ${#OS_LIST[@]} OS(es): ${OS_LIST[*]}"
log "========================================="

FAILED=()
for os in "${OS_LIST[@]}"; do
  if ! run_test_cycle "$os"; then
    FAILED+=("$os")
    send_discord "❌ [$os] PODMAN_INSTALL.sh test failed: ${FAIL_REASON[$os]:-unknown}"
  fi
  log "-----------------------------------------"
done

log "========================================="
log "RESULTS:"
for os in "${OS_LIST[@]}"; do
  if [[ "${RESULTS[$os]:-}" == "pass" ]]; then
    log "  [$os] PASS"
  else
    log "  [$os] FAIL -- ${FAIL_REASON[$os]:-unknown}"
  fi
done
log "========================================="

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### PODMAN_INSTALL.sh test results"
    echo
    echo "| OS | Result | Notes |"
    echo "|---|---|---|"
    for os in "${OS_LIST[@]}"; do
      if [[ "${RESULTS[$os]:-}" == "pass" ]]; then
        echo "| $os | ✅ Pass | |"
      else
        echo "| $os | ❌ Fail | ${FAIL_REASON[$os]:-unknown} |"
      fi
    done
  } >> "$GITHUB_STEP_SUMMARY"
fi

if [[ ${#FAILED[@]} -gt 0 ]]; then
  log "COMPLETED WITH FAILURES: ${FAILED[*]}"
  exit 1
fi

log "ALL TESTS PASSED"
exit 0

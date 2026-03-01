#!/bin/bash

readonly CONF_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
readonly LOCK_FILE="/tmp/swap_cleanup.lock"
readonly INI_FILE="/etc/openpanel/openadmin/config/notifications.ini"
readonly LOG_FILE="/var/log/openpanel/admin/notifications.log"
readonly DNS_STAMP="/tmp/sentinel_dns_last_run"
readonly DISPLAY_TIME=$(date +"%Y-%m-%d %H:%M:%S")
HOSTNAME=$(hostname)

[ ! -f "$INI_FILE" ] && { echo "Error: INI file not found: $INI_FILE"; exit 1; }

mkdir -p "$(dirname "$LOG_FILE")"
[[ ! -f "$LOG_FILE" ]] && > "$LOG_FILE"

STATUS=0 PASS=0 WARN=0 FAIL=0

conf_get() { awk -F= "/^${1}=/{print \$2; exit}" "$CONF_FILE" 2>/dev/null; }
ini_get()  { awk -F= "/^${1}=/{print \$2; exit}" "$INI_FILE"  2>/dev/null; }

validate_yes_no() { [[ "$1" == "yes" || "$1" == "no" ]] && echo "$1" || echo "yes"; }
validate_number() { [[ "$1" =~ ^([1-9][0-9]?|100)$ ]] && echo "$1" || echo "$2"; }

EMAIL=$(conf_get email)
EMAIL_ALERT=$([[ -n "$EMAIL" ]] && echo "yes" || echo "no")

REBOOT=$(validate_yes_no "$(ini_get reboot)")
MAIN_DOMAIN_AND_NS=$(validate_yes_no "$(ini_get dns)")
LOGIN=$(validate_yes_no "$(ini_get login)")
SSH_LOGIN=$(validate_yes_no "$(ini_get ssh)")
SERVICES=$(ini_get services); SERVICES="${SERVICES:-admin,docker,mysql,csf,panel}"

LOAD_THRESHOLD=$(validate_number "$(ini_get load)" 20)
CPU_THRESHOLD=$(validate_number  "$(ini_get cpu)"  90)
RAM_THRESHOLD=$(validate_number  "$(ini_get ram)"  85)
DISK_THRESHOLD=$(validate_number "$(ini_get du)"   85)
SWAP_THRESHOLD=$(validate_number "$(ini_get swap)" 40)

is_unread_message_present() { grep -qF "UNREAD $1" "$LOG_FILE"; }

email_notification() {
  local title=$1 message=$2
  local token; token=$(tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 64)
  awk -v t="$token" '/^mail_security_token=/{$0="mail_security_token="t} 1' \
    "$CONF_FILE" > "${CONF_FILE}.tmp" && mv "${CONF_FILE}.tmp" "$CONF_FILE"

  local domain; domain=$(opencli domain)
  local proto="http"
  [[ "$domain" =~ ^[a-zA-Z0-9.-]+$ ]] && proto="https"

  local auth_opt=""
  if awk -F= '/^basic_auth=/{exit ($2=="yes")?0:1}' \
      /etc/openpanel/openadmin/config/admin.ini 2>/dev/null; then
    local u p
    u=$(awk -F= '/^basic_auth_username=/{print $2; exit}' /etc/openpanel/openadmin/config/admin.ini)
    p=$(awk -F= '/^basic_auth_password=/{print $2; exit}' /etc/openpanel/openadmin/config/admin.ini)
    auth_opt="--user ${u}:${p}"
  fi

  local resp
  resp=$(curl -4 --max-time 5 -ksf -X POST "$proto://$domain:2087/send_email" \
    $auth_opt \
    -F "transient=$token" -F "recipient=$EMAIL" \
    -F "subject=$title"   -F "body=$message" 2>/dev/null)

  case "$resp" in
    *'"error"'*)             echo "Error sending email: $resp" ;;
    *'"sent successfully"'*) echo "Email sent." ;;
  esac
}

write_notification() {
  local title=$1 message=$2
  is_unread_message_present "$title" && return
  echo "$DISPLAY_TIME UNREAD $title MESSAGE: $message" >> "$LOG_FILE"
  [[ "$EMAIL_ALERT" == "yes" ]] && email_notification "$title" "$message"
}

ensure_installed() {
  command -v "$1" &>/dev/null && return
  local pkg="${2:-$1}"
  if   command -v apt-get &>/dev/null; then apt-get install -y -qq "$pkg" &>/dev/null
  elif command -v yum     &>/dev/null; then yum install -y -q  "$pkg" &>/dev/null
  elif command -v dnf     &>/dev/null; then dnf install -y -q  "$pkg" &>/dev/null
  else echo "Error: cannot install $pkg"; exit 1; fi
  command -v "$1" &>/dev/null || { echo "Error: $1 install failed"; exit 1; }
}

ip_in_cidr() {
  local ip=$1 cidr=$2 network mask
  local i1 i2 i3 i4 n1 n2 n3 n4
  IFS=/ read -r network mask <<< "$cidr"
  IFS=. read -r i1 i2 i3 i4 <<< "$ip"
  IFS=. read -r n1 n2 n3 n4 <<< "$network"
  local ipbin=$(( (i1<<24)|(i2<<16)|(i3<<8)|i4 ))
  local netbin=$(( (n1<<24)|(n2<<16)|(n3<<8)|n4 ))
  local maskbin=$(( (0xFFFFFFFF << (32-mask)) & 0xFFFFFFFF ))
  (( (ipbin & maskbin) == (netbin & maskbin) ))
}

perform_startup_action() {
  if [[ "$REBOOT" == "no" ]]; then
    ((WARN++)); echo "[!] Reboot check disabled."; return
  fi
  write_notification "SYSTEM REBOOT!" "System was rebooted. $(uptime)"
}

email_daily_report() {
  if [[ "$EMAIL_ALERT" == "no" ]]; then
    echo "Email alerts disabled."; return
  fi
  email_notification "Daily Usage Report" "Daily Usage Report"
}

check_service_status() {
  local svc=$1 title=$2
  if systemctl is-active --quiet "$svc"; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m $svc is active."
  else
    if [[ "$svc" == "admin" && ! -f /root/openadmin_is_disabled ]]; then
      ((PASS++)); echo -e "\e[32m[✔]\e[0m $svc disabled by Administrator."; return
    fi
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m $svc is not active."
    local log; log=$(journalctl -n 5 -u "$svc" 2>/dev/null | awk '{printf "%s|",$0}')
    [[ -n "$log" ]] && write_notification "$title" "$log"
    systemctl restart "$svc"
    systemctl is-active --quiet "$svc" \
      && echo -e "\e[32m[✔]\e[0m $svc restarted." \
      || echo -e "\e[31m[✘]\e[0m Failed to restart $svc."
  fi
}

_docker_ps()  { docker --context=default ps --format "{{.Names}}"; }
_docker_log() { docker --context=default logs --tail 10 "$1" 2>&1 \
                  | awk '{gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s\\n",$0}'; }

_caddy_http_ok() {
  local code
  code=$(curl -so /dev/null -w "%{http_code}" --connect-timeout 1 --max-time 1 "http://localhost/check")
  [[ "$code" == "200" || "$code" == "404" ]]
}

_docker_check_after_restart() {
  local svc=$1 title=$2
  if _docker_ps | grep -wq "$svc"; then
    ((WARN--)); echo -e "\e[32m[✔]\e[0m $svc restarted successfully."
  else
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m $svc failed to restart."
    write_notification "$title" "$(_docker_log "$svc")"
  fi
}

docker_containers_status() {
  local svc=$1 title=$2

  if _docker_ps | grep -wq "$svc"; then
    if [[ "$svc" == "caddy" ]]; then
      if _caddy_http_ok; then
        ((PASS++)); echo -e "\e[32m[✔]\e[0m caddy is active and responding."
      else
        ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m caddy running but unresponsive — restarting."
        cd /root && docker --context=default compose up -d caddy &>/dev/null
        sleep 2
        if _caddy_http_ok; then
          ((PASS++)); ((WARN--)); echo -e "\e[32m[✔]\e[0m caddy recovered."
        else
          ((WARN--)); ((FAIL++)); STATUS=2
          echo -e "\e[31m[✘]\e[0m caddy still unresponsive after restart."
          write_notification "$title" "$(_docker_log caddy)"
        fi
      fi
    else
      ((PASS++)); echo -e "\e[32m[✔]\e[0m $svc docker container is active."
    fi
    return
  fi

  ((WARN++))
  case "$svc" in
    openpanel)
      local users; users=$(opencli user-list --json 2>/dev/null \
        | awk -F'"' '/username/{print $4}' | grep -v SUSPENDED)
      if [[ -z "$users" || "$users" == "No users." ]]; then
        ((WARN--)); echo "  - No users found; $svc not needed."
      else
        cd /root && docker --context=default compose up -d openpanel &>/dev/null
        _docker_check_after_restart "$svc" "$title"
      fi ;;
    openpanel_dns)
      if ls /etc/bind/zones/*.zone &>/dev/null; then
        cd /root && docker --context=default compose up -d bind9 &>/dev/null
        _docker_check_after_restart "$svc" "$title"
      else
        ((WARN--)); echo "  - No DNS zones; bind9 not needed."
      fi ;;
    caddy)
      if ls /etc/openpanel/caddy/domains &>/dev/null; then
        cd /root && docker --context=default compose up -d caddy &>/dev/null
        _docker_check_after_restart "$svc" "$title"
      else
        ((WARN--)); echo "  - No domains; caddy not needed."
      fi ;;
    *)
      docker --context=default restart "$svc" &>/dev/null
      _docker_check_after_restart "$svc" "$title" ;;
  esac
}

mysql_docker_containers_status() {
  local title="MySQL service not active!"
  if _docker_ps | grep -q "openpanel_mysql"; then
    if mysql -Ne "SELECT 'PONG' AS PING;" 2>/dev/null | grep -q "PONG"; then
      ((PASS++)); echo -e "\e[32m[✔]\e[0m MySQL container active and responding."
    else
      echo -e "\e[31m[✘]\e[0m MySQL running but not responding — restarting."
      cd /root && docker --context=default compose up -d openpanel_mysql &>/dev/null
    fi
  else
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m MySQL container not running — restarting."
    cd /root && docker --context=default compose up -d openpanel_mysql &>/dev/null
    sleep 5
    if mysql -Ne "SELECT 'PONG' AS PING;" 2>/dev/null | grep -q "PONG"; then
      ((FAIL--)); STATUS=1
      echo "    MySQL is back online."
      write_notification "MySQL restarted successfully!" \
        "Sentinel restarted MySQL and it is responding now."
    else
      echo "    Error: MySQL still not responding!"
      write_notification "$title" "MySQL did not respond after restart. Please check ASAP."
    fi
  fi
}

check_services() {
  local svc
  for svc in caddy csf admin docker panel mysql named; do
    [[ ",$SERVICES," != *",$svc,"* ]] && continue
    case "$svc" in
      caddy)  docker_containers_status  'caddy'         'Caddy not active — websites down!'             ;;
      csf)    check_service_status      'csf'           'CSF Firewall not active — server unprotected!' ;;
      admin)  check_service_status      'admin'         'OpenAdmin service not accessible!'             ;;
      docker) check_service_status      'docker'        'Docker not active — user websites down!'       ;;
      panel)  docker_containers_status  'openpanel'     'OpenPanel container not running!'              ;;
      mysql)  mysql_docker_containers_status                                                            ;;
      named)  docker_containers_status  'openpanel_dns' 'BIND9 not active — DNS broken!'               ;;
    esac
  done
}

check_new_logins() {
  if [[ "$LOGIN" == "no" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Login check disabled."; return
  fi

  local login_log="/var/log/openpanel/admin/login.log"
  [[ ! -f "$login_log" ]] && > "$login_log"

  local last_login; last_login=$(tail -n 1 "$login_log" 2>/dev/null)
  if [[ -z "$last_login" ]]; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No logins to OpenAdmin detected."; return
  fi

  local username ip_address
  read -r _ _ username ip_address _ <<< "$last_login"

  if [[ ! "$ip_address" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ || "$ip_address" == "127.0.0.1" ]]; then
    echo "Invalid/loopback IP: $ip_address"; return 1
  fi

  local count; count=$(grep -c "$username" "$login_log")
  if (( count == 1 )); then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m First login: $username from $ip_address."
  elif ! head -n -1 "$login_log" | grep -qF "$username $ip_address"; then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m $username logged in from new IP: $ip_address"
    write_notification "Admin $username accessed from new IP" \
      "Admin account $username was accessed from new IP: $ip_address"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m $username from known IP: $ip_address"
  fi
}

check_ssh_logins() {
  [[ "$SSH_LOGIN" == "no" ]] && return

  if is_unread_message_present "Suspicious SSH login detected"; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Unread SSH notification exists. Skipping."; return
  fi

  local ssh_ips
  ssh_ips=$(who | awk '/pts/{gsub(/[():]/, "", $5); n=split($5,a,":"); print a[1]}')
  if [[ -z "$ssh_ips" ]]; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No active SSH sessions."; return
  fi

  local login_log="/var/log/openpanel/admin/login.log"
  local login_ips; login_ips=$(awk '{print $NF}' "$login_log" 2>/dev/null)
  if [[ -z "$login_ips" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m SSH user detected; postponing check until OpenAdmin is ready."
    return
  fi

  local wl_file="/etc/openpanel/openadmin/ssh_whitelist.conf"
  local -a wl_ips wl_cidrs
  if [[ -f "$wl_file" ]]; then
    while IFS= read -r e; do
      [[ -z "$e" ]] && continue
      [[ "$e" == */* ]] && wl_cidrs+=("$e") || wl_ips+=("$e")
    done < "$wl_file"
  fi

  local -a suspicious safe
  local ip ok cidr
  for ip in $ssh_ips; do
    [[ "$ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || continue
    ok=0
    [[ " ${wl_ips[*]} " == *" $ip "* ]] && ok=1
    if (( !ok )); then
      for cidr in "${wl_cidrs[@]}"; do ip_in_cidr "$ip" "$cidr" && { ok=1; break; }; done
    fi
    if (( !ok )) && ! grep -qF "$ip" <<< "$login_ips"; then
      suspicious+=("$ip")
    else
      safe+=("$ip")
    fi
  done

  if (( ${#suspicious[@]} > 0 )); then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m Suspicious SSH IPs: ${suspicious[*]}"
    write_notification "Suspicious SSH login detected" "${suspicious[*]}"
  else
    ((PASS++))
    echo -e "\e[32m[✔]\e[0m ${#safe[@]} SSH session(s) from known IPs: ${safe[*]}"
  fi
}

check_disk_usage() {
  local title="Running out of Disk Space!"
  if is_unread_message_present "$title"; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Unread DU notification. Skipping."; return
  fi
  local pct; pct=$(df --output=pcent / | awk 'NR==2{gsub(/%/,"",$1); print $1+0}')
  if (( pct > DISK_THRESHOLD )); then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m Disk ${pct}% > threshold ${DISK_THRESHOLD}%"
    write_notification "$title" "Disk usage: ${pct}% | $(df -h | awk '{printf "%s|",$0}')"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m Disk ${pct}% < threshold ${DISK_THRESHOLD}%"
  fi
}

check_system_load() {
  local title="High System Load!"
  local load_raw; read -r load_raw _ < /proc/loadavg
  local load=${load_raw%%.*}
  if (( load > LOAD_THRESHOLD )); then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m Load ${load} > threshold ${LOAD_THRESHOLD}. Generating crash report."
    generate_crashlog_report
    write_notification "$title" "Load: $load"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m Load ${load} < threshold ${LOAD_THRESHOLD}."
  fi
}

check_ram_usage() {
  local title="High Memory Usage!"
  if is_unread_message_present "Used RAM"; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Unread RAM notification. Skipping."; return
  fi
  local _ total used _rest
  read -r _ total used _rest < <(free -m | awk '/^Mem:/')
  local pct=$(( used * 100 / total ))
  if (( pct > RAM_THRESHOLD )); then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m RAM ${pct}% > threshold ${RAM_THRESHOLD}%"
    write_notification "$title" "Used RAM: ${used}MB / ${total}MB (${pct}%)"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m RAM ${pct}% < threshold ${RAM_THRESHOLD}%"
  fi
}

check_cpu_usage() {
  local title="High CPU Usage!"
  local -a f1 f2
  read -ra f1 < <(grep '^cpu ' /proc/stat)
  sleep 0.2
  read -ra f2 < <(grep '^cpu ' /proc/stat)
  local idle1=${f1[4]} idle2=${f2[4]}
  local total1=0 total2=0 v
  for v in "${f1[@]:1}"; do (( total1 += v )); done
  for v in "${f2[@]:1}"; do (( total2 += v )); done
  local diff_total=$(( total2 - total1 ))
  local pct=$(( diff_total > 0 ? 100*(diff_total - (idle2-idle1)) / diff_total : 0 ))
  if (( pct > CPU_THRESHOLD )); then
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m CPU ${pct}% > threshold ${CPU_THRESHOLD}%"
    local procs; procs=$(ps ax --sort=-%cpu -o pid:7,pcpu:6,comm:20 | head -10 | awk '{printf "%s|",$0}')
    write_notification "$title" "CPU: ${pct}% | Top: $procs"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m CPU ${pct}% < threshold ${CPU_THRESHOLD}%"
  fi
}

check_swap_usage() {
  local title="High SWAP usage!"
  local _ stotal sused _rest
  read -r _ stotal sused _rest < <(free | awk '/^Swap:/')
  if (( stotal == 0 )); then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No SWAP configured."; return
  fi
  local pct=$(( sused * 100 / stotal ))

  if (( pct <= SWAP_THRESHOLD )); then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m SWAP ${pct}% < threshold ${SWAP_THRESHOLD}%"
    rm -f "$LOCK_FILE"; return
  fi

  if [[ -f "$LOCK_FILE" ]]; then
    local age=$(( $(date +%s) - $(date -r "$LOCK_FILE" +%s) ))
    if (( age <= 21600 )); then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m SWAP cleanup already in progress. Skipping."; return
    fi
    rm -f "$LOCK_FILE"
  fi

  echo -e "\e[31m[✘]\e[0m SWAP ${pct}% > threshold ${SWAP_THRESHOLD}%. Clearing..."
  write_notification "$title" "SWAP: ${pct}%. Cleanup starting."
  touch "$LOCK_FILE"

  echo 3 > /proc/sys/vm/drop_caches
  swapoff -a && swapon -a

  local stotal2 sused2
  read -r _ stotal2 sused2 _rest < <(free | awk '/^Swap:/')
  local pct2=$(( stotal2 > 0 ? sused2*100/stotal2 : 0 ))
  if (( pct2 < SWAP_THRESHOLD )); then
    rm -f "$LOCK_FILE"
    write_notification "SWAP cleared — now ${pct2}%" "Sentinel cleared SWAP on $HOSTNAME."
    echo -e "\e[32m[✔]\e[0m SWAP cleared successfully. Now: ${pct2}%"
  else
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m SWAP still high after cleanup: ${pct2}%"
    write_notification "URGENT: SWAP not cleared on $HOSTNAME" \
      "SWAP cleanup attempted at $DISPLAY_TIME but usage still ${pct2}%."
  fi
}

generate_crashlog_report() {
  local dir="/var/log/openpanel/admin/crashlog"
  mkdir -p "$dir"
  local report="$dir/$(date +%s).txt"
  {
    echo "=== GENERAL === Hostname: $HOSTNAME | Date: $DISPLAY_TIME"
    echo "=== LOAD ===";     cat /proc/loadavg
    echo "=== TOP (MEM) ==="; ps -eo pid:7,%mem:5,comm:20 --sort=-%mem | head -11
    echo "=== TOP (CPU) ==="; ps -eo pid:7,%cpu:5,comm:20 --sort=-%cpu | head -11
    echo "=== SWAP TOP ===";  grep VmSwap /proc/*/status 2>/dev/null | sort -k2 -hr | head -10
    echo "=== DISKSTATS ==="; head -10 /proc/diskstats
  } > "$report"
}

check_if_panel_domain_and_ns_resolve_to_server() {
  if [[ "$MAIN_DOMAIN_AND_NS" == "no" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m DNS check disabled."; return
  fi

  if [[ -f "$DNS_STAMP" ]]; then
    local age=$(( $(date +%s) - $(date -r "$DNS_STAMP" +%s) ))
    if (( age < 3600 )); then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m DNS check skipped (last run $((age/60))m ago)."; return
    fi
  fi
  touch "$DNS_STAMP"

  local FORCED_DOMAIN; FORCED_DOMAIN=$(opencli domain 2>/dev/null)
  local CHECK_DOMAIN="no" CHECK_NS="no"
  [[ "$FORCED_DOMAIN" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]] && CHECK_DOMAIN="yes"

  local NS1; NS1=$(conf_get ns1)
  if [[ "$NS1" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    CHECK_NS="yes"
    local NS2 NS3 NS4
    NS2=$(conf_get ns2); NS3=$(conf_get ns3); NS4=$(conf_get ns4)
  fi

  if [[ "$CHECK_DOMAIN" == "no" && "$CHECK_NS" == "no" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m No valid domain/NS configured. Skipping DNS check."
    return
  fi

  local GNS="8.8.8.8"
  local SERVER_IP; SERVER_IP=$(curl --silent --max-time 2 -4 "https://ip.openpanel.com")
  [[ -z "$SERVER_IP" ]] && \
    SERVER_IP=$(ip -4 addr show scope global | awk '/inet /{split($2,a,"/"); print a[1]; exit}')

  if [[ "$CHECK_DOMAIN" == "yes" ]]; then
    ensure_installed dig bind-utils
    local domain_ip; domain_ip=$(dig +short @"$GNS" "$FORCED_DOMAIN" 2>/dev/null)
    if [[ "$domain_ip" == "$SERVER_IP" ]]; then
      ((PASS++)); echo -e "\e[32m[✔]\e[0m $FORCED_DOMAIN → $SERVER_IP"
    else
      local ns_rec; ns_rec=$(dig +short @"$GNS" NS "$FORCED_DOMAIN" 2>/dev/null)
      if [[ "${ns_rec,,}" == *cloudflare* ]]; then
        ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m $FORCED_DOMAIN uses Cloudflare proxy — skipping IP check."
      else
        ((FAIL++)); STATUS=2
        echo -e "\e[31m[✘]\e[0m $FORCED_DOMAIN resolves to $domain_ip, expected $SERVER_IP"
        write_notification "$FORCED_DOMAIN does not resolve to $SERVER_IP" \
          "$FORCED_DOMAIN should point to $SERVER_IP."
      fi
    fi
  else
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Domain not set; skipping domain DNS check."
  fi

  if [[ "$CHECK_NS" == "yes" ]]; then
    if [[ -z "$NS2" ]]; then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Only one NS set — add at least two for redundancy."
      return
    fi
    local all_ips; all_ips=$(hostname -I)
    local -a failed
    local ns_var ns_host ns_ip
    for ns_var in NS1 NS2 NS3 NS4; do
      ns_host="${!ns_var}"
      [[ -z "$ns_host" ]] && continue
      ns_ip=$(dig +short @"$GNS" "$ns_host" 2>/dev/null)
      grep -qw "$ns_ip" <<< "$all_ips" || \
        failed+=("$ns_var ($ns_host) → $ns_ip (expected: $all_ips)")
    done
    if (( ${#failed[@]} == 0 )); then
      ((PASS++)); echo -e "\e[32m[✔]\e[0m All nameservers resolve to local IPs."
    else
      ((FAIL++)); STATUS=2
      printf '    %s\n' "${failed[@]}"
      local IFS='|'; write_notification "Nameservers do not resolve to local IPs" "${failed[*]}"
    fi
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No nameservers configured; skipping NS check."
  fi
}

hr() { printf '%*s\n' "${COLUMNS:-80}" '' | tr ' ' '-'; }

summary() {
  hr
  case $STATUS in
    0) echo -e "\e[32mAll Tests Passed!\e[0m" ;;
    1) echo -e "\e[93mSome non-critical tests failed.\e[0m" ;;
    *) echo -e "\e[41mOne or more tests failed.\e[0m" ;;
  esac
  hr
  echo -e "\e[1m${PASS} PASS  ${WARN} WARN  ${FAIL} FAIL\e[0m"
  hr
}

case "$1" in
  --startup) perform_startup_action; exit 0 ;;
  --report)  email_daily_report;     exit 0 ;;
esac

exec 200>/root/sentinel_run.lock
flock -n 200 || { echo "Error: Another instance is already running."; exit 1; }

hr
echo "  Sentinel - OpenPanel server health monitor"
hr
echo "Checking services:"
check_services
hr
echo "Checking logins:"
check_new_logins
check_ssh_logins
hr
echo "Checking resources:"
check_disk_usage
check_system_load
check_ram_usage
check_cpu_usage
check_swap_usage
hr
echo "Checking DNS:"
check_if_panel_domain_and_ns_resolve_to_server
summary

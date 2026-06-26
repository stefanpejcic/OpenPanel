#!/bin/bash
################################################################################
# Script Name: sentinel.sh
# Description: Check system services, traffic and resource usage.
# Usage: opencli sentinel
# Author: Stefan Pejcic
# Created: 01.11.2023
# Last Modified: 25.06.2026
# Company: openpanel.com
# Copyright (c) Stefan Pejcic <stefan@pejcic.rs>
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
################################################################################

# config
readonly CONF_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
readonly INI_FILE="/etc/openpanel/openadmin/config/notifications.ini"
readonly LOG_FILE="/var/log/openpanel/admin/notifications.log"

readonly DISPLAY_TIME=$(date +"%Y-%m-%d %H:%M:%S")
HOSTNAME=$(hostname)

# lock files
readonly SNAPSHOT_FILE="/var/log/openpanel/admin/sentinel_snapshots.jsonl"
readonly SNAPSHOT_MAX_LINES=8640  # 30d
readonly LOCK_FILE_FOR_DNS_CHECK="/tmp/sentinel.dns"
readonly LOCK_FILE_FOR_OOM_CHECK="/tmp/sentinel.oom"
readonly LOCK_FILE_FOR_DOCKER_PRUNE="/tmp/sentinel.docker"

[ ! -f "$INI_FILE" ] && { echo "Error: OpenAdmin notifications settings file not found: $INI_FILE"; exit 1; }
[ ! -f "$CONF_FILE" ] && { echo "Error: OpenPanel main configuration file not found: $CONF_FILE"; exit 1; }

mkdir -p "$(dirname "$LOG_FILE")"
[[ ! -f "$LOG_FILE" ]] && > "$LOG_FILE"

STATUS=0 PASS=0 WARN=0 FAIL=0

conf_get() { awk -F= "/^${1}=/{print \$2; exit}" "$CONF_FILE" 2>/dev/null; }
ini_get()  { awk -F= "/^${1}=/{print \$2; exit}" "$INI_FILE"  2>/dev/null; }

validate_yes_no() { [[ "$1" == "yes" || "$1" == "no" ]] && echo "$1" || echo "yes"; }
validate_number() { [[ "$1" =~ ^([1-9][0-9]?|100)$ ]] && echo "$1" || echo "$2"; }

EMAIL=$(conf_get email)
EMAIL_ALERT=$([[ -n "$EMAIL" ]] && echo "yes" || echo "no")

WEBHOOK_URL=$(ini_get webhook_url)

REBOOT=$(validate_yes_no "$(ini_get reboot)")
MAIN_DOMAIN_AND_NS=$(validate_yes_no "$(ini_get dns)")
LOGIN=$(validate_yes_no "$(ini_get login)")
SSH_LOGIN=$(validate_yes_no "$(ini_get ssh)")
SERVICES=$(ini_get services); SERVICES="${SERVICES:-admin,docker,mysql,csf,panel}"

LIMIT=$(validate_yes_no "$(ini_get limit)")
ATTACK=$(validate_yes_no "$(ini_get attack)")
MAX_TOTAL_CONN=$(validate_number "$(ini_get max_total_conn)"   5000)
MAX_CONN_PER_IP=$(validate_number "$(ini_get max_conn_per_ip)" 500)

LOAD_THRESHOLD=$(validate_number "$(ini_get load)" 20)
CPU_THRESHOLD=$(validate_number  "$(ini_get cpu)"  90)
RAM_THRESHOLD=$(validate_number  "$(ini_get ram)"  85)
DISK_THRESHOLD=$(validate_number "$(ini_get du)"   85)
SWAP_THRESHOLD=$(validate_number "$(ini_get swap)" 40)

is_unread_message_present() { grep -qF "UNREAD $1" "$LOG_FILE"; }

readonly IP_CACHE_FILE="/tmp/public.ipv4"

get_public_ip() {
  if [[ -f "$IP_CACHE_FILE" ]]; then
    local age=$(( $(date +%s) - $(stat -c %Y "$IP_CACHE_FILE") ))
    if (( age < 21600 )); then #6h
      cat "$IP_CACHE_FILE"; return
    fi
  fi
  local ip
  ip=$(curl --silent --max-time 3 -4 "https://ip.openpanel.com" 2>/dev/null \
    || curl --silent --max-time 3 -4 "https://ifconfig.me/ip" 2>/dev/null)
  if [[ -z "$ip" ]]; then
    ip=$(ip -4 addr show scope global | awk '/inet /{split($2,a,"/"); print a[1]; exit}')
  fi
  [[ -n "$ip" ]] && echo "$ip" > "$IP_CACHE_FILE"
  echo "$ip"
}


webhook_notification() {
  local title=$1 message=$2
  [[ -z "$WEBHOOK_URL" ]] && return
  local clean_msg=$(echo "$message" | sed 's/"/\\"/g' | tr '\n' ' ')
  local payload="{\"text\": \"*${title}*\n${clean_msg}\", \"username\": \"OpenAdmin-$HOSTNAME\", \"content\": \"**${title}**\n${clean_msg}\"}"
  curl -X POST -H "Content-Type: application/json" -d "$payload" --max-time 1 "$WEBHOOK_URL" >/dev/null 2>&1
}

email_notification() {
  local title=$1 message=$2
  local token; token=$(tr -dc 'a-zA-Z0-9' </dev/urandom | head -c 64)
  awk -v t="$token" '/^mail_security_token=/{$0="mail_security_token="t} 1' "$CONF_FILE" > "${CONF_FILE}.tmp" && mv "${CONF_FILE}.tmp" "$CONF_FILE"

  local domain; domain=$(opencli domain)
  local cert_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.crt"
  local key_path_on_hosts="/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/${domain}/${domain}.key"
  local fallback_cert_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.crt"
  local fallback_key_path="/etc/openpanel/caddy/ssl/custom/${domain}/${domain}.key"

  local proto="http"
  if { [ -f "$cert_path_on_hosts" ] && [ -f "$key_path_on_hosts" ]; } || { [ -f "$fallback_cert_path" ] && [ -f "$fallback_key_path" ]; }; then
      proto="https"
  fi

  local auth_opt=""
  local admin_ini="/etc/openpanel/openadmin/config/admin.ini"  
  if awk -F= '/^basic_auth=/{exit ($2=="yes")?0:1}' "$admin_ini" 2>/dev/null; then
    local u p
    u=$(awk -F= '/^basic_auth_username=/{print $2; exit}' "$admin_ini")
    p=$(awk -F= '/^basic_auth_password=/{print $2; exit}' "$admin_ini")
    auth_opt="--user ${u}:${p}"
  fi

  local admin_port resp
  admin_port=$(awk '/# START HOSTNAME DOMAIN #/{flag=1; next} /# END HOSTNAME DOMAIN #/{flag=0} flag' "/etc/openpanel/caddy/Caddyfile" | grep -oP 'localhost:\K[0-9]+' | head -n 1)
  resp=$(curl -4 --max-time 5 -ksf -X POST "$proto://$domain:$admin_port/send_email" $auth_opt -F "transient=$token" -F "recipient=$EMAIL" -F "subject=$title"   -F "body=$message" 2>/dev/null)

  case "$resp" in
    *'"error"'*)             echo "Error sending email: $resp" ;;
    *'"sent successfully"'*) echo "Email sent." ;;
  esac
}

write_notification() {
  local title="$1" message="$2" action="$3"

  # if user action: check if admin enabled notification
  if [ "$action" ]; then
    [[ -n "$RUN_ACTION_LOCKED" ]] && return
    export RUN_ACTION_LOCKED=1
    
    ACTION=$(validate_yes_no "$(ini_get $action)")
    if [[ "$ACTION" == "no" ]]; then
      echo "[!] Notifications are disabled for action: $action"; return
    fi
  else
    # if system action: check if admin already notified
    is_unread_message_present "$title" && return
  fi

  # Save to OpenAdmin > Notifications
  echo "$DISPLAY_TIME UNREAD $title MESSAGE: $message" >> "$LOG_FILE"

  # Trigger Email
  [[ "$EMAIL_ALERT" == "yes" ]] && email_notification "$title" "$message"

  # Trigger Webhook
  [[ -n "$WEBHOOK_URL" ]] && webhook_notification "$title" "$message"
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

perform_startup_actions() {
  mkdir -p /tmp/redis ; chmod 777 /tmp/redis    # for redis
  mkdir -p /tmp/ssh_cm ; chmod 700 /tmp/ssh_cm  # for ssh connections to slave servers
  if [[ "$REBOOT" == "no" ]]; then
    ((WARN++)); echo "[!] Reboot notifications are disabled."; return
  fi
  local title="SYSTEM REBOOT!"
  local message="System was rebooted. $(uptime)"
  [[ -n "$WEBHOOK_URL" ]] && webhook_notification "$title" "$message"
  write_notification "$title" "$message"
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
    if [[ "$svc" == "admin" && -f /root/openadmin_is_disabled ]]; then
      ((PASS++)); echo -e "\e[32m[✔]\e[0m $svc disabled by Administrator."; return
    fi
    local log; log=$(journalctl -n 5 -u "$svc" 2>/dev/null | sed ':a;N;$!ba;s/\n/\\n/g')
    if echo "$log" | grep -q "start-limit-hit"; then
      ((FAIL++)); STATUS=2
      echo -e "\e[31m[✘]\e[0m $svc hit start rate limit — resetting and restarting."
      systemctl reset-failed "$svc"
      systemctl restart "$svc"
      systemctl is-active --quiet "$svc" && { ((FAIL--)); echo -e "\e[32m[✔]\e[0m $svc restarted successfully."; } || { write_notification "$title" "$log"; echo -e "\e[31m[✘]\e[0m Failed to restart $svc."; }
      return
    fi
    if echo "$log" | grep -q "Deactivated successfully"; then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m $svc is inactive (disabled by Administrator), skipping restart."; return
    fi
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m $svc is not active."
    [[ -n "$log" ]] && write_notification "$title" "$log"
    systemctl restart "$svc"
    systemctl is-active --quiet "$svc" && echo -e "\e[32m[✔]\e[0m $svc restarted." || echo -e "\e[31m[✘]\e[0m Failed to restart $svc."
  fi
}

# Cache docker ps once per run — called 6+ times otherwise
DOCKER_PS_CACHE=""
_docker_ps_refresh() { DOCKER_PS_CACHE=$(docker --context=default ps --format "{{.Names}}" 2>/dev/null); }
_docker_ps() { echo "$DOCKER_PS_CACHE"; }
_docker_log() { docker --context=default logs --tail 10 "$1" 2>&1 | awk '{gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s\\n",$0}'; }

_caddy_http_ok() {
  local code
  code=$(curl -so /dev/null -w "%{http_code}" --connect-timeout 1 --max-time 1 "http://localhost/check")
  [[ "$code" == "200" || "$code" == "404" ]]
}

_docker_check_after_restart() {
  local svc=$1 title=$2
  _docker_ps_refresh
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
      CADDY_IS_ACTIVE=true
      if _caddy_http_ok; then
        ((PASS++)); echo -e "\e[32m[✔]\e[0m caddy is active and responding."
      else
        ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m caddy running but unresponsive — restarting."
        docker --context=default restart caddy &>/dev/null
        cd /root && docker --context=default compose up -d caddy &>/dev/null
        sleep 2
        _docker_ps_refresh
        if _caddy_http_ok; then
          ((PASS++)); ((WARN--)); echo -e "\e[32m[✔]\e[0m caddy recovered."
          write_notification "Caddy restarted and websites are up!" "$(_docker_log caddy)"
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
      local users; users=$(opencli user-list --json 2>/dev/null | awk -F'"' '/username/{print $4}' | grep -v SUSPENDED)
      if [[ -z "$users" || "$users" == "No users." ]]; then
        ((WARN--)); echo "  - No users found; $svc not needed."
      else
        cd /root && docker --context=default compose up -d openpanel &>/dev/null
        _docker_check_after_restart "$svc" "$title"
      fi ;;
    openpanel_dns)
      enabled_modules_line=$(grep '^enabled_modules=' "$CONF_FILE")
      if [[ "$enabled_modules_line" == *"dns"* ]]; then
          if ls /etc/bind/zones/*.zone &>/dev/null; then
              cd /root
              docker --context=default compose up -d bind9 &>/dev/null
              _docker_check_after_restart "$svc" "$title"
          else
              ((WARN--))
              echo "  - No DNS zones; bind9 not needed."
          fi
      else
          ((WARN--))
          echo "  - DNS module not enabled; bind9 not starting."
      fi ;;
    phpmyadmin)
      enabled_modules_line=$(grep '^enabled_modules=' "$CONF_FILE")
      if [[ "$enabled_modules_line" == *"phpmyadmin"* ]]; then
          if ls /home/*/sockets/mysqld/mysqld.sock &>/dev/null; then
              cd /root
              docker --context=default compose up -d phpmyadmin &>/dev/null
              _docker_check_after_restart "$svc" "$title"
          else
              ((WARN--))
              echo "  - No mysql/mariadb services yet; phpmyadmin not needed."
          fi
      else
          ((WARN--))
          echo "  - phpmyadmin module not enabled; phpmyadmin not starting."
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
      write_notification "MySQL service restarted!" "MySQL service running but not responding, attempting restart."
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
      write_notification "MySQL restarted successfully!" "Sentinel restarted MySQL and it is responding now."
    else
      echo "    Error: MySQL still not responding!"
      write_notification "$title" "MySQL did not respond after restart. Please check ASAP."
    fi
  fi
}

check_services() {
  local svc
  for svc in caddy csf admin docker panel mysql phpmyadmin named; do
    [[ ",$SERVICES," != *",$svc,"* ]] && continue
    case "$svc" in
      caddy)  docker_containers_status  'caddy'         'Caddy not active — websites down!'             ;;
      phpmyadmin)  docker_containers_status  'phpmyadmin'         'phpmyadmin not active — users can not access databases!'             ;;
      csf)    check_service_status      'csf'           'CSF Firewall not active — server unprotected!' ;;
      admin)  check_service_status      'admin'         'OpenAdmin service not accessible!'             ;;
      docker) check_service_status      'docker'        'Docker not active — user websites down!'       ;;
      panel)  docker_containers_status  'openpanel'     'OpenPanel container not running!'              ;;
      mysql)  mysql_docker_containers_status                                                            ;;
      named)  docker_containers_status  'openpanel_dns' 'BIND9 not active — DNS broken!'                ;;
    esac
  done
}

check_oom_logs() {
  if [[ "$LIMIT" == "no" ]]; then
    ((WARN++)); echo "[!] OOM errors check disabled."; return
  fi

  local NOW EPOCH_LAST DIFF

  NOW=$(date +%s)
  if [[ -f "$LOCK_FILE_FOR_OOM_CHECK" ]]; then
    EPOCH_LAST=$(cat "$LOCK_FILE_FOR_OOM_CHECK" 2>/dev/null)
    DIFF=$((NOW - EPOCH_LAST))
    [[ "$DIFF" -lt 86400 ]] && return
  fi

  echo "$NOW" > "$LOCK_FILE_FOR_OOM_CHECK"

  local TODAY LOG
  TODAY=$(date +%Y-%m-%d)
  if [[ -f /var/log/syslog ]]; then
    LOG="/var/log/syslog"
  elif [[ -f /var/log/messages ]]; then
    LOG="/var/log/messages"
  else
    return
  fi

  local SYSTEM_COUNT=0
  local USER_COUNT=0
  local SYSTEM_MSG=""
  local USER_MSG=""

  while read -r line; do
    uid=$(echo "$line" | sed -n 's/.*UID:\([0-9]\+\).*/\1/p')
    [[ -z "$uid" ]] && continue

    if [[ "$uid" -eq 0 ]]; then
        ((SYSTEM_COUNT++))
        SYSTEM_MSG+=$' | '"$line"
    elif [[ "$uid" -ge 1000 ]]; then
        ((USER_COUNT++))
        user=$(getent passwd "$uid" | cut -d: -f1)
        [[ -z "$user" ]] && continue
        USER_MSG+=$' | '"$user: $line"
    fi

  done < <(grep "Memory cgroup out of memory: Killed process" "$LOG" | grep "^$TODAY")

  if [[ "$SYSTEM_COUNT" -eq 0 && "$USER_COUNT" -eq 0 ]]; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No OOM errors detected."; return
  else
    ((FAIL++)); STATUS=2
  fi

  title_parts=()

  [[ "$SYSTEM_COUNT" -gt 0 ]] && title_parts+=("System: $SYSTEM_COUNT")
  [[ "$USER_COUNT" -gt 0 ]] && title_parts+=("User: $USER_COUNT")

  title="OOM Alert - $TODAY - $(IFS=' | '; echo "${title_parts[*]}")"
  message=""

  if [[ "$SYSTEM_COUNT" -gt 0 ]]; then
    message+="$SYSTEM_COUNT system service(s) killed by OOM in the last 24 hours $SYSTEM_MSG"
    echo -e "\e[31m[✘]\e[0m $SYSTEM_COUNT system service(s) killed by OOM in the last 24 hours"
  fi

  if [[ "$USER_COUNT" -gt 0 ]]; then
    message+="$USER_COUNT user process(es) killed by OOM in the last 24 hours $USER_MSG"
    echo -e "\e[31m[✘]\e[0m $USER_COUNT user process(es) killed by OOM in the last 24 hours"
  fi

  [[ -n "$WEBHOOK_URL" ]] && webhook_notification "$title" "$message"
  write_notification "$title" "$message"
}

check_new_logins() {
  if [[ "$LOGIN" == "no" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Login check disabled."; return
  fi

  local login_log="/var/log/openpanel/admin/login.log"
  local watermark_file="/tmp/sentinel.login_watermark"
  [[ ! -f "$login_log" ]] && > "$login_log"

  local last_count=0
  [[ -f "$watermark_file" ]] && last_count=$(cat "$watermark_file" 2>/dev/null)
  local current_count; current_count=$(wc -l < "$login_log")
  echo "$current_count" > "$watermark_file"

  if (( current_count <= last_count )); then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin."; return
  fi

  local new_lines; new_lines=$(tail -n +"$((last_count + 1))" "$login_log")

  if [[ -z "$new_lines" ]]; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin."; return
  fi

  local seen_pairs=""
  if (( last_count > 0 )); then
    seen_pairs=$(head -n "$last_count" "$login_log" | awk '{print $(NF-1), $NF}')
  fi

  local found_new=0
  while IFS= read -r line; do
    local username ip_address
    read -r _ _ username ip_address _ <<< "$line"

    [[ ! "$ip_address" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] && continue
    [[ "$ip_address" == "127.0.0.1" ]] && continue

    if ! grep -qF "$username $ip_address" <<< "$seen_pairs"; then
      ((FAIL++)); STATUS=2; found_new=1
      echo -e "\e[31m[✘]\e[0m $username logged in from new IP: $ip_address"
      write_notification "Admin $username accessed from new IP" "Admin account $username was accessed from new IP: $ip_address"
    else
      echo -e "\e[32m[✔]\e[0m $username from known IP: $ip_address"
    fi
  done <<< "$new_lines"

  (( found_new == 0 )) && ((PASS++))
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
  local flag_tt=86400
  local title="Running out of Disk Space!"
  if is_unread_message_present "$title"; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Unread DU notification. Skipping."; return
  fi
  local pct; pct=$(df --output=pcent / | awk 'NR==2{gsub(/%/,"",$1); print $1+0}')
  if (( pct > DISK_THRESHOLD )); then
    # Try cleanup if not done in last 24h
    if [ -f "$LOCK_FILE_FOR_DOCKER_PRUNE" ]; then
      local age=$(( $(date +%s) - $(stat -c %Y "$LOCK_FILE_FOR_DOCKER_PRUNE") ))
      if [ "$age" -lt "$flag_tt" ]; then
        $DEBUG && echo "[$context] Skipping cleanup — ran ${age}s ago"
        write_notification "$title" "Disk usage: ${pct}% | Partitions: $(df -h | sort -r -k 5 -i | sed ':a;N;$!ba;s/\n/\\n/g')"
        return
      fi
    fi

    local kb_before; kb_before=$(df / | awk 'NR==2 {print $3}')

    timeout 30 docker --context="default" system prune -f --filter "until=24h" > /dev/null 2>&1
    for context in /home/*; do
       [ -d "$context" ] || continue
       context_name=$(basename "$context")
       timeout 15 docker --context="$context_name" system prune -f --filter "until=24h" > /dev/null 2>&1
    done
    touch "$LOCK_FILE_FOR_DOCKER_PRUNE"

    local kb_after; kb_after=$(df / | awk 'NR==2 {print $3}')
    local freed_gb=$(( (kb_before - kb_after) / 1024 / 1024 ))
    local pct_after; pct_after=$(df --output=pcent / | awk 'NR==2{gsub(/%/,"",$1); print $1+0}')

    if (( freed_gb >= 1 )); then
      write_notification "$title" "Disk was ${pct}%, freed ${freed_gb}GB, now at ${pct_after}% | Partitions: $(df -h | sort -r -k 5 -i | sed ':a;N;$!ba;s/\n/\\n/g')"
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m Disk was ${pct}%, sentinel freed ${freed_gb}GB, disk is now at ${pct_after}%."; return   
    else
      ((FAIL++)); STATUS=2
      echo -e "\e[31m[✘]\e[0m Disk was ${pct}% > threshold ${DISK_THRESHOLD}%"
      write_notification "$title" "Disk usage: ${pct}% | Partitions: $(df -h | sort -r -k 5 -i | sed ':a;N;$!ba;s/\n/\\n/g')"
    fi
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
    write_notification "$title" "Load: $load | Crashlog: $REPORT"
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
    local procs; procs=$(ps ax --sort=-%mem -o pid:7,pmem:6,comm:20 | head -10 | sed ':a;N;$!ba;s/\n/\\n/g')
    write_notification "$title" "Used RAM: ${used}MB / ${total}MB (${pct}%) | $procs"
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
    local procs; procs=$(ps ax --sort=-%cpu -o pid:7,pcpu:6,comm:20 | head -10 | sed ':a;N;$!ba;s/\n/\\n/g')
    write_notification "$title" "CPU: ${pct}% | $procs"
  else
    ((PASS++)); echo -e "\e[32m[✔]\e[0m CPU ${pct}% < threshold ${CPU_THRESHOLD}%"
  fi
}

check_https_traffic() {
  if [[ "$ATTACK" == "no" ]]; then
    ((WARN++)); echo "[!] Website traffic checks are disabled."; return
  fi
  if [[ "$CADDY_IS_ACTIVE" != "true" ]]; then
    ((WARN++)); echo "[!] Skipping website traffic checks because Caddy is not running."; return
  fi
  local ALERT=0
  local ALL_CONNS
  ALL_CONNS=$(ss -tn '( sport = :80 or sport = :443 )' | tail -n +2)

  # SYN flooding
  while read -r COUNT PORT; do
    if (( COUNT >= 1000 )); then
      echo -e "\e[31m[✘]\e[0m Possible SYN flood on :$PORT — $COUNT in SYN_RECV"
      write_notification "Possible SYN flood" "Port $PORT: $COUNT connections in SYN_RECV state"
      ALERT=1
    fi
  done < <(awk '/SYN-RECV/{split($4,a,":"); print a[length(a)]}' <<< "$ALL_CONNS" | sort | uniq -c | awk '{print $1, $2}')

  local HIGH_TRAFFIC_LINES=()
  while read -r COUNT PORT IP; do
    if (( COUNT >= MAX_CONN_PER_IP )); then
      echo -e "\e[31m[✘]\e[0m High connections from $IP on port $PORT: $COUNT"

      local DOMAINS
      DOMAINS=$(
        find /var/log/caddy/domlogs -type f -name "access.log" 2>/dev/null |
        while read -r f; do
          if timeout 1s bash -c '
            tail -n 200 "$1" | grep -q "$2"
          ' _ "$f" "$IP"; then
            echo "$f"
          fi
        done | sed 's|/var/log/caddy/domlogs/||; s|/access\.log||' | sort -u | tr '\n' ', ' | sed 's/,$//'
      )

      if [[ -n "$DOMAINS" ]]; then
        echo -e "    \e[33m[→]\e[0m Domain hit: $DOMAINS"
        HIGH_TRAFFIC_LINES+=("$IP (port $PORT: $COUNT conns | domains: $DOMAINS)")
      else
        echo -e "    \e[33m[→]\e[0m No matching domain logs found, check manually with: 'grep -Rli --include=access.log $IP /var/log/caddy/domlogs/ | xargs -n1 dirname | xargs -n1 basename'"
        HIGH_TRAFFIC_LINES+=("$IP (port $PORT: $COUNT conns)")
      fi

      ALERT=1
    fi
  done < <(
    awk '/ESTAB/{
      split($4, a, ":")
      port = a[length(a)]
      peer = $5
      if (match(peer, /^\[(.+)\]:[0-9]+$/, m)) {
          ip = m[1]
      } else {
          n = split(peer, b, ":")
          ip = b[1]
      }
      sub(/^::ffff:/, "", ip)
      print port, ip
    }' <<< "$ALL_CONNS" |
    sort | uniq -c | awk '{print $1, $2, $3}'
  )

  if (( ${#HIGH_TRAFFIC_LINES[@]} > 0 )); then
    local NOTIF_BODY
    printf -v NOTIF_BODY '%s\\n' "${HIGH_TRAFFIC_LINES[@]}"
    NOTIF_BODY="${NOTIF_BODY%$'\n'}"
    if (( ${#HIGH_TRAFFIC_LINES[@]} == 1 )); then
      write_notification "High traffic from ${HIGH_TRAFFIC_LINES[0]%%(*}" "$NOTIF_BODY"
    else
      write_notification "High traffic from ${#HIGH_TRAFFIC_LINES[@]} IPs" "$NOTIF_BODY"
    fi
  fi

  # Total established
  local TOTAL_CONN
  TOTAL_CONN=$(awk '/ESTAB/' <<< "$ALL_CONNS" | wc -l)
  if (( TOTAL_CONN >= MAX_TOTAL_CONN )); then
    echo -e "\e[31m[✘]\e[0m High total connections: $TOTAL_CONN"
    write_notification "High total connections" "$TOTAL_CONN total connections on ports 80/443"
    ALERT=1
  fi

  if [[ $ALERT -eq 0 ]]; then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m No unusual traffic detected on web ports (80|443)."
  else
    ((FAIL++)); STATUS=2
  fi
}

check_swap_usage() {
  local title="High SWAP usage!"
  local _ stotal sused _rest
  read -r _ stotal sused _rest < <(free | awk '/^Swap:/')

  if (( stotal == 0 )); then
    local swap_devices
    swap_devices=$(swapon --show=NAME --noheadings 2>/dev/null)
    if [[ -z "$swap_devices" ]]; then
      local fstab_swap
      fstab_swap=$(awk '$3=="swap"{print $1}' /etc/fstab)
      if [[ -n "$fstab_swap" ]]; then
        echo -e "\e[38;5;214m[!]\e[0m SWAP is off but fstab entries exist. Attempting to re-enable..."
        if swapon -a 2>/dev/null; then
          read -r _ stotal sused _rest < <(free | awk '/^Swap:/')
          if (( stotal > 0 )); then
            ((WARN++))
            write_notification "SWAP re-enabled on $HOSTNAME" "Sentinel detected swap was off and re-enabled it. Now: ${stotal}kB total."
            echo -e "\e[32m[✔]\e[0m SWAP successfully re-enabled (${stotal}kB total)."
          else
            ((WARN++))
            echo -e "\e[38;5;214m[!]\e[0m swapon -a ran but swap still reports 0. Check swap device health."
            write_notification "SWAP re-enable failed on $HOSTNAME" "swapon -a returned success but swap is still 0."
            return
          fi
        else
          ((WARN++))
          echo -e "\e[31m[✘]\e[0m Failed to re-enable SWAP. Check swap device/file."
          write_notification "SWAP re-enable failed on $HOSTNAME" "swapon -a failed on $HOSTNAME at $DISPLAY_TIME."
          return
        fi
      else
        ((PASS++)); echo -e "\e[32m[✔]\e[0m No SWAP configured."; return
      fi
    else
      ((PASS++)); echo -e "\e[32m[✔]\e[0m No SWAP configured."; return
    fi
  fi

  local pct=$(( sused * 100 / stotal ))
  if (( pct <= SWAP_THRESHOLD )); then
    ((PASS++)); echo -e "\e[32m[✔]\e[0m SWAP ${pct}% < threshold ${SWAP_THRESHOLD}%"
    rm -f "$LOCK_FILE_FOR_SWAP_CLEANUP"; return
  fi

  if [[ -f "$LOCK_FILE_FOR_SWAP_CLEANUP" ]]; then
    local age=$(( $(date +%s) - $(date -r "$LOCK_FILE_FOR_SWAP_CLEANUP" +%s) ))
    if (( age <= 86400 )); then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m SWAP cleanup already in progress. Skipping."; return
    fi
    rm -f "$LOCK_FILE_FOR_SWAP_CLEANUP"
  fi

  local available_mb
  available_mb=$(awk '/^MemAvailable:/{print int($2/1024)}' /proc/meminfo)
  if (( available_mb < sused )); then
    echo "Not enough free RAM to safely clear swap (${available_mb}MB available, ${sused}MB in swap). Skipping."
    write_notification "SWAP high but cannot safely clear" "SWAP: ${pct}%, only ${available_mb}MB RAM available vs ${sused}MB in swap."
    ((WARN++)); return
  fi

  echo -e "\e[31m[✘]\e[0m SWAP ${pct}% > threshold ${SWAP_THRESHOLD}%. Clearing..."
  write_notification "$title" "SWAP: ${pct}%. Cleanup starting."
  touch "$LOCK_FILE_FOR_SWAP_CLEANUP"
  sync ; echo 3 > /proc/sys/vm/drop_caches
  swapoff -a ; swapon -a

  local stotal2 sused2
  read -r _ stotal2 sused2 _rest < <(free | awk '/^Swap:/')
  local pct2=$(( stotal2 > 0 ? sused2*100/stotal2 : 0 ))
  if (( pct2 < SWAP_THRESHOLD )); then
    rm -f "$LOCK_FILE_FOR_SWAP_CLEANUP"
    write_notification "SWAP cleared — now ${pct2}%" "Sentinel cleared SWAP on $HOSTNAME."
    echo -e "\e[32m[✔]\e[0m SWAP cleared successfully. Now: ${pct2}%"
  else
    ((FAIL++)); STATUS=2
    echo -e "\e[31m[✘]\e[0m SWAP still high after cleanup: ${pct2}%"
    write_notification "URGENT: SWAP not cleared on $HOSTNAME" "SWAP cleanup attempted at $DISPLAY_TIME but usage still ${pct2}%."
  fi
}

generate_crashlog_report() {
  local dir="/var/log/openpanel/admin/crashlog"
  mkdir -p "$dir"
  REPORT="$dir/$(date +%s).txt"
  {
    echo "=== GENERAL === Hostname: $HOSTNAME | Date: $DISPLAY_TIME"
    echo "=== LOAD ===";     cat /proc/loadavg
    echo "=== TOP (MEM) ==="; ps -eo pid:7,%mem:5,comm:20 --sort=-%mem | head -11
    echo "=== TOP (CPU) ==="; ps -eo pid:7,%cpu:5,comm:20 --sort=-%cpu | head -11
    echo "=== SWAP TOP ===";  grep VmSwap /proc/*/status 2>/dev/null | sort -k2 -hr | head -10
    echo "=== DISKSTATS ==="; head -10 /proc/diskstats
  } > "$REPORT"
}

check_if_panel_domain_and_ns_resolve_to_server() {
  if [[ "$MAIN_DOMAIN_AND_NS" == "no" ]]; then
    ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m DNS check disabled."; return
  fi

  if [[ -f "$LOCK_FILE_FOR_DNS_CHECK" ]]; then
    local age=$(( $(date +%s) - $(date -r "$LOCK_FILE_FOR_DNS_CHECK" +%s) ))
    if (( age < 3600 )); then
      ((WARN++)); echo -e "\e[38;5;214m[!]\e[0m DNS check skipped (last run $((age/60))m ago)."; return
    fi
  fi
  touch "$LOCK_FILE_FOR_DNS_CHECK"

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
  local SERVER_IP; SERVER_IP=$(get_public_ip)

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
        write_notification "$FORCED_DOMAIN does not resolve to $SERVER_IP" "$FORCED_DOMAIN should point to $SERVER_IP."
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

write_snapshot() {
  mkdir -p "$(dirname "$SNAPSHOT_FILE")"

  local load_1m load_5m load_15m
  read -r load_1m load_5m load_15m _ < /proc/loadavg

  local mem_total mem_used mem_free mem_pct=0
  read -r _ mem_total mem_used mem_free _ < <(free -m | awk '/^Mem:/')
  (( mem_total > 0 )) && mem_pct=$(( mem_used * 100 / mem_total ))

  local swap_total=0 swap_used=0 swap_pct=0
  read -r _ swap_total swap_used _ < <(free -m | awk '/^Swap:/')
  (( swap_total > 0 )) && swap_pct=$(( swap_used * 100 / swap_total ))

  local disk_pct; disk_pct=$(df --output=pcent / | awk 'NR==2{gsub(/%/,"",$1); print $1+0}')

  local cpu_pct=0
  local -a c1 c2
  read -ra c1 < <(grep '^cpu ' /proc/stat)
  sleep 0.1
  read -ra c2 < <(grep '^cpu ' /proc/stat)
  local t1=0 t2=0 v
  for v in "${c1[@]:1}"; do (( t1 += v )); done
  for v in "${c2[@]:1}"; do (( t2 += v )); done
  local dt=$(( t2 - t1 ))
  (( dt > 0 )) && cpu_pct=$(( 100 * (dt - (c2[4]-c1[4])) / dt ))

  local total_conn=0
  total_conn=$(ss -tn '( sport = :80 or sport = :443 )' 2>/dev/null | grep -c ESTAB || true)

  local json
  printf -v json \
    '{"ts":"%s","load":{"1m":"%s","5m":"%s","15m":"%s"},"cpu_pct":%d,"mem":{"total_mb":%d,"used_mb":%d,"pct":%d},"swap":{"total_mb":%d,"used_mb":%d,"pct":%d},"disk_pct":%d,"web_conn":%d,"status":%d,"pass":%d,"warn":%d,"fail":%d}' \
    "$DISPLAY_TIME" "$load_1m" "$load_5m" "$load_15m" \
    "$cpu_pct" \
    "$mem_total" "$mem_used" "$mem_pct" \
    "$swap_total" "$swap_used" "$swap_pct" \
    "$disk_pct" \
    "$total_conn" \
    "$STATUS" "$PASS" "$WARN" "$FAIL"

  local utc_json
  utc_json=$(echo "$json" | sed 's/"ts":"[^"]*"/"ts":"'"$(date -u +"%Y-%m-%d %H:%M:%S")"'"/')
  echo "$utc_json" >> "$SNAPSHOT_FILE"


  local line_count; line_count=$(wc -l < "$SNAPSHOT_FILE")
  if (( line_count > SNAPSHOT_MAX_LINES )); then
    local tmp; tmp=$(mktemp)
    tail -n "$SNAPSHOT_MAX_LINES" "$SNAPSHOT_FILE" > "$tmp" && mv "$tmp" "$SNAPSHOT_FILE"
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


for arg in "$@"; do
  case "$arg" in
    --startup) perform_startup_actions; exit 0 ;;
    --report) email_daily_report; exit 0 ;;
    --action=*) action="${arg#*=}" ;;
    --message=*) message="${arg#*=}" ;;
    --title=*) title="${arg#*=}";;
  esac
done

if [[ -n "$action" ]]; then
  write_notification "$title" "$message" "$action"
  exit 0
fi

exec 200>/root/sentinel_run.lock
flock -n 200 || { echo "Error: Another instance is already running."; exit 1; }

hr
echo "  Sentinel - OpenPanel server health monitor"
hr

echo "Checking services:"
_docker_ps_refresh
check_services
check_oom_logs

hr
echo "Checking traffic:"
check_https_traffic

hr
echo "Checking logins, resources, and DNS..."

declare -A _pids _outfiles
_parallel_tasks=(
  check_new_logins
  check_ssh_logins
  check_disk_usage
  check_system_load
  check_ram_usage
  check_cpu_usage
  check_swap_usage
  check_if_panel_domain_and_ns_resolve_to_server
)

for _task in "${_parallel_tasks[@]}"; do
  _out=$(mktemp /tmp/sentinel.par.XXXXXX)
  _outfiles[$_task]="$_out"
  (
    STATUS=0 PASS=0 WARN=0 FAIL=0
    $_task
    echo "__COUNTERS__ $STATUS $PASS $WARN $FAIL"
  ) > "$_out" 2>&1 &
  _pids[$_task]=$!
done

for _task in "${_parallel_tasks[@]}"; do
  wait "${_pids[$_task]}" 2>/dev/null
  _out="${_outfiles[$_task]}"
  [[ -f "$_out" ]] || continue
  grep -v '^__COUNTERS__' "$_out"
  read -r _ _s _p _w _f < <(grep '^__COUNTERS__' "$_out")
  (( STATUS  = STATUS  > _s ? STATUS  : _s ))
  (( PASS   += _p ))
  (( WARN   += _w ))
  (( FAIL   += _f ))
  rm -f "$_out"
done

write_snapshot
summary

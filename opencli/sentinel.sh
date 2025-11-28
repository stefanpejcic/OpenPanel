#!/bin/bash
################################################################################
# Script Name: sentinel.sh
# Description: OpenAdmin Notifications
# Usage: opencli sentinel [-report|--startup]
# Author: Stefan Pejcic
# Created: 15.11.2023
# Last Modified: 27.11.2025
# Company: openpanel.com
# Copyright (c) openpanel.com
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

PID=$$
DEBUG=false


# ======================================================================
# Constants
VERSION="20.251.104"


if [ "$1" = "--debug" ]; then
    DEBUG=true
fi

# ======================================================================
# Counters
STATUS=0
PASS=0
WARN=0
FAIL=0

readonly CONF_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
readonly LOCK_FILE="/tmp/swap_cleanup.lock"
TIME=$(date +%s%3N)
readonly INI_FILE="/etc/openpanel/openadmin/config/notifications.ini"
HOSTNAME=$(hostname)
readonly LOG_FILE="/var/log/openpanel/admin/notifications.log"
LOG_DIR=$(dirname "$LOG_FILE")

mkdir -p "$LOG_DIR"
touch "$LOG_FILE"

show_execution_time() {
    end_time=$(date +%s%3N)
    execution_time=$((end_time - TIME))
    seconds=$((execution_time / 1000))
    milliseconds=$((execution_time % 1000))
    memory_usage=$(ps -p $$ -o rss=)
    echo "Elapsed time: ${seconds}.${milliseconds} seconds"
    echo "Memory usage: ${memory_usage} KB"
}

trap show_execution_time EXIT

if [ ! -f "$INI_FILE" ]; then
    echo "Error: INI file not found: $INI_FILE"
    exit 1
fi





# ======================================================================
# Helper functions

print_header() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -e ""
    echo -e " _____               _    _               _ "
    echo -e "/  ___|             | |  (_)             | |"
    echo -e "\\ \`--.   ___  _ __  | |_  _  _ __    ___ | |"
    echo -e " \`--. \\ / _ \\| \`_ \\ | __|| || \`_ \\  / _ \\| |"
    echo -e "/\\__/ /|  __/| | | || |_ | || | | ||  __/| |"
    echo -e "\\____/  \\___||_| |_| \\__||_||_| |_| \\___||_|"
    echo -e ""
    echo -e "            version: $VERSION"
    echo -e ""
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}

check_for_debug_and_print_info(){
  if [ "$DEBUG" = true ]; then
      echo ""
      echo "----------------- DEBUG INFORMATION ------------------"
      echo "HOSTNAME:       $HOSTNAME"
      echo "VERSION:        $VERSION"
      echo "PID:            $PID"
      echo "TIME:           $TIME"
      echo "CONF_FILE:      $CONF_FILE"
      echo "LOCK_FILE:      $LOCK_FILE"
      echo "INI_FILE:       $INI_FILE"
      echo "LOG_FILE:       $LOG_FILE"
      echo "LOG_DIR:        $LOG_DIR"
      echo ""
      echo "----------------- CONFIGURATIONS ------------------"
      echo ""
      echo "EMAIL_ALERT:     $EMAIL_ALERT"
      echo "EMAIL:           $EMAIL"
      echo "REBOOT:          $REBOOT"
      echo "LOGIN:           $LOGIN"
      echo "ATTACK:          $ATTACK"
      echo "LIMIT:           $LIMIT"
      echo "UPDATE:          $UPDATE"
      echo "SERVICES:        $SERVICES"
      echo "LOAD_THRESHOLD:  $LOAD_THRESHOLD"
      echo "CPU_THRESHOLD:   $CPU_THRESHOLD"
      echo "RAM_THRESHOLD:   $RAM_THRESHOLD"
      echo "DISK_THRESHOLD:  $DISK_THRESHOLD"
      echo "SWAP_THRESHOLD:  $SWAP_THRESHOLD"
      echo "------------------------------------------------------"
      echo ""
  fi
}


generate_random_token() {
    tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64
}

generate_random_token_one_time_only() {
    TOKEN_ONE_TIME="$(generate_random_token)"
    local new_value="mail_security_token=$TOKEN_ONE_TIME"
    sed -i "s|^mail_security_token=.*$|$new_value|" "$CONF_FILE"
}

is_valid_number() {
  local value="$1"
  [[ "$value" =~ ^[1-9][0-9]?$|^100$ ]]
}

get_last_message_content() {
  tail -n 1 "$LOG_FILE" 2>/dev/null
}

is_unread_message_present() {
  local unread_message_content="$1"
  grep -q "UNREAD.*$unread_message_content" "$LOG_FILE" && return 0 || return 1
}

ensure_installed() {
    local cmd="$1"
    local pkg="$2"
    pkg="${pkg:-$cmd}"

    if ! command -v "$cmd" &> /dev/null; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get update -qq > /dev/null 2>&1
            sudo apt-get install -y -qq "$pkg" > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q "$pkg" > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q "$pkg" > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install $cmd manually."
            exit 1
        fi

        if ! command -v "$cmd" &> /dev/null; then
            echo "Error: $cmd installation failed. Please install $pkg manually."
            exit 1
        fi
    fi
}

email_notification() {
  local title="$1"
  local message="$2"

  generate_random_token_one_time_only

  TRANSIENT=$(awk -F'=' '/^mail_security_token/ {print $2}' "$CONF_FILE")
  DOMAIN=$(opencli domain)
  if [[ "$DOMAIN" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    PROTOCOL="https"
  else
    PROTOCOL="http"
  fi

  admin_conf_file="/etc/openpanel/openadmin/config/admin.ini"
  AUTH_OPTION=""
  if grep -q '^basic_auth=yes' "$admin_conf_file"; then
      username=$(grep '^basic_auth_username=' "$admin_conf_file" | cut -d'=' -f2)
      password=$(grep '^basic_auth_password=' "$admin_conf_file" | cut -d'=' -f2)
      AUTH_OPTION="--user ${username}:${password}"
  fi

  response=$(curl -4 --max-time 5 -k -X POST "$PROTOCOL://$DOMAIN:2087/send_email" \
    $AUTH_OPTION \
    -F "transient=$TRANSIENT" \
    -F "recipient=$EMAIL" \
    -F "subject=$title" \
    -F "body=$message" 2>/dev/null)

  if echo "$response" | grep -q '"error"'; then
    echo "Error: Failed to send email. Response: $response"
  elif echo "$response" | grep -q '"sent successfully"'; then
    echo "Email sent successfully."
  fi
}

write_notification() {
  local title="$1"
  local message="$2"
  local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"
  local last_message_content=$(get_last_message_content)
  if [ "$message" != "$last_message_content" ] && ! is_unread_message_present "$title"; then
    echo "$current_message" >> "$LOG_FILE"
    if [ "$EMAIL_ALERT" == "no" ]; then
      echo "Email alerts are disabled."
    else
      email_notification "$title" "$message"
      if command -v notify-send &>/dev/null; then
          notify-send "$title" "$message"
      fi
    fi
  fi
}



# ======================================================================
# What to check

EMAIL=$(awk -F'=' '/^email/ {print $2}' "$CONF_FILE")
EMAIL_ALERT=$([[ -n "$EMAIL" ]] && echo "yes" || echo "no")


REBOOT=$(awk -F'=' '/^reboot/ {print $2}' "$INI_FILE")
REBOOT=${REBOOT:-yes}
[[ "$REBOOT" =~ ^(yes|no)$ ]] || REBOOT=yes


MAIN_DOMAIN_AND_NS=$(awk -F'=' '/^dns/ {print $2}' "$INI_FILE")
MAIN_DOMAIN_AND_NS=${MAIN_DOMAIN_AND_NS:-yes}
[[ "$MAIN_DOMAIN_AND_NS" =~ ^(yes|no)$ ]] || MAIN_DOMAIN_AND_NS=yes


LOGIN=$(awk -F'=' '/^login/ {print $2}' "$INI_FILE")
LOGIN=${LOGIN:-yes}
[[ "$LOGIN" =~ ^(yes|no)$ ]] || LOGIN=yes

SSH_LOGIN=$(awk -F'=' '/^ssh/ {print $2}' "$INI_FILE")
SSH_LOGIN=${SSH_LOGIN:-yes}
[[ "$SSH_LOGIN" =~ ^(yes|no)$ ]] || SSH_LOGIN=yes

ATTACK=$(awk -F'=' '/^attack/ {print $2}' "$INI_FILE")
ATTACK=${ATTACK:-yes}
[[ "$ATTACK" =~ ^(yes|no)$ ]] || ATTACK=yes

LIMIT=$(awk -F'=' '/^limit/ {print $2}' "$INI_FILE")
LIMIT=${LIMIT:-yes}
[[ "$LIMIT" =~ ^(yes|no)$ ]] || LIMIT=yes

UPDATE=$(awk -F'=' '/^update/ {print $2}' "$INI_FILE")
UPDATE=${UPDATE:-yes}
[[ "$UPDATE" =~ ^(yes|no)$ ]] || UPDATE=yes

SERVICES=$(awk -F'=' '/^services/ {print $2}' "$INI_FILE")
SERVICES=${SERVICES:-"admin,docker,mysql,csf,panel"}

LOAD_THRESHOLD=$(awk -F'=' '/^load/ {print $2}' "$INI_FILE")
LOAD_THRESHOLD=${LOAD_THRESHOLD:-20}
is_valid_number "$LOAD_THRESHOLD" || LOAD_THRESHOLD=20

CPU_THRESHOLD=$(awk -F'=' '/^cpu/ {print $2}' "$INI_FILE")
CPU_THRESHOLD=${CPU_THRESHOLD:-90}
is_valid_number "$CPU_THRESHOLD" || CPU_THRESHOLD=90

RAM_THRESHOLD=$(awk -F'=' '/^ram/ {print $2}' "$INI_FILE")
RAM_THRESHOLD=${RAM_THRESHOLD:-85}
is_valid_number "$RAM_THRESHOLD" || RAM_THRESHOLD=85

DISK_THRESHOLD=$(awk -F'=' '/^du/ {print $2}' "$INI_FILE")
DISK_THRESHOLD=${DISK_THRESHOLD:-85}
is_valid_number "$DISK_THRESHOLD" || DISK_THRESHOLD=85

SWAP_THRESHOLD=$(awk -F'=' '/^swap/ {print $2}' "$INI_FILE")
SWAP_THRESHOLD=${SWAP_THRESHOLD:-40}
is_valid_number "$SWAP_THRESHOLD" || SWAP_THRESHOLD=40






# ====================================================================== #
#                             MAIN FUNCTIONS                             #



# ====== ON REBOOT
perform_startup_action() {
  if [ "$REBOOT" != "no" ]; then
    local title="SYSTEM REBOOT!"
    local uptime=$(uptime)
    local message="System was rebooted. $uptime"
    write_notification "$title" "$message"
  else
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Reboot is explicitly set to 'no' in the INI file. Skipping logging.."
  fi
}

# ====== CHECK SSH LOGINS
check_ssh_logins() {

  if [ "$SSH_LOGIN" != "no" ]; then

    local title="Suspicious SSH login detected"
    local message_to_check_in_file="Suspicious SSH login detected"

    if is_unread_message_present "$message_to_check_in_file"; then
      ((WARN++))
      echo -e "\e[38;5;214m[!]\e[0m Unread SSH login notification already exists. Skipping."
      return
    fi

    ssh_ips=$(who | grep 'pts' | awk '{print $5}' | sed -E 's/[():]//g' | cut -d':' -f1)

    if [ -z "$ssh_ips" ]; then
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m No currently logged in SSH users detected."
        return
    fi

    login_ips=$(awk '{print $NF}' /var/log/openpanel/admin/login.log)

    if [ -z "$login_ips" ]; then
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Detected logged-in SSH user, but login checks will be postponed until OpenAdmin interface is ready."
        return
    fi

  counter=0
  safe_counter=0
  suspecious_ips=()
  safe_ips=()

  for ip in $ssh_ips; do
    # make sure we get ip!
    if [[ $ip =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
      if ! echo "$login_ips" | grep -q "$ip"; then
          ((counter++))
          suspecious_ips+=("$ip")
      else
          ((safe_counter++))
          safe_ips+=("$ip")
      fi
    fi
  done

  if [ $counter -gt 0 ]; then
      for unmatched_ip in "${suspecious_ips[@]}"; do
          message+="  $unmatched_ip"
      done

    ((FAIL++))
    STATUS=2

    echo -e "\e[31m[✘]\e[0m $message | Writing notification."
    write_notification "$title" "$message"

  else
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m Detected $safe_counter currently active SSH user(s), but marked as safe since the IP address has previously logged into the OpenAdmin interface."
      message=""
      for ip in "${safe_ips[@]}"; do
          message+="${ip}    "
      done
      echo -e "    Currently active IP addresses: $message"
  fi


  fi


}

# ====== ADMIN LOGINS
check_new_logins() {
  if [ "$LOGIN" != "no" ]; then
    touch /var/log/openpanel/admin/login.log
    if [ ! -s "$LOG_FILE" ]; then
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin detected."
    else

    last_login=$(tail -n 1 /var/log/openpanel/admin/login.log)
    if [ -z "$last_login" ]; then
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin detected."
      return 1
    fi

    username=$(echo "$last_login" | awk '{print $3}')
    ip_address=$(echo "$last_login" | awk '{print $4}')

    if [[ ! $ip_address =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo "Invalid IP address format: $ip_address"
      return 1
    fi

    if [[ $ip_address == "127.0.0.1" ]]; then
      echo "IP address 127.0.0.1 is not allowed."
      return 1
    fi

    if [ $(grep -c $username /var/log/openpanel/admin/login.log) -eq 1 ]; then
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m First time login detected for user: $username from IP address: $ip_address."
    else
      if ! grep -q "$username $ip_address" <(sed '$d' /var/log/openpanel/admin/login.log); then
        echo -e "\e[31m[✘]\e[0m Admin account $username accessed from new IP address $ip_address"
        ((FAIL++))
        STATUS=2
        local title="Admin account $username accessed from new IP address"
        local message="Admin account $username was accessed from a new IP address: $ip_address"
        write_notification "$title" "$message"
      else
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m Admin account $username accessed from previously logged IP address: $ip_address"
      fi
    fi
    fi
  else
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m New login detected for admin user: $username from IP: $ip_address, but notifications are disabled by admin user. Skipping logging."
  fi
}

# ====== MYSQL SERVICE
mysql_docker_containers_status() {
      if docker --context=default ps --format "{{.Names}}" | grep -q "openpanel_mysql"; then
        if mysql -Ne "SELECT 'PONG' AS PING;" 2>/dev/null | grep -q "PONG"; then
        ((PASS++))
          echo -e "\e[32m[✔]\e[0m MySQL container is active and service running."
        else
          echo -e "\e[31m[✘]\e[0m MySQL container is running but not responding correctly, initiating restart.."
          cd /root && docker --context default compose up -d openpanel_mysql 2>/dev/null
        fi
      else
        ((FAIL++))
        STATUS=2
        echo -e "\e[31m[✘]\e[0m MySQL Docker container is not active, initiating restart.." #Writing notification to log file."
        cd /root && docker --context default compose up -d openpanel_mysql 2>/dev/null

        title="MySQL service is not active. Users are unable to log into OpenPanel!"
        message="MySQL container is running, but is not respond to queries. Sentinel failed to restart mysql and users are unable to login to OpenPanel, please check ASAP."
        sleep 5
        if mysql -Ne "SELECT 'PONG' AS PING;" 2>/dev/null | grep -q "PONG"; then
          ((FAIL--))
          STATUS=1
          echo "    Success: MySQL is now back online and responding to queries."
          title="MySQL service was restarted successfully!"
          message="MySQL container was running, but did not respond to queries. Sentinel restarted mysql and it is responding now."
        else
          echo "    Error: MySQL restared but still not responding to queries!"
        fi
        write_notification "$title" "$message"
      fi
}

# ====== ANY CONTAINER NAME
docker_containers_status() {
    service_name="$1"
    title="$2"

    check_status_after_restart(){
        if docker --context=default ps --format "{{.Names}}" | grep -wq "$service_name"; then
            echo -e "\e[32m[✔]\e[0m $service_name docker container successfully restarted."
            ((WARN--))
        else
            echo -e "\e[31m[✘]\e[0m $service_name docker container failed to restart."
            ((FAIL++))
            STATUS=2
            error_log=$(docker --context=default logs --tail 10 "$service_name" 2>&1 | awk '{gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s\\n", $0}')
            write_notification "$title" "$error_log"
        fi
    }


    check_caddy_health_after_restart() {
        local url="http://localhost/check"
        local status_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 1 --max-time 1 "$url")
      if [ "$status_code" -eq 200 ] || [ "$status_code" -eq 404 ]; then
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m $service_name is now up and running (status code: $status_code)."
      elif [ "$status_code" = "000" ]; then
        echo -e "\e[31m[✘]\e[0m Caddy service restart failed. The container appears to be running, but the service is unresponsive. Please review the Caddy container logs."
        ((WARN--))
        ((FAIL++))
        error_log=$(docker --context=default logs --tail 10 "$service_name" 2>&1 | awk '{gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s\\n", $0}')
        write_notification "$title" "$error_log"
      fi
    }


    check_caddy_health() {
        local url="http://localhost/check"
        local status_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 1 --max-time 1 "$url")
      if [ "$status_code" -eq 200 ] || [ "$status_code" -eq 404 ]; then
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m $service_name docker container is active (status code: $status_code)."
      else
        start_caddy
      fi
    }

    start_caddy() {
        ((WARN++))
        echo -e "\e[38;5;214m[!]\e[0m $service_name docker container is not running."
  
        echo "  - Checking if domains exist and if caddy service should be started..."
        if ls /etc/openpanel/caddy/domains > /dev/null 2>&1; then
            echo "  - Domains are hosted on this server, starting caddy service.."
            cd /root && docker --context=default compose up -d caddy > /dev/null 2>&1
            check_caddy_health_after_restart
        else
            ((WARN--))
            echo "  - No domains detected. Caddy is not yet needed."
        fi
      }

    if docker --context=default ps --format "{{.Names}}" | grep -wq "$service_name"; then
      if [ "$service_name" == "caddy" ]; then
        check_caddy_health
      else
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m $service_name docker container is active."
      fi
    else 
        if [ "$service_name" == "openpanel" ]; then
            echo "  - Checking if users exist and if openpanel service should be started..."
            users=$(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}')
            if [[ -z "$users" || "$users" == "No users." ]]; then
              ((WARN--))
              echo "  - No users found in the database."
            else
                echo "  - User accounts are hosted on this server, starting openpanel service.."
                cd /root && docker --context=default compose up -d openpanel > /dev/null 2>&1
                check_status_after_restart
            fi 
        elif [ "$service_name" == "openpanel_dns" ]; then
            echo "  - Checking if DNS zones exist and if BIND9 service should be started..."
            if ls /etc/bind/zones/*.zone > /dev/null 2>&1; then
                echo "  - DNS zones are hosted on this server, starting BIND9 service.."
                cd /root && docker --context=default compose up -d bind9 > /dev/null 2>&1
                check_status_after_restart
            else
                ((WARN--))
                echo "  - No domains detected. DNS is not yet needed."
            fi
        elif [ "$service_name" == "caddy" ]; then
            check_caddy_health
        else
            docker --context=default restart "$service_name"  > /dev/null 2>&1
            check_status_after_restart
        fi
    fi
}

# ====== CHECK ANY SERVICE
check_service_status() {
  local service_name="$1"
  local title="$2"

  if systemctl is-active --quiet "$service_name"; then
  ((PASS++))
    echo -e "\e[32m[✔]\e[0m $service_name is active."
  else

    if [[ "$service_name" == "admin" && ! -f /root/openadmin_is_disabled ]]; then
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m $service_name is disabled by Administrator."
    else

      ((FAIL++))
      STATUS=2
    echo -e "\e[31m[✘]\e[0m $service_name is not active."
    local error_log=""

      error_log=$(journalctl -n 5 -u "$service_name" 2>/dev/null | sed ':a;N;$!ba;s/\n/\\n/g')

      if [ -n "$error_log" ]; then
        write_notification "$title" "$error_log"
      else
        echo "no logs."
      fi

      echo -e "\e[33m[⚠️]\e[0m Restarting $service_name..."
      systemctl restart "$service_name"

      if systemctl is-active --quiet "$service_name"; then
        echo -e "\e[32m[✔]\e[0m $service_name has been restarted and is now active."
      else
        echo -e "\e[31m[✘]\e[0m Failed to restart $service_name."
      fi
      
    fi
  fi
}

# ====== GENERATE CRASH REPORT
generate_crashlog_report() {
  local crashlog_dir="/var/log/openpanel/admin/crashlog"
  local filename=$(date +%s)
  generated_report="${crashlog_dir}/${filename}.txt"
  local formatted_date=$(date '+%d %m %Y %H:%M')
  local hostname=$(hostname)
  local break_line="+---------------------------------------------------------------------------------------------------------------+"

  mkdir -p "$crashlog_dir" || return 1

  local current_load
  current_load=$(awk -F'load average:' '{print $2}' < <(uptime) | sed 's/^ *//')

  local top_mem
  top_mem=$(ps -eo user,pid,%cpu,%mem,command --sort=-%mem | head -n 11 | \
    awk '{ printf "  %-10s %-6s %-6s %-6s %-s\n", $1, $2, $3, $4, substr($0, index($0,$5), 60) }')

  local top_cpu
  top_cpu=$(ps -eo user,pid,%cpu,%mem,command --sort=-%cpu | head -n 11 | \
    awk '{ printf "  %-10s %-6s %-6s %-6s %-s\n", $1, $2, $3, $4, substr($0, index($0,$5), 60) }')

  local mysql
  mysql=$(mysql -e "SHOW PROCESSLIST" 2>/dev/null | grep -v "^Id" || echo "MySQL not available or permission denied.")

  local net_stat
  net_stat=$(ss -s 2>/dev/null || netstat -s 2>/dev/null || echo "No network stats available.")

  local swap
  swap=$(awk '/Name|VmSwap/ { printf "%s %s ", $2, $3 } /VmSwap/ {print ""}' /proc/*/status 2>/dev/null | \
      sort -k 2 -n -r | head -10)

  local io
  if command -v iotop &>/dev/null; then
    io=$(iotop -oqqqk | head -10)
  elif command -v iostat &>/dev/null; then
    io=$(iostat -xz 1 1 | head -n 10)
  else
    io="I/O tools not available."
  fi

  {
    printf "%s\n GENERAL INFO\n%s\n\n- Hostname: %s\n- Date: %s\n\n" "$break_line" "$break_line" "$hostname" "$formatted_date"
    printf "%s\n CURRENT LOAD\n%s\n\nLoad average:%s\n\n" "$break_line" "$break_line" "$current_load"
    printf "%s\n TOP PROCESSES BY MEMORY\n%s\n\n%s\n\n" "$break_line" "$break_line" "$top_mem"
    printf "%s\n TOP PROCESSES BY CPU\n%s\n\n%s\n\n" "$break_line" "$break_line" "$top_cpu"
    printf "%s\n MYSQL INFO\n%s\n\n%s\n\n" "$break_line" "$break_line" "$mysql"
    printf "%s\n I/O USAGE\n%s\n\n%s\n\n" "$break_line" "$break_line" "$io"
    printf "%s\n SWAP USAGE\n%s\n\n%s\n\n" "$break_line" "$break_line" "$swap"
    printf "%s\n NETWORK STATISTICS\n%s\n\n%s\n\n" "$break_line" "$break_line" "$net_stat"
    printf "%s\nThis report was generated by the Sentinel service on OpenPanel server \"%s\" at %s.\n\n" \
      "$break_line" "$hostname" "$formatted_date"
    printf "Report file saved to: %s\n%s\n" "$generated_report" "$break_line"
  } > "$generated_report"
}



# ====== CHECK LOAD
check_system_load() {
  local title="High System Load!"
  current_load=$(uptime | awk -F'average:' '{print $2}' | awk -F', ' '{print $1}')
  ensure_installed bc
  if (( $(echo "$current_load > $LOAD_THRESHOLD" | bc -l) )); then
      ((FAIL++))
      STATUS=2
      echo -e "\e[31m[✘]\e[0m Average Load usage ($current_load) is higher than threshold value ($LOAD_THRESHOLD). Generating crash-report and writing notification."
      generate_crashlog_report
      write_notification "$title" "Current load: $current_load | detailed report: $generated_report"
  else
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m Current Load usage $current_load is lower than the threshold value $LOAD_THRESHOLD. Skipping."
  fi
}



# ====== CHECK SWAP
check_swap_usage() {
    local title="High SWAP usage!"
    free_output=$(free -t)
    swap_used=$(echo "$free_output" | awk 'FNR == 3 {print $3}')
    swap_total=$(echo "$free_output" | awk 'FNR == 3 {print $2}')
    if [ "$swap_total" -gt 0 ]; then
        SWAP_USAGE=$(awk "BEGIN {print ($swap_used / $swap_total) * 100}")
    else
        SWAP_USAGE=0
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m Total SWAP is $SWAP_USAGE (skipping swap check for ${SWAP_THRESHOLD}% treshold)"
        return
    fi

    SWAP_USAGE_NO_DECIMALS=$(printf %.0f $SWAP_USAGE)
    if [ "$SWAP_USAGE_NO_DECIMALS" -gt "$SWAP_THRESHOLD" ]; then
      if [ -f "$LOCK_FILE" ]; then
          file_age=$(($(date +%s) - $(date -r "$LOCK_FILE" +%s)))
          if [ "$file_age" -gt 21600 ]; then
              echo "Lock file is older than 6 hours. Deleting..."
              rm -f "$LOCK_FILE"
          else
              ((WARN++))
              echo -e "\e[38;5;214m[!]\e[0m Previous SWAP cleanup is still in progress. Skipping the current run."
              exit 0
          fi
      fi

        echo "Current SWAP usage ($SWAP_USAGE_NO_DECIMALS) is higher than treshold value ($SWAP_THRESHOLD). Writing notification."        
        write_notification "$title" "Current SWAP usage: $SWAP_USAGE_NO_DECIMALS Starting the cleanup process now... you will get a new notification once everything is completed..."
        touch "$LOCK_FILE" # create when we start
        
        echo 3 >/proc/sys/vm/drop_caches
        swapoff -a
        swapon -a

        swap_usage=$(free -t | awk 'FNR == 3 {print $3/$2*100}')
        swap_usage_no_decimals=$(printf %.0f $SWAP_USAGE)
        local title_ok="SWAP is cleared - Current value: $swap_usage_no_decimals"
        local title_not_ok="URGENT! SWAP could not be cleared on $HOSTNAME"
        if [ "$swap_usage_no_decimals" -lt "$SWAP_THRESHOLD" ]; then
            echo -e "The Sentinel service has completed clearing SWAP on server $HOSTNAME! \n THANK YOU FOR USING THIS SERVICE! PLEASE REPORT ALL BUGS AND ERRORS TO sentinel@openpanel.co"
            write_notification "$title_ok" "The Sentinel service has completed clearing SWAP on server $HOSTNAME!"
            echo -e "SWAP Usage was abnormal, healing completed, notification sent! This SWAP check was performed at: $TIME"
            rm -f "$LOCK_FILE" # delete after success
        else
            ((FAIL++))
            STATUS=2
            echo -e "\e[31m[✘]\e[0m URGENT! SWAP could not be cleared on $HOSTNAME"
            write_notification "$title_not_ok" "The Sentinel service detected abnormal SWAP usage at $TIME and tried clearing the space but SWAP usage is still above the $swap_usage_no_decimals treshold."
        fi
    else
    ((PASS++))
        echo -e "\e[32m[✔]\e[0m Current SWAP usage is $SWAP_USAGE_NO_DECIMALS (bellow the ${SWAP_THRESHOLD}% treshold) - SWAP check was performed at: $TIME "
        rm -f "$LOCK_FILE" # delete if failed but on next run it is ok
    fi
}


# ====== CHECK RAM
check_ram_usage() {
  local title="High Memory Usage!"
  local total_ram=$(free -m | awk '/^Mem:/{print $2}')
  local used_ram=$(free -m | awk '/^Mem:/{print $3}')
  local ram_percentage=$((used_ram * 100 / total_ram))
  
  local message="Used RAM: $used_ram MB, Total RAM: $total_ram MB, Usage: $ram_percentage%"
  local message_to_check_in_file="Used RAM"

  if is_unread_message_present "$message_to_check_in_file"; then
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Unread RAM usage notification already exists. Skipping."
    return
  fi

  if [ "$ram_percentage" -gt "$RAM_THRESHOLD" ]; then
              ((FAIL++))
            STATUS=2

    echo -e "\e[31m[✘]\e[0m RAM usage ($ram_percentage) is higher than treshold value ($RAM_THRESHOLD). Writing notification."
    write_notification "$title" "$message"
  else
  ((PASS++))
    echo -e "\e[32m[✔]\e[0m Current RAM usage $ram_percentage is lower than the treshold value $RAM_THRESHOLD. Skipping."
  fi
}

# ====== CHECK CPU
function check_cpu_usage() {
  local title="High CPU Usage!"

  local cpu_percentage=$(top -bn1 | awk '/^%Cpu/{print $2}' | awk -F'.' '{print $1}')
  
if [ "$cpu_percentage" -gt "$CPU_THRESHOLD" ]; then
            ((FAIL++))
            STATUS=2
  echo -e "\e[31m[✘]\e[0m CPU usage ($cpu_percentage) is higher than treshold ($CPU_THRESHOLD). Writing notification."
  top_processes=$(ps aux --sort -%cpu | head -10 | sed ':a;N;$!ba;s/\n/\\n/g')
  write_notification "$title" "CPU Usage: $cpu_percentage% | Top Processes: $top_processes"
else
((PASS++))
  echo -e "\e[32m[✔]\e[0m Current CPU usage $cpu_percentage is lower than the treshold value $CPU_THRESHOLD. Skipping."
fi
}

# ====== CHECK DISK USAGE
function check_disk_usage() {
  local title="Running out of Disk Space!"

  local disk_percentage=$(df -h --output=pcent / | tail -n 1 | tr -d '%')
  if [ "$disk_percentage" -gt "$DISK_THRESHOLD" ]; then

    if is_unread_message_present "$title"; then
      ((WARN++))
      echo -e "\e[38;5;214m[!]\e[0m Unread DU notification already exists. Skipping."
      return
    fi
              ((FAIL++))
            STATUS=2
    echo -e "\e[31m[✘]\e[0m Disk usage ($disk_percentage) is higher than the treshold value $DISK_THRESHOLD. Writing notification."
    disk_partitions_usage=$(df -h | sort -r -k 5 -i | sed ':a;N;$!ba;s/\n/\\n/g')
    write_notification "$title" "Disk Usage: $disk_percentage% | Partitions: $disk_partitions_usage"
  else
  ((PASS++))
  echo -e "\e[32m[✔]\e[0m Current Disk usage $disk_percentage is lower than the treshold value $DISK_THRESHOLD. Skipping."
  fi
}

# ====== CHECK DNS
check_if_panel_domain_and_ns_resolve_to_server (){

  if [ "$MAIN_DOMAIN_AND_NS" != "no" ]; then
      RESULT=$(opencli domain)
      
      if [[ "$RESULT" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
          FORCED_DOMAIN=$RESULT
          CHECK_DOMAIN_ALSO="yes"
      else
          CHECK_DOMAIN_ALSO="no"
      fi
    
      NS1_SET=$(awk -F'=' '/^ns1/ {print $2}' "$CONF_FILE")
      
      if [ -n "$NS1_SET" ]; then
          if [[ "$NS1_SET" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
              NS1=$NS1_SET
              CHECK_NS_ALSO="yes"
              NS2_SET=$(awk -F'=' '/^ns2/ {print $2}' "$CONF_FILE")
              NS2=$NS2_SET
              NS3_SET=$(awk -F'=' '/^ns3/ {print $2}' "$CONF_FILE")
              NS3=$NS3_SET
              NS4_SET=$(awk -F'=' '/^ns4/ {print $2}' "$CONF_FILE")
              NS4=$NS4_SET
          else
              echo "NS1 '$NS1_SET' is not a valid domain."
              CHECK_NS_ALSO="no"
          fi
      else
          CHECK_NS_ALSO="no"
      fi
    
    if [ "$CHECK_DOMAIN_ALSO" == "no" ] && [ "$CHECK_NS_ALSO" == "no" ]; then
        ((WARN++))
        echo -e "\e[38;5;214m[!]\e[0m Missing or invalid custom domain and nameservers, skipping DNS checks.."
    else
    
      GOOGLE_DNS_SERVER="8.8.8.8"
      SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
      if [ -f "$SCRIPT_PATH" ]; then
          source "$SCRIPT_PATH"
      else
          IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
      fi
    
      SERVER_IP=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 --tries=1 -4 --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
    
      if [ -z "$SERVER_IP" ]; then
          SERVER_IP=$(ip addr | grep 'inet ' | grep global | head -n1 | awk '{print $2}' | cut -f1 -d/)
      fi
    
      if [ "$CHECK_DOMAIN_ALSO" == "yes" ] || [ "$CHECK_NS_ALSO" == "yes" ]; then
          if [ "$CHECK_DOMAIN_ALSO" == "no" ]; then
              ((WARN++))
              echo -e "\e[38;5;214m[!]\e[0m Domain is not set for accessing OpenPanel, skip checking DNS."
          else
              if [ -n "$FORCED_DOMAIN" ]; then
                  ensure_installed dig bind-utils
                  domain_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$FORCED_DOMAIN")
                  
                  if [ "$domain_ip" == "$SERVER_IP" ]; then
                  ((PASS++))
                      echo -e "\e[32m[✔]\e[0m $FORCED_DOMAIN resolves to $SERVER_IP"
                  else
                      ns_records=$(dig +short @"$GOOGLE_DNS_SERVER" NS "$FORCED_DOMAIN")
    
                      if echo "$ns_records" | grep -q 'cloudflare'; then
                         ((WARN++))
                         echo -e "\e[38;5;214m[!]\e[0m $FORCED_DOMAIN does not resolve to $SERVER_IP, but it is using Cloudflare DNS, so it is possible that proxy option is enabled. Skipping."
                      else
                          ((FAIL++))
                STATUS=2
                          echo -e "\e[31m[✘]\e[0m $FORCED_DOMAIN does not resolve to $SERVER_IP"
                          local title="$FORCED_DOMAIN does not resolve to $SERVER_IP"
                          local message="$FORCED_DOMAIN is set as a domain but does not resolve to the server IP $SERVER_IP. This may cause accessibility issues."
                          write_notification "$title" "$message"
                      fi
                  fi
              fi
          fi
    
          if [ "$CHECK_NS_ALSO" == "no" ]; then
          ((PASS++))
              echo -e "\e[32m[✔]\e[0m Skip checking nameservers as they are not set."
          else
              local_fail=0
              if [ -n "$NS1" ] && [ -n "$NS2" ]; then
                  ns1_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS1")
                  ns2_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS2")
                  [ -n "$NS3" ] && ns3_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS3")
                  [ -n "$NS4" ] && ns4_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS4")
                  
                  all_server_ips=$(hostname -I)
                  failed_nameservers=()
                            
                  for ns in NS1 NS2; do
                      ip_var="${ns,,}_ip"
                      ip_val="${!ip_var}"
                      if ! echo "$all_server_ips" | grep -qw "$ip_val"; then
                          ((local_fail++))
                          failed_nameservers+=("$ns resolves to $ip_val (expected one of: $all_server_ips)")
                      fi
                  done
                  # optional
                  for ns in NS3 NS4; do
                      if [ -n "${!ns}" ]; then
                          ip_var="${ns,,}_ip"
                          ip_val="${!ip_var}"
                          if ! echo "$all_server_ips" | grep -qw "$ip_val"; then
                              ((local_fail++))
                              failed_nameservers+=("$ns resolves to $ip_val (expected one of: $all_server_ips)")
                          fi
                      fi
                  done
    
                  if [ $local_fail -eq 0 ]; then
                      ((PASS++))
                      echo -e "\e[32m[✔]\e[0m All configured nameservers resolve to local IPs ($(hostname -I))"
                  else
                      STATUS=2
                      echo -e "\e[31m[✘]\e[0m Nameservers do not resolve correctly:"
                      for fail_msg in "${failed_nameservers[@]}"; do
                          echo "    $fail_msg"
                      done
                      local title="Configured nameservers do not resolve to local IPs"
                      local message
                      message=$(IFS=' | '; echo "${failed_nameservers[*]}")
                      write_notification "$title" "$message"
                      ((FAIL++))
                  fi
              else 
                ((WARN++))
                echo -e "\e[38;5;214m[!]\e[0m Only one nameserver is currently set. Please add at least two nameservers to ensure DNS redundancy."
              fi
          fi
      fi
    fi
    
  else
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Checking panel domain and nameservers (dns) is explicitly set to 'no' in the INI file. Skipping logging.."
  fi

}

# ====== DAILY REPORT
email_daily_report() {
  local title="Daily Usage Report"
  local message="Daily Usage Report"
  if [ "$EMAIL_ALERT" != "no" ]; then
    echo "Generating daily usage report.."
    email_notification "$title" "$message"
  else
      echo "Email alerts are disabled - daily usage report will not be generated."
  fi
}






# ======================================================================
# Got flags?
: '
script can be run:

  opencli sentinel --reboot     - sends reboot alert 
  opencli sentinel              - checks what admin configured
  opencli sentinel --report     - generates daily usage report
'

if [ "$1" == "--startup" ]; then
  perform_startup_action
  exit 0
elif [ "$1" == "--report" ]; then
  email_daily_report
  exit 0
else
  print_header
  check_for_debug_and_print_info
  echo "Checking health for monitored services:"
  echo ""
  
  check_services() {
    declare -A service_checks=(
      [caddy]="docker_containers_status 'caddy' 'Caddy is not active. Users websites are not working!'"
      [csf]="check_service_status 'csf' 'ConfigService Firewall (CSF) is not active. Server and websites are not protected!'"
      [admin]="check_service_status 'admin' 'Admin service is not active. OpenAdmin service is not accessible!'"
      [docker]="check_service_status 'docker' 'Docker service is not active. User websites are down!'"
      [panel]="docker_containers_status 'openpanel' 'OpenPanel docker container is not running. Users are unable to access the OpenPanel interface!'"
      [mysql]="mysql_docker_containers_status"
      [named]="docker_containers_status 'openpanel_dns' 'Named (BIND9) service is not active. DNS resolving of domains is not working!'"
    )
  
    for svc in "${!service_checks[@]}"; do
      if echo "$SERVICES" | grep -q "$svc"; then
        eval "${service_checks[$svc]}"
      fi
    done
  }

  
  start_login_section() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo "Checking SSH and OpenAdmin logins:"
    echo ""
  }
  
  
  check_resources_section(){
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo "Checking server resource usage:"
    echo ""
  }
  
  
  check_dns_section() {
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo "Checking DNS:"
    echo ""
  }

  
  summary(){
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
  
    if [[ $STATUS == 0 ]]; then
        echo -en "\e[32mAll Tests Passed!\e[0m\n"
    elif [[ $STATUS == 1 ]]; then
        echo -en "\e[93mSome non-critical tests failed.  Please review these items.\e[0m\e[0m\n"
    else
        echo -en "\e[41mOne or more tests failed.  Please review these items.\e[0m\n"
    fi
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    echo -en "\e[1m${PASS} Tests PASSED\e[0m\n"
    echo -en "\e[1m${WARN} WARNINGS\e[0m\n"
    echo -en "\e[1m${FAIL} Tests FAILED\e[0m\n"
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
    show_execution_time
  }


  
  
  # Progress bar script
  PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
  PROGRESS_BAR_FILE="progress_bar.sh"
  wget --timeout=5 --tries=3 -4 "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1
  if [ ! -f "$PROGRESS_BAR_FILE" ]; then
      echo "Failed to download progress_bar.sh"
      exit 1
  fi
    source "$PROGRESS_BAR_FILE"
    
    FUNCTIONS=(
      check_services
      start_login_section
      check_new_logins
      check_ssh_logins
      check_resources_section
      check_disk_usage
      check_system_load
      check_ram_usage
      check_cpu_usage
      check_swap_usage
      check_dns_section
      check_if_panel_domain_and_ns_resolve_to_server
      summary
    )

  
    TOTAL_STEPS=${#FUNCTIONS[@]}
    CURRENT_STEP=0
    
    update_progress() {
        CURRENT_STEP=$((CURRENT_STEP + 1))
        PERCENTAGE=$(($CURRENT_STEP * 100 / $TOTAL_STEPS))
        draw_progress_bar $PERCENTAGE
    }
  
    main() {
        enable_trapping
        setup_scroll_area
        for func in "${FUNCTIONS[@]}"
        do
            $func
            update_progress
        done
        destroy_scroll_area
    }
  
    main
  fi

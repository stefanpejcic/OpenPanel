#!/bin/bash
################################################################################
# Script Name: report.sh
# Description: Generate a system report and send it to OpenPanel support team.
# Usage: opencli report
#        opencli report [--public] [--cli] [--csf]
# Author: Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 20.03.2026
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

# ======================================================================
# Constants
GREEN='\033[0;32m'
RESET='\033[0m'

# ======================================================================
# Variables
cli_flag=false
csf_flag=false
upload_flag=false
non_interactive=false


# ======================================================================
# Helpers

create_local_path() {
	output_dir="/var/log/openpanel/admin/reports"
	mkdir -p "$output_dir"
	output_file="$output_dir/system_info_$(date +'%Y%m%d%H%M%S').txt"
	export output_file
}

parse_args() {
	while [[ $# -gt 0 ]]; do
	    case $1 in
	        --non-interactive)
	            non_interactive=true
	            ;;
	        --cli)
	            cli_flag=true
	            ;;
	        --csf)
	            csf_flag=true
	            ;; 
	        --public|--link|--upload)
	            upload_flag=true
	            ;; 
	        *)
	            echo "Unknown option: $1"
	            exit 1
	            ;;
	    esac
	    shift
	done
}


# ======================================================================
# Helpers

run_command() {
  local cmd="$1"
  local label="$2"
  local tmpfile="$3"
  {
    echo "# $label:"
    eval "$cmd" 2>&1
    echo "# ======================================================================"
    echo
  } >> "$tmpfile"
}

# Each collect_* function writes to its own temp file, then we merge them in order.

collect_os_info() {
  local tmp="$1"
  local os_info
  os_info=$(awk -F= '/^(NAME|VERSION_ID)/{gsub(/"/, "", $2); printf("%s ", $2)}' /etc/os-release)
  run_command "echo $os_info"  "Checking OS distribution"          "$tmp"
  run_command "uptime"         "Checking server uptime"            "$tmp"
  run_command "free -h"        "Collecting physical memory and swap usage" "$tmp"
  run_command "df -h"          "Collecting disk information"       "$tmp"
}

collect_opencli_info() {
  local tmp="$1"
  run_command "opencli --version" "Listing OpenPanel version" "$tmp"
}

collect_mysql_info() {
  local tmp="$1"
  run_command "mysql --protocol=tcp --version" "Checking MySQL Version" "$tmp"
}

collect_admin_info() {
  local tmp="$1"
  run_command "/usr/local/admin/venv/bin/python3 --version" "Checking Python version for OpenAdmin venv" "$tmp"
}

collect_docker_info() {
  local tmp="$1"
  run_command "docker info" "Collecting host docker information" "$tmp"
}

collect_opencli_commands() {
  local tmp="$1"
  if [ "$cli_flag" = true ]; then
    echo "=== OpenCLI Information ===" >> "$tmp"
    run_command "opencli commands" "Checking available OpenCLI Commands" "$tmp"
  fi
}

collect_csf_rules() {
  local tmp="$1"
  if [ "$csf_flag" = true ]; then
    echo "=== Sentinel Firewall Rules ===" >> "$tmp"
    run_command "csf -l" "Collecting Firewall Rules" "$tmp"
  fi
}

collect_openpanel_settings() {
  local tmp="$1"
  echo "=== OpenPanel Settings ===" >> "$tmp"
  run_command "cat /etc/openpanel/openpanel/conf/openpanel.config" "Listing OpenPanel configuration file" "$tmp"
  run_command "cat /etc/openpanel/caddy/Caddyfile" "Listing Caddyfile" "$tmp"

}

collect_openadmin_settings() {
  local tmp="$1"
  echo "=== OpenAdmin Service ===" >> "$tmp"
  run_command "cat /etc/openpanel/openadmin/config/admin.ini"        "Listing OpenAdmin configuration file"           "$tmp"
  run_command "tail -30 /var/log/openpanel/admin/error.log"          "Checking OpenAdmin log for errors"               "$tmp"
}

collect_mysql_information() {
  local tmp="$1"
  echo "=== MySQL Information ===" >> "$tmp"
  run_command "docker logs --tail 30 openpanel_mysql"  "Checking MySQL service for errors"   "$tmp"
  run_command "cat /etc/openpanel/mysql/*_my.cnf"      "Viewing MySQL login information"      "$tmp"
}

collect_services_status() {
  local tmp="$1"
  echo "=== Services Status ===" >> "$tmp"
  run_command "docker --context=default compose ls"       "Listing OpenPanel Stack"                      "$tmp"
  run_command "docker --context=default ps -a"           "Checking system containers status"            "$tmp"
  run_command "systemctl status admin"                   "Checking status of OpenAdmin service"         "$tmp"
  run_command "systemctl status docker"                  "Checking status of Docker service"            "$tmp"
  run_command "systemctl status csf"                     "Checking if Sentinel Firewall (CSF) is running" "$tmp"
}

collect_user_services() {
  local tmp="$1"
  echo "=== Docker Context Services ===" >> "$tmp"
  for dir in /home/*; do
      local file="$dir/docker-compose.yml"
      local user
      user=$(basename "$dir")
      if [[ -f "$file" ]]; then
        run_command "echo '- User: $user' && echo '' && docker --context=$user compose -f $dir/docker-compose.yml config --services" \
          "Listing services for user: $user" "$tmp"
      fi
  done
}


# ======================================================================
# Parallel runner
#
# Functions are split into two groups:
#   PARALLEL_FUNCS  – independent, run concurrently via xargs
#   SERIAL_FUNCS    – must run after parallel phase (e.g. need flags set, or
#                     iterate over dynamic data like /home/*)
#
# Each function writes to its own numbered temp file so output order is
# deterministic regardless of which job finishes first. The temp files are
# merged in order at the end.

export output_file cli_flag csf_flag non_interactive

# Ordered list – index determines merge order
ORDERED_FUNCS=(
  collect_os_info
  collect_opencli_info
  collect_mysql_info
  collect_admin_info
  collect_docker_info
  collect_opencli_commands
  collect_csf_rules
  collect_openpanel_settings
  collect_openadmin_settings
  collect_mysql_information
  collect_services_status
  collect_user_services
)

TMPDIR_REPORT=$(mktemp -d)
export TMPDIR_REPORT

# Export all collect_* functions and run_command so xargs subshells can see them
export -f run_command \
           collect_os_info \
           collect_opencli_info \
           collect_mysql_info \
           collect_admin_info \
           collect_docker_info \
           collect_opencli_commands \
           collect_csf_rules \
           collect_openpanel_settings \
           collect_openadmin_settings \
           collect_mysql_information \
           collect_services_status \
           collect_user_services

# Wrapper called by xargs: receives "<index> <func_name>"
run_indexed() {
  local idx="$1"
  local func="$2"
  local tmpfile
  tmpfile=$(printf "%s/%04d.txt" "$TMPDIR_REPORT" "$idx")
  $func "$tmpfile"
}
export -f run_indexed

main() {
  # Build index:funcname pairs for xargs
  local pairs=()
  local i=0
  for func in "${ORDERED_FUNCS[@]}"; do
    pairs+=("$i $func")
    (( i++ ))
  done

  if [ "$non_interactive" = false ]; then
    echo "Collecting system information (parallel)..."
  fi

  # Run all collectors in parallel (up to nproc jobs at once)
  printf '%s\n' "${pairs[@]}" | \
    xargs -P "$(nproc)" -I '{}' bash -c 'run_indexed $@' _ {}

  # Merge temp files in order into the final report
  for tmpfile in $(ls "$TMPDIR_REPORT"/*.txt 2>/dev/null | sort); do
    cat "$tmpfile" >> "$output_file"
  done

  rm -rf "$TMPDIR_REPORT"

  upload_report
}


upload_report() {
	if [ ! -f "$output_file" ]; then
	  echo "Information not collected! report file does not exist: $output_file"
   	else
	  if [ "$non_interactive" = false ]; then
	  	clear
	  fi
	  if [ "$upload_flag" = true ]; then
	    response=$(curl -F "file=@$output_file" https://support.openpanel.org/opencli_server_info.php 2>/dev/null)
	    if echo "$response" | grep -q "File upload failed."; then
	      echo ""
	      echo -e "Information collected successfully but uploading to support.openpanel.org failed. Please provide content from the following file to the support team:\n$output_file"
	    elif echo "$response" | grep -q "name="; then
	      FILE_NAME=$(echo "$response" | cut -d'=' -f2)
       		if [ "$non_interactive" = true ]; then
		      echo -e "Information collected successfully. Please provide the following key to the support team:\n${FILE_NAME}"
       		else
		      echo -e "Information collected successfully. Please provide the following key to the support team:\n${GREEN}${FILE_NAME}${RESET}"
  		fi
	    else
	      echo -e "Unexpected upload response:\n$response"
	      echo -e "Please send the content of the following file manually:\n$output_file"
	    fi
	  else
	    echo -e "Information collected successfully. Please provide content of the following file to the support team:\n$output_file"
	  fi
	fi
}


# ======================================================================
# flock
(
flock -n 200 || { echo "Error: Another instance of the report script is already running. Exiting."; exit 1; }
create_local_path
parse_args "$@"
main
)200>/root/openpanel_install.lock

exit 0

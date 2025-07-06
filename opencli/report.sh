#!/bin/bash
################################################################################
# Script Name: report.sh
# Description: Generate a system report and send it to OpenPanel support team.
# Usage: opencli report
#        opencli report [--public] [--cli] [--csf]
# Author: Stefan Pejcic
# Created: 07.10.2023
# Last Modified: 04.07.2025
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


# Create directory if it doesn't exist
output_dir="/var/log/openpanel/admin/reports"
mkdir -p "$output_dir"
output_file="$output_dir/system_info_$(date +'%Y%m%d%H%M%S').txt"

GREEN='\033[0;32m'
RESET='\033[0m'





setup_progress_bar_script(){
	# Progress bar script
	PROGRESS_BAR_URL="https://raw.githubusercontent.com/pollev/bash_progress_bar/master/progress_bar.sh"
	PROGRESS_BAR_FILE="progress_bar.sh"

	# Check if wget is available
	if command -v wget &> /dev/null; then
	    wget --timeout=5 --inet4-only "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        echo "ERROR: wget failed or timed out after 5 seconds while downloading from github"
	        exit 1
	    fi
	# If wget is not available, check if curl is available *(fallback for fedora)
	elif command -v curl -4 &> /dev/null; then
	    curl -4 --max-time 5 -s "$PROGRESS_BAR_URL" -o "$PROGRESS_BAR_FILE" > /dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        echo "ERROR: curl failed or timed out after 5 seconds while downloading progress_bar.sh"
	        exit 1
	    fi
	else
	    echo "Neither wget nor curl is available. Please install one of them to proceed."
	    exit 1
	fi

	if [ ! -f "$PROGRESS_BAR_FILE" ]; then
	    echo "ERROR: Failed to download progress_bar.sh - Github may be unreachable from your server: $PROGRESS_BAR_URL"
	    exit 1
	fi
}







setup_progress_bar_script
source "$PROGRESS_BAR_FILE"               # Source the progress bar script

FUNCTIONS=(
get_os_info
get_opencli_info
get_mysql_info
get_admin_info
get_docker_info
run_opencli # Run OpenCLI commands if --cli flag is provided
run_csf_rules
display_openpanel_settings # Display OpenPanel settings
display_openadmin_settings # Display OpenAdmin settings
display_mysql_information # Display MySQL information
check_services_status # Check the status of services
list_user_services # list users services
upload_report
)


TOTAL_STEPS=${#FUNCTIONS[@]}
CURRENT_STEP=0

update_progress() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    PERCENTAGE=$(($CURRENT_STEP * 100 / $TOTAL_STEPS))
    draw_progress_bar $PERCENTAGE
}

main() {
	if [ "$non_interactive" = true ]; then
		for func in "${FUNCTIONS[@]}"; do
		  $func
		done
	else
	    enable_trapping                       # clean on CTRL+C
	    setup_scroll_area                     # load progress bar
	    for func in "${FUNCTIONS[@]}"
	    do
	        $func                             # Execute each function
	        update_progress                   # update progress after each
	    done
	    destroy_scroll_area
	fi
}












# Function to run a command and print its output with a custom message
run_command() {
  echo "# $2:" >> "$output_file"
  if [ "$non_interactive" = false ]; then
  	echo "- $2"
  fi
  eval "$1" >> "$output_file" 2>&1
  echo >> "$output_file"
}


# HELPERS

run_opencli() {
  if [ "$cli_flag" = true ]; then
    echo "=== OpenCLI Information ===" >> "$output_file"
    run_command "opencli commands" "Checking available OpenCLI Commands"
  fi
}


run_csf_rules() {
  if [ "$csf_flag" = true ]; then
    echo "=== Firewall Rules ===" >> "$output_file"
    run_command "csf -l" "Collecting Firewall Rules"
  fi
}


# Function to check the status of services
check_services_status() {
  echo "=== Services Status ===" >> "$output_file"
  run_command "docker --context=default compose ls" "Listing OpenPanel Stack"
  run_command "systemctl status admin" "Checking status of OpenAdmin service"
  run_command "systemctl status docker" "Checking status of Docker service"
  run_command "systemctl status csf" "Checking if CSF is running"
}

# Function to display OpenPanel settings
display_openpanel_settings() {
  echo "=== OpenPanel Settings ===" >> "$output_file"
  run_command "cat /etc/openpanel/openpanel/conf/openpanel.config" "Listing OpenPanel configuration file"
}

# admin in 0.2.3
display_openadmin_settings() {
  echo "=== OpenAdmin Service ===" >> "$output_file"
  run_command "cat /etc/openpanel/openadmin/config/admin.ini" "Listing OpenAdmin configuration file"
  run_command "/usr/local/admin/venv/bin/python3 -m pip list" "Listing PIP packages in OpenAdmin venv"
  run_command "tail -30 /var/log/openpanel/admin/error.log" "Checking OpenAdmin log for errors"
}


# Function to display MySQL information
display_mysql_information() {
  echo "=== MySQL Information ===" >> "$output_file"
  run_command "docker logs --tail 30 openpanel_mysql" "Checking MySQL service for errors"
  run_command "cat /etc/openpanel/mysql/*_my.cnf" "Viewing MySQL login information"
}

# added in 1.2.2
list_user_services() {
  echo "=== Docker Context Services ===" >> "$output_file"
  for dir in /home/*; do
      file="$dir/docker-compose.yml"
      user=$(basename "$dir")
      if [[ -f "$file" ]]; then
        run_command "echo '- User: $user' && echo '' && docker --context=$user compose -f  $dir/docker-compose.yml config --services" "Listing services for user: $user"
      else
        run_command "echo 'Most likely not an OpenPanel user'" "Skipping services for folder: $user"
      fi
  done
}


get_os_info() {
  os_info=$(awk -F= '/^(NAME|VERSION_ID)/{gsub(/"/, "", $2); printf("%s ", $2)}' /etc/os-release)
  run_command "echo $os_info" "Checking OS distribution"
  run_command "uptime" "Checking server uptime"
  run_command "free -h" "Collecting physical memory and swap usage"
  run_command "df -h" "Collecting disk information"
}

get_opencli_info() {
  run_command "opencli --version" "Listing OpenPanel version"
}

get_mysql_info() {
  run_command "mysql --protocol=tcp --version" "Checking MySQL Version"
}

get_admin_info() {
  run_command "/usr/local/admin/venv/bin/python3 --version" "Checking Python version for OpenAdmin venv"
}

get_docker_info() {
  run_command "docker info" "Collecting host docker information"
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




# Default values
cli_flag=false
csf_flag=false
upload_flag=false
non_interactive=false


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



(
flock -n 200 || { echo "Error: Another instance of the report script is already running. Exiting."; exit 1; }
parse_args "$@"
main
)200>/root/openpanel_install.lock

exit 0

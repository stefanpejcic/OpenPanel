#!/bin/bash
VERSION="20.250.801"
# Record the process ID of the script
PID=$$

# Script should be run with SUDO
if [ "$EUID" -ne 0 ]
  then echo "[Error] - This script must be run with sudo or as the root user."
  exit 1
fi


DEBUG=false  # Default value for DEBUG
if [ "$1" = "--debug" ]; then
    DEBUG=true
fi





# counters
STATUS=0
PASS=0
WARN=0
FAIL=0

# notifications conf file
CONF_FILE="/etc/openpanel/openpanel/conf/openpanel.config"
# swap lock file to not repeat cleanup
LOCK_FILE="/tmp/swap_cleanup.lock"
TIME=$(date +%s%3N)
INI_FILE="/etc/openpanel/openadmin/config/notifications.ini"
HOSTNAME=$(hostname)
LOG_FILE="/var/log/openpanel/admin/notifications.log"
LOG_DIR=$(dirname "$LOG_FILE")

# Function to create folder and file if they do not exist
create_folder_and_file() {
  if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
  fi

  if [ ! -f "$LOG_FILE" ]; then
    touch "$LOG_FILE"
  fi
}


create_folder_and_file

show_execution_time() {
    end_time=$(date +%s%3N)
    execution_time=$((end_time - TIME))
    seconds=$((execution_time / 1000))
    milliseconds=$((execution_time % 1000))
    memory_usage=$(ps -p $$ -o rss=)
    echo "Elapsed time: ${seconds}.${milliseconds} seconds"
    echo "Memory usage: ${memory_usage} KB"
}

# Trap the EXIT signal to call show_execution_time function
trap show_execution_time EXIT
# TODO: make it work with progress script


# TODO: Check if running under cron




# Check if the INI file exists
if [ ! -f "$INI_FILE" ]; then
    echo "Error: INI file not found: $INI_FILE"
    exit 1
fi


# helper function to generate random token
generate_random_token() {
    tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 64
}

generate_random_token_one_time_only() {
    TOKEN_ONE_TIME="$(generate_random_token)"
    local new_value="mail_security_token=$TOKEN_ONE_TIME"
    # Use sed to replace the line in the file
    sed -i "s|^mail_security_token=.*$|$new_value|" "$CONF_FILE"
}

# Function to check if a value is a number between 1 and 100
is_valid_number() {
  local value="$1"
  [[ "$value" =~ ^[1-9][0-9]?$|^100$ ]]
}

# Extract email address from the configuration file
EMAIL_ALERT=$(awk -F'=' '/^email/ {print $2}' "$CONF_FILE")

# If email address is found, set EMAIL_ALERT to "yes" and set EMAIL to that address
if [ -n "$EMAIL_ALERT" ]; then
    EMAIL=$EMAIL_ALERT
    EMAIL_ALERT=yes
else
    # If no email address is found, set EMAIL_ALERT to "no" by default
    EMAIL_ALERT=no
fi



# Read values from the INI file or set fallback values

REBOOT=$(awk -F'=' '/^reboot/ {print $2}' "$INI_FILE")
REBOOT=${REBOOT:-yes}
[[ "$REBOOT" =~ ^(yes|no)$ ]] || REBOOT=yes

LOGIN=$(awk -F'=' '/^login/ {print $2}' "$INI_FILE")
LOGIN=${LOGIN:-yes}
[[ "$LOGIN" =~ ^(yes|no)$ ]] || LOGIN=yes

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





# Function to get the last message content from the log file
get_last_message_content() {
  tail -n 1 "$LOG_FILE" 2>/dev/null
}

# Function to check if an unread message with the same content exists in the log file
is_unread_message_present() {
  local unread_message_content="$1"
  grep -q "UNREAD.*$unread_message_content" "$LOG_FILE" && return 0 || return 1
}

# Send an email alert
email_notification() {
  local title="$1"
  local message="$2"


  #set random token
  generate_random_token_one_time_only

  # use the token
  TRANSIENT=$(awk -F'=' '/^mail_security_token/ {print $2}' "$CONF_FILE")

  # FOR DEBUG ONLY OTHERWISE IT CAN BE SEEN IN LOGS!!!!!!!
  #
  # echo $TRANSIENT
  #
  # curl -k -X POST   https://127.0.0.1:2087/send_email  -F 'transient=z3t5LPt4HirqpmW1KHbZdEXtgNR4Rl4bIw6xv4irUZIxXkIXZ8SJHjduOhjvDEe8' -F 'recipient=stefan@pejcic.rs' -F 'subject=proba sa servera' -F 'body=Da li je dosao mejl? Hvala.'


# Check for SSL
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




# Check if DEBUG is true before printing debug messages


# helper: make sure dig is available
ensure_dig_installed() {
    if ! command -v dig &> /dev/null; then
        # Detect the package manager and install bind-utils
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq bind-utils > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q bind-utils > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q bind-utils > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install dig command (bind-utils) manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v dig &> /dev/null; then
            echo "Error: dig command installation failed. Please install bind-utils manually and try again."
            exit 1
        fi
    fi
}

ensure_bc_installed() {
    if ! command -v bc &> /dev/null; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get update > /dev/null 2>&1
            sudo apt-get install -y -qq bc > /dev/null 2>&1
        elif command -v yum &> /dev/null; then
            sudo yum install -y -q bc > /dev/null 2>&1
        elif command -v dnf &> /dev/null; then
            sudo dnf install -y -q bc > /dev/null 2>&1
        else
            echo "Error: No compatible package manager found. Please install bc command manually and try again."
            exit 1
        fi

        # Check if installation was successful
        if ! command -v bc &> /dev/null; then
            echo "Error: bc command installation failed. Please install bc manually and try again."
            exit 1
        fi
    fi
}



# Function to write notification to log file if it's different from the last message content
write_notification() {
  local title="$1"
  local message="$2"
  local current_message="$(date '+%Y-%m-%d %H:%M:%S') UNREAD $title MESSAGE: $message"
  local last_message_content=$(get_last_message_content)

  # Check if the current message content is the same as the last one and has "UNREAD" status
  if [ "$message" != "$last_message_content" ] && ! is_unread_message_present "$title"; then
    echo "$current_message" >> "$LOG_FILE"
    if [ "$EMAIL_ALERT" == "no" ]; then
      echo "Email alerts are disabled."
    else
      email_notification "$title" "$message"
    fi

  fi
}




# Function to perform startup action (system reboot notification)
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





check_ssh_logins() {
    local title="Suspicious SSH login detected"
    local message_to_check_in_file="Suspicious SSH login detected"

    # Check if there is an unread notification
    if is_unread_message_present "$message_to_check_in_file"; then
      ((WARN++))
      echo -e "\e[38;5;214m[!]\e[0m Unread SSH login notification already exists. Skipping."
      return
    fi

    # Get IP addresses from the 'who' command
    ssh_ips=$(who | grep 'pts' | awk '{print $5}' | sed -E 's/[():]//g' | cut -d':' -f1)

    # Check if there are any IPs currently logged to SSH
    if [ -z "$ssh_ips" ]; then
        #echo "No active SSH sessions found."
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m No currently logged in SSH users detected."
        return
    fi

  # Get IP addresses from the login log file
  login_ips=$(awk '{print $NF}' /var/log/openpanel/admin/login.log)

    #avoid alerts when user installed openpanel, but not yet logged to admin!
    if [ -z "$login_ips" ]; then
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Detected logged-in SSH user, but login checks will be postponed until OpenAdmin interface is ready."
        return
    fi



  # Initialize counter and array to store unmatched IPs
  counter=0
  safe_counter=0
  suspecious_ips=()
  safe_ips=()

  # Loop through each IP from the 'who' command
  for ip in $ssh_ips; do
      # Check if the IP is not in the login log
      if ! echo "$login_ips" | grep -q "$ip"; then
          # Increment counter
          ((counter++))
          # Add IP to the array of unmatched IPs
          suspecious_ips+=("$ip")
      else
          # Increment counter
          ((safe_counter++))
          # Add IP to the array of matched IPs
          safe_ips+=("$ip")
      fi
  done

  # Output the unmatched IPs
  if [ $counter -gt 0 ]; then

      #echo "Number of IPs in the SSH sessions but not in the login log file: $counter"

      #message="$counter ${result[1]} "
      for unmatched_ip in "${suspecious_ips[@]}"; do
          message+="  $unmatched_ip"
      done

    ((FAIL++))
    STATUS=2

    echo -e "\e[31m[✘]\e[0m $message | Writing notification."
    write_notification "$title" "$message"

  else
      #echo "No unmatched IPs found."
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m Detected $safe_counter currently active SSH user(s), but marked as safe since the IP address has previously logged into the OpenAdmin interface."

      # Prepare the message
      message=""
      for ip in "${safe_ips[@]}"; do
          message+="${ip}    "
      done
      echo -e "    Currently active IP addresses: $message"

  fi


}







# Notify when admin account is accessed from new IP address
check_new_logins() {
  if [ "$LOGIN" != "no" ]; then
    touch /var/log/openpanel/admin/login.log

    # Check if the file is empty
    if [ ! -s "$LOG_FILE" ]; then
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin detected."
    else



    # Extract the last line from the login log file
    last_login=$(tail -n 1 /var/log/openpanel/admin/login.log)
        
    # Skip empty lines
    if [ -z "$last_login" ]; then
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m No new logins to OpenAdmin detected."
      return 1
    fi

    # Parse username and IP address from the last login entry
    username=$(echo "$last_login" | awk '{print $3}')
    ip_address=$(echo "$last_login" | awk '{print $4}')

    # Validate IP address format
    if [[ ! $ip_address =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo "Invalid IP address format: $ip_address"
      return 1
    fi

    # Exclude 127.0.0.1
    if [[ $ip_address == "127.0.0.1" ]]; then
      echo "IP address 127.0.0.1 is not allowed."
      return 1
    fi

    # Check if the username appears more than once in the log file
    if [ $(grep -c $username /var/log/openpanel/admin/login.log) -eq 1 ]; then
      ((PASS++))
      echo -e "\e[32m[✔]\e[0m First time login detected for user: $username from IP address: $ip_address."
    else
      # Check if the combination of username and IP address has appeared before
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




mysql_docker_containers_status() {

      # Check if the MySQL Docker container is running
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

                    # Log the error and write notification
                    error_log=$(docker --context=default logs -f --tail 10 "$service_name" 2>/dev/null | sed ':a;N;$!ba;s/\n/\\n/g')
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
                check_status_after_restart
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




# Function to check service status and write notification if not active
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
    echo -e "\e[31m[✘]\e[0m $service_name is not active." #Writing notification to log file."
    local error_log=""

      error_log=$(journalctl -n 5 -u "$service_name" 2>/dev/null | sed ':a;N;$!ba;s/\n/\\n/g')

      # Check if there's an error log and include it in the message
      if [ -n "$error_log" ]; then
        write_notification "$title" "$error_log"
      else
        echo "no logs."
      fi

      # Restart the service if it's not running
      echo -e "\e[33m[⚠️]\e[0m Restarting $service_name..."
      systemctl restart "$service_name"

      # Verify if the restart was successful
      if systemctl is-active --quiet "$service_name"; then
        echo -e "\e[32m[✔]\e[0m $service_name has been restarted and is now active."
      else
        echo -e "\e[31m[✘]\e[0m Failed to restart $service_name."
      fi
      
    fi
    

  fi
}



# Generate system crash report (optimized/refactored)
generate_crashlog_report() {
  # Prepare variables
  local crashlog_dir="/var/log/openpanel/admin/crashlog"
  local filename=$(date +%s)
  generated_report="${crashlog_dir}/${filename}.txt"
  local formatted_date=$(date '+%d %m %Y %H:%M')
  local hostname=$(hostname)
  local break_line="+---------------------------------------------------------------------------------------------------------------+"

  mkdir -p "$crashlog_dir" || return 1

  # Capture all command outputs up front
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

  # Write report to file
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




# Function to check system load and write notification if it exceeds the threshold
check_system_load() {
  local title="High System Load!"

  current_load=$(uptime | awk -F'average:' '{print $2}' | awk -F', ' '{print $1}')

  ensure_bc_installed

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




check_swap_usage() {
    local title="High SWAP usage!"

    # Run the 'free' command and capture the output
    free_output=$(free -t)

    # Extract the used and total swap values
    swap_used=$(echo "$free_output" | awk 'FNR == 3 {print $3}')
    swap_total=$(echo "$free_output" | awk 'FNR == 3 {print $2}')

    # Check if swap_total is greater than 0 to avoid division by zero
    if [ "$swap_total" -gt 0 ]; then
        # Calculate swap usage percentage
        SWAP_USAGE=$(awk "BEGIN {print ($swap_used / $swap_total) * 100}")
    else
        # If swap_total is 0, set SWAP_USAGE to 0 or handle it as appropriate
        SWAP_USAGE=0
        ((PASS++))
        echo -e "\e[32m[✔]\e[0m Total SWAP is $SWAP_USAGE (skipping swap check for ${SWAP_THRESHOLD}% treshold)"
        #echo "SWAP check was performed at: $TIME "
        return
    fi




    SWAP_USAGE_NO_DECIMALS=$(printf %.0f $SWAP_USAGE)
    
    #Execute check
    if [ "$SWAP_USAGE_NO_DECIMALS" -gt "$SWAP_THRESHOLD" ]; then

      # Check if the lock file exists and is older than 6 hours, then delete it
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
        # create when we start
        touch "$LOCK_FILE"
        
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
            # delete after success
            rm -f "$LOCK_FILE"
        else
                    ((FAIL++))
            STATUS=2

             echo -e "\e[31m[✘]\e[0m URGENT! SWAP could not be cleared on $HOSTNAME"
            write_notification "$title_not_ok" "The Sentinel service detected abnormal SWAP usage at $TIME and tried clearing the space but SWAP usage is still above the $swap_usage_no_decimals treshold."
        fi
    else
    ((PASS++))
        echo -e "\e[32m[✔]\e[0m Current SWAP usage is $SWAP_USAGE_NO_DECIMALS (bellow the ${SWAP_THRESHOLD}% treshold) - SWAP check was performed at: $TIME "
        # delete if failed but on next run it is ok
        rm -f "$LOCK_FILE"
    fi
}


# Function to check RAM usage and write notification if it exceeds the threshold
check_ram_usage() {
  local title="High Memory Usage!"
  local total_ram=$(free -m | awk '/^Mem:/{print $2}')
  local used_ram=$(free -m | awk '/^Mem:/{print $3}')
  local ram_percentage=$((used_ram * 100 / total_ram))
  
  local message="Used RAM: $used_ram MB, Total RAM: $total_ram MB, Usage: $ram_percentage%"
  local message_to_check_in_file="Used RAM"

  # Check if there is an unread RAM notification
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

function check_disk_usage() {
  local title="Running out of Disk Space!"

  local disk_percentage=$(df -h --output=pcent / | tail -n 1 | tr -d '%')

  if [ "$disk_percentage" -gt "$DISK_THRESHOLD" ]; then

  # Check if there is an unread DU notification
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







check_if_panel_domain_and_ns_resolve_to_server (){

# Extract force domain address from the configuration file
RESULT=$(opencli domain)

if [[ "$RESULT" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    # If it's a valid domain, use it
    FORCED_DOMAIN=$RESULT
    CHECK_DOMAIN_ALSO="yes"
else
    # If it's not a valid domain (could be an IP or invalid), do not proceed
    CHECK_DOMAIN_ALSO="no"
fi

# Extract NS1 from the configuration file
NS1_SET=$(awk -F'=' '/^ns1/ {print $2}' "$CONF_FILE")

# Check if NS1_SET is not empty and is a valid domain
if [ -n "$NS1_SET" ]; then
    if [[ "$NS1_SET" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        #echo "NS1_SET is a valid domain: $NS1_SET"
        NS1=$NS1_SET
        CHECK_NS_ALSO="yes"

        #check ns2 only if ns1 is set!
        NS2_SET=$(awk -F'=' '/^ns2/ {print $2}' "$CONF_FILE")
        NS2=$NS2_SET

    else
        echo "NS1 '$NS1_SET' is not a valid domain."
        CHECK_NS_ALSO="no"
    fi
else
    #echo "NS1_SET is empty or not found in $CONF_FILE."
    CHECK_NS_ALSO="no"
fi


if [ "$CHECK_DOMAIN_ALSO" == "no" ] && [ "$CHECK_NS_ALSO" == "no" ]; then
    ((WARN++))
    echo -e "\e[38;5;214m[!]\e[0m Missing or invalid custom domain and nameservers, skipping DNS checks.."
else

  # Google's DNS servers
  GOOGLE_DNS_SERVER="8.8.8.8"

  # Get server ipv4

  # IP SERVERS
  SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
  if [ -f "$SCRIPT_PATH" ]; then
      source "$SCRIPT_PATH"
  else
      IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
  fi


  SERVER_IP=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)

  # If server IP is not available from external service, use local IP
  if [ -z "$SERVER_IP" ]; then
      SERVER_IP=$(ip addr | grep 'inet ' | grep global | head -n1 | awk '{print $2}' | cut -f1 -d/)
  fi

  # Run checks only if either domain or nameserver is set
  if [ "$CHECK_DOMAIN_ALSO" == "yes" ] || [ "$CHECK_NS_ALSO" == "yes" ]; then
      # Check if FORCE_DOMAIN resolves to this server IP
      if [ "$CHECK_DOMAIN_ALSO" == "no" ]; then
          ((WARN++))
          echo -e "\e[38;5;214m[!]\e[0m Domain is not set for accessing OpenPanel, skip checking DNS."
      else
          if [ -n "$FORCED_DOMAIN" ]; then
              #echo "Checking if $FORCED_DOMAIN resolves to this server IP ($SERVER_IP).."

              ensure_dig_installed
              
              domain_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$FORCED_DOMAIN")
              

              if [ "$domain_ip" == "$SERVER_IP" ]; then
              ((PASS++))
                  echo -e "\e[32m[✔]\e[0m $FORCED_DOMAIN resolves to $SERVER_IP"
              else
                  # Check if Cloudflare proxy
                  ns_records=$(dig +short @"$GOOGLE_DNS_SERVER" NS "$FORCED_DOMAIN")

                  if echo "$ns_records" | grep -q 'cloudflare'; then
                     ((WARN++))
                     echo -e "\e[38;5;214m[!]\e[0m $FORCED_DOMAIN does not resolve to $SERVER_IP, but it is using Cloudflare DNS, so it is possible that proxy option is enabled. Skipping."
                  else
                      #echo "$FORCED_DOMAIN is not using Cloudflare DNS."
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

      # Check if NS1 and NS2 resolve to this server IP
      if [ "$CHECK_NS_ALSO" == "no" ]; then
      ((PASS++))
          echo -e "\e[32m[✔]\e[0m Skip checking nameservers as they are not set."
      else
          if [ -n "$NS1" ] && [ -n "$NS2" ]; then
              ns1_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS1")
              ns2_ip=$(dig +short @"$GOOGLE_DNS_SERVER" "$NS2")
              all_server_ips=$(hostname -I)

              if echo "$all_server_ips" | grep -qw "$ns1_ip" && echo "$all_server_ips" | grep -qw "$ns2_ip"; then
                  ((PASS++))
                  echo -e "\e[32m[✔]\e[0m $NS1 and $NS2 both resolve to local IPs ($(hostname -I))"
              else
                  ((FAIL++))
                  STATUS=2
                  echo -e "\e[31m[✘]\e[0m Nameservers do not resolve correctly:"
                  echo "    $NS1 resolves to $ns1_ip (expected one of: $all_server_ips)"
                  echo "    $NS2 resolves to $ns2_ip (expected one of: $all_server_ips)"
                  local title="Configured nameservers do not resolve to local IPs"
                  local message="$NS1 resolves to $ns1_ip (expected one of: $all_server_ips) | $NS2 resolves to $ns2_ip (expected one of: $all_server_ips)"
                  write_notification "$title" "$message"
              fi
          else 
            ((WARN++))
            echo -e "\e[38;5;214m[!]\e[0m Only one nameserver is currently set. Please add at least two nameservers to ensure DNS redundancy."
          fi
      fi
  fi
fi

}







# logo
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






# Check for flags
if [ "$1" == "--startup" ]; then
  perform_startup_action
elif [ "$1" == "--report" ]; then
  email_daily_report
else
  print_header
  check_for_debug_and_print_info

  # Check service statuses and write notifications if needed

  echo "Checking health for monitored services:"
  echo ""



###### def funkcije


######








check_services() {
  if echo "$SERVICES" | grep -q "caddy"; then
    docker_containers_status "caddy" "Caddy container is not active. Users' websites are not working!"
  fi

  if echo "$SERVICES" | grep -q "csf"; then
    check_service_status "csf" "ConfigService Firewall (CSF) is not active. Server and websites are not protected!"
  fi
  
  if echo "$SERVICES" | grep -q "admin"; then
      check_service_status "admin" "Admin service is not active. OpenAdmin service is not accessible!"
  fi

  if echo "$SERVICES" | grep -q "docker"; then
    check_service_status "docker" "Docker service is not active. User websites are down!"
  fi

  if echo "$SERVICES" | grep -q "panel"; then
    docker_containers_status "openpanel" "OpenPanel docker container is not running. Users are unable to access the OpenPanel interface!"
  fi

  if echo "$SERVICES" | grep -q "mysql"; then
    mysql_docker_containers_status
  fi

  if echo "$SERVICES" | grep -q "named"; then
    docker_containers_status "openpanel_dns" "Named (BIND9) service is not active. DNS resolving of domains is not working!"
  fi
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
  # Summary
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

wget "$PROGRESS_BAR_URL" -O "$PROGRESS_BAR_FILE" > /dev/null 2>&1

if [ ! -f "$PROGRESS_BAR_FILE" ]; then
    echo "Failed to download progress_bar.sh"
    exit 1
fi

# Source the progress bar script
source "$PROGRESS_BAR_FILE"

# Dsiplay progress bar
FUNCTIONS=(
  # SERVICES
  check_services

  #LOGINS
  start_login_section
  check_new_logins
  check_ssh_logins

  #check_user_panel_logins

  #USAGE
  check_resources_section
  check_disk_usage
  check_system_load
  check_ram_usage
  check_cpu_usage
  check_swap_usage

  #DNS
  check_dns_section
  check_if_panel_domain_and_ns_resolve_to_server

  # kraj
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
    # Make sure that the progress bar is cleaned up when user presses ctrl+c
    enable_trapping
    
    # Create progress bar
    setup_scroll_area
    for func in "${FUNCTIONS[@]}"
    do
        # Execute each function
        $func
        update_progress
    done
    destroy_scroll_area
}



  main
  # TODO: add no of services, alerts notified and emailed.
fi

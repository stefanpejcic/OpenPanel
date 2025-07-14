#!/bin/bash
################################################################################
# Script Name: user/add.sh
# Description: Create a new user with the provided plan_name.
# Usage: opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug]  [--webserver="<nginx|apache|openresty|varnish+nginx|varnish+apache|varnish+openresty>"] [--sql=<mysql|mariadb>] [--reseller=<RESELLER_USERNAME>][--server=<IP_ADDRESS>]  [--key=<SSH_KEY_PATH>]
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 01.10.2023
# Last Modified: 13.07.2025
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

# Constants
FORBIDDEN_USERNAMES_FILE="/etc/openpanel/openadmin/config/forbidden_usernames.txt"
DB_CONFIG_FILE="/usr/local/opencli/db.sh"
SEND_EMAIL_FILE="/usr/local/opencli/send_mail"
PANEL_CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"

username="${1,,}"
password="$2"
email="$3"
plan_name="$4"
DEBUG=false             # Default value for DEBUG
SEND_EMAIL=false        # Don't send email by default
server=""               # Default value for context
key_flag=""

usage() {
    echo "Usage: opencli user-add <username> <password|generate> <email> '<plan_name>' [--send-email] [--debug] [--reseller=<RESELLER_USER>] [--server=<IP_ADDRESS>] [--key=<KEY_PATH>]"
    echo
    echo "Required arguments:"
    echo "  <username>                 The username of the new user."
    echo "  <password|generate>        The password for the new user, or 'generate' to auto-generate a password."
    echo "  <email>                    The email address associated with the new user."
    echo "  <plan_name>                The plan to assign to the new user."
    echo
    echo "Optional flags:"

    if [ -n "$key_value" ]; then
	    # Reseller section
	    printf "%-25s %-45s\n" "  --reseller=<RESELLER_USER>" "Set the reseller for the user."
	    echo
	
	    # Clustering section
	    printf "%-25s %-45s\n" "  --server=<IP_ADDRESS>" "Specify the IPv4 of slave server to use for user."
	    printf "%-25s %-45s\n" "  --key=<KEY_PATH>" "Specify the path to a key file for authentication to slave server."
	    echo
    fi

    # Other
    printf "%-25s %-45s\n" "  --send-email" "Send a welcome email to the user."
    printf "%-25s %-45s\n" "  --debug" "Enable debug mode for additional output."
    echo
    exit 1
}



check_enterprise() {
    key_value=$(grep "^key=" $PANEL_CONFIG_FILE | cut -d'=' -f2-)
}





#1. check if enterprise license, so we can display more info in usage()
check_enterprise

#2. Check the number of arguments, we need at least 4
if [ "$#" -lt 4 ] || [ "$#" -gt 11 ]; then
    usage
fi


#4. remove loc file on exit
cleanup() {
  #echo "[✘] Script failed. Cleaning up..."
  rm /var/lock/openpanel_user_add.lock > /dev/null 2>&1
  exit 1
}
trap cleanup EXIT


hard_cleanup() {
  # todo: remove user, files, container..
  killall -u $username -9  > /dev/null 2>&1
  deluser --remove-home $username  > /dev/null 2>&1   # command missing on alma!
  rm -rf /etc/openpanel/openpanel/core/stats/$username
  rm -rf /etc/openpanel/openpanel/core/users/$username
  docker context rm $username  > /dev/null 2>&1
  quotacheck -avm >/dev/null 2>&1
  repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
  exit 1
}










check_if_default_slave_server_is_set() {

	: '
	[CLUSTERING]
	default_node="11.22.33.44"
	default_ssh_key_path="/root/some-key.rsa"
	'

	get_config_value() {
	    local file="/etc/openpanel/openadmin/config/admin.ini"
	    local key=$1
	    local value=$(grep -E "^$key=" "$file" | awk -F= '{sub(/^ /, "", $2); print $2}')
	
	    if [[ -z "$value" ]]; then
	        echo ""
	    else
	        echo "$value"
	    fi
	}
	
	
	# Check and get values
	default_node=$(get_config_value "default_node")
	
	if [[ -n "$default_node" ]]; then
		default_ssh_key_path=$(get_config_value "default_ssh_key_path")
		 if [[ -n "$default_ssh_key_path" ]]; then
	              server=$default_node
		      key=$default_ssh_key_path
	              echo 'Using default node "$server" and ssh key path'
		 fi
	fi

}





check_if_default_slave_server_is_set         # we run it before parse_flags so it can be overwritten!


	for arg in "$@"; do
	    case $arg in
	        --debug)
	            DEBUG=true
	            ;;
	        --send-email)
	            SEND_EMAIL=true
	            ;;
	        --reseller=*)
	            reseller="${arg#*=}"
		    ;;
	        --server=*)
	            server="${arg#*=}"
		    ;;
		--key=*)
	             key="${arg#*=}"
		     key_flag="-i $key"
	            ;;
	        --sql=*)
	            sql_type="${arg#*=}"
		    ;;	    
	        --webserver=*)
	            webserver="${arg#*=}"
		    ;;
	    esac
	done


log() {
    if $DEBUG; then
        echo "$1"
    fi
}


get_hostname_of_master() {
	hostname=$(hostname)
}






check_if_reseller() {

	local db_file_path="/etc/openpanel/openadmin/users.db"
	if [ -n "$reseller" ]; then
	    log "Checking if reseller user exists and can create new users.."
	
    	    local user_exists=$(sqlite3 "$db_file_path" "SELECT COUNT(*) FROM user WHERE username='$reseller' AND role='reseller';")
		
	    if [ "$user_exists" -lt 1 ]; then
	        echo -e "ERROR: User '$reseller' is not a reseller or not allowed to create new users. Contact support."
	    fi

	    local reseller_limits_file="/etc/openpanel/openadmin/resellers/$reseller.json"
     
	    if [ -f "$reseller_limits_file" ]; then
	  	log "Checking reseller limits.."
    
		local query_for_owner="SELECT COUNT(*) FROM users WHERE owner='$reseller';"
		current_accounts=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -se "$query_for_owner")
		
		#echo "DEBUG: current_accounts='$current_accounts'"

		if [ $? -eq 0 ]; then
		    if [[ -z "$current_accounts" ]]; then
		        current_accounts=0
		    fi
		
		    jq --argjson current_accounts "$current_accounts" '.current_accounts = $current_accounts' "$reseller_limits_file" > "/tmp/${reseller}_config.json" \
		        && mv "/tmp/${reseller}_config.json" "$reseller_limits_file"
		else
		    log "Error fetching current account count for reseller $reseller from MySQL."
		    echo "ERROR: Unable to retrieve the number of users from the database. Is MySQL running?"
		    exit 1
		fi

	  
	        max_accounts=$(jq -r '.max_accounts // "unlimited"' $reseller_limits_file)
                local allowed_plans=$(jq -r '.allowed_plans | join(",")' $reseller_limits_file)

	           # Compare current account count with the max limit
	        if [ "$max_accounts" != "unlimited" ] && [ "$current_accounts" -ge "$max_accounts" ]; then
	            echo "ERROR: Reseller has reached the maximum account limit. Cannot create more users."
	            exit 1
	        fi
	 
	        # Check if the current plan ID is in the allowed plans
	        if echo "$allowed_plans" | grep -wq "$plan_id"; then
	            :
	            #log "Current plan ID '$plan_id' is allowed for this reseller."
	        else
	            echo "ERROR: Current plan ID '$plan_id' is not assigned for this reseller. Please select another plan."
	            exit 1
	        fi
	   
	    else
	 	  log "WARNING: Reseller $reseller has no limits configured and can create unlimited number of users."
	    		# TODO: notify admin - log in adminn log!
	    fi
    fi
}



get_slave_if_set() {
     
     
	if [ -n "$server" ]; then
	    # Check if the format of the server is a valid IPv4 address
	    if [[ "$server" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
	        # Check if each octet is in the range 0-255
	        IFS='.' read -r -a octets <<< "$server"
	        if [[ ${octets[0]} -ge 0 && ${octets[0]} -le 255 &&
	              ${octets[1]} -ge 0 && ${octets[1]} -le 255 &&
	              ${octets[2]} -ge 0 && ${octets[2]} -le 255 &&
	              ${octets[3]} -ge 0 && ${octets[3]} -le 255 ]]; then
	           	
			context_flag="--context $server"     
			hostname=$(ssh $key_flag "root@$server" "hostname")
			if [ -z "$hostname" ]; then
			  echo "ERROR: Unable to reach the node $server - Exiting."
     			  echo "       Make sure you can connect to the node from terminal with: 'ssh $key_flag root@$server -vvv'"
			  exit 1
			fi
   
   			node_ip_address=$server
      			context=$username # so we show it on debug!
  
	     		log "Container will be created on node: $node_ip_address ($hostname)"
	        else
	            echo "ERROR: $server is not a valid IPv4 address (octets out of range)."
	        fi
	    else
	        echo "ERROR: $server is not a valid IPv4 address (invalid format)."
	 	exit 1
	    fi
     	else
      		# local values
                context_flag="" 
		hostname=$(hostname)
	fi
    
}





validate_password_in_lists() {
    local password_to_check="$1"
    weakpass=$(grep -E "^weakpass=" "$PANEL_CONFIG_FILE" | awk -F= '{print $2}')

    if [ -z "$weakpass" ]; then
      if [ "$DEBUG" = true ]; then
        echo "weakpass value not found in openpanel.config. Defaulting to 'yes'."
      fi
      weakpass="yes"
    fi
    
    # https://weakpass.com/wordlist
    # https://github.com/steveklabnik/password-cracker/blob/master/dictionary.txt
    
    if [ "$weakpass" = "no" ]; then
      if [ "$DEBUG" = true ]; then
        echo "Checking the password against weakpass dictionaries"
      fi

       wget -O /tmp/weakpass.txt https://github.com/steveklabnik/password-cracker/blob/master/dictionary.txt > /dev/null 2>&1
       
       if [ -f "/tmp/weakpass.txt" ]; then
            DICTIONARY="dictionary.txt"
            local input_lower=$(echo "$password_to_check" | tr '[:upper:]' '[:lower:]')
        
            # Check if input contains any common dictionary word
            if grep -qi "^$input_lower$" "$DICTIONARY"; then
                echo "[✘] ERROR: password contains a common dictionary word from https://weakpass.com/wordlist"
                echo "       Please use stronger password or disable weakpass check with: 'opencli config update weakpass no'."
                rm dictionary.txt
                exit 1
            fi
            rm dictionary.txt
       else
	       echo "[!] WARNING: Error downloading dictionary from https://weakpass.com/wordlist"
       fi
    elif [ "$weakpass" = "yes" ]; then
      :
    else
      if [ "$DEBUG" = true ]; then
        echo "Invalid weakpass value '$weakpass'. Defaulting to 'yes'."
      fi
      weakpass="yes"
      :
    fi
}


check_username_is_valid() {


    is_username_forbidden() {
        local check_username="$1"
        log "Checking if username $username is in the forbidden usernames list"
        readarray -t forbidden_usernames < "$FORBIDDEN_USERNAMES_FILE"
    
        # Check against forbidden usernames
        for forbidden_username in "${forbidden_usernames[@]}"; do
            if [[ "${check_username,,}" == "${forbidden_username,,}" ]]; then
                return 0
            fi
        done
    
        return 1
    }



	is_username_valid() {
	    local check_username="$1"
	    log "Checking if username $check_username is valid"
	    
	    if [[ "$check_username" =~ [[:space:]] ]]; then
	        echo "[✘] Error: The username cannot contain spaces."
	        return 0
	    fi
	    
	    if [[ "$check_username" =~ [-_] ]]; then
	        echo "[✘] Error: The username cannot contain hyphens or underscores."
	        return 0
	    fi
	    
	    if [[ ! "$check_username" =~ ^[a-zA-Z0-9]+$ ]]; then
	        echo "[✘] Error: The username can only contain letters and numbers."
	        return 0
	    fi
	    
	    if [[ "$check_username" =~ ^[0-9]+$ ]]; then
	        echo "[✘] Error: The username cannot consist entirely of numbers."
	        return 0
	    fi
	    
	    if (( ${#check_username} < 3 )); then
	        echo "[✘] Error: The username must be at least 3 characters long."
	        return 0
	    fi
	    
	    if (( ${#check_username} > 20 )); then
	        echo "[✘] Error: The username cannot be longer than 20 characters."
	        return 0
	    fi
	    
	    return 1
	}

    
    # Validate username
    if is_username_valid "$username"; then
    	echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#openpanel"
        exit 1
    elif is_username_forbidden "$username"; then
        echo "[✘] Error: The username '$username' is not allowed."
        echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#reserved-usernames"
        exit 1
    fi
}


# Source the database config file
. "$DB_CONFIG_FILE"


get_existing_users_count() {
    
    # Check if 'enterprise edition'
    if [ -n "$key_value" ]; then
        if [ -n "$reseller" ]; then
		:
        else
        	log "Enterprise edition detected: unlimited number of users can be created"
	 fi
    else
    	if [ -n "$reseller" ]; then
            echo "[✘] ERROR: Resellers feature requires the Enterprise edition."
	    echo "If you require reseller accounts, please consider purchasing the Enterprise version that allows unlimited number of users and resellers."
            exit 1
     	fi
       log "Checking if the limit of 3 users on Community edition is reached"
        # Check the number of users from the database
        user_count_query="SELECT COUNT(*) FROM users"
        user_count=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$user_count_query" -sN)
    
        # Check if successful
        if [ $? -ne 0 ]; then
            echo "[✘] ERROR: Unable to get total user count from the database. Is mysql running?"
            exit 1
        fi
    
        # Check if the number of users is >= 3
        if [ "$user_count" -gt 2 ]; then
            echo "[✘] ERROR: OpenPanel Community edition has a limit of 3 user accounts - which should be enough for private use."
	    echo "If you require more than 3 accounts, please consider purchasing the Enterprise version that allows unlimited number of users and domains/websites."
            exit 1
        fi
    fi

}

# Function to check if username already exists in the database
check_username_exists() {
    local username_exists_query="SELECT COUNT(*) FROM users WHERE username = '$username'"
    local username_exists_count=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$username_exists_query" -sN)

    # Check if successful
    if [ $? -ne 0 ]; then
        echo "[✘] Error: Unable to check username existence in the database. Is mysql running?"
        exit 1
    fi

    # Return the count of usernames found
    echo "$username_exists_count"
}


# Check if the username exists in the database
username_exists_count=$(check_username_exists)

# Check if the username exists
if [ "$username_exists_count" -gt 0 ]; then
    echo "[✘] Error: Username '$username' is already taken."
    exit 1
fi


#########################################
# TODO
# USE REMOTE CONTEXT! context_flag
#
#########################################
#
#


sshfs_mounts() {
    if [ -n "$node_ip_address" ]; then


ssh -q $key_flag root@$node_ip_address << EOF
  if [ ! -d "/etc/openpanel/openpanel" ]; then
	echo "PubkeyAuthentication yes" >> /etc/ssh/sshd_config
	echo "AuthorizedKeysFile     .ssh/authorized_keys" >> /etc/ssh/sshd_config
	service ssh restart
fi
EOF

	sleep 5

# mount openpanel dir on slave

# SSH into the slave server and check if /etc/openpanel exists
ssh -q  $key_flag root@$node_ip_address << EOF
  if [ ! -d "/etc/openpanel/openpanel" ]; then
    echo "Node is not yet configured to be used as an OpenPanel slave server. Configuring.."

    # Check for the package manager and install sshfs accordingly
    if command -v apt-get &> /dev/null; then
      export DEBIAN_FRONTEND=noninteractive
      apt-get update > /dev/null 2>&1 && apt-get -yq install systemd-container uidmap
    elif command -v dnf &> /dev/null; then
      dnf install -y systemd-container uidmap
    elif command -v yum &> /dev/null; then
      yum install -y systemd-container uidmap
    else
      echo "[✘] ERROR: Unable to setup the slave server. Contact support."
      exit 1
    fi
    mkdir -p /etc/openpanel
    git clone https://github.com/stefanpejcic/OpenPanel-configuration /etc/openpanel
EOF


# https://docs.docker.com/engine/security/rootless/#limiting-resources
ssh -q  $key_flag root@$node_ip_address << 'EOF'
 
  if [ ! -d "/etc/openpanel/openpanel" ]; then
    echo "Adding permissions for users to limit CPU% - more info: https://docs.docker.com/engine/security/rootless/#limiting-resources"
  
    mkdir -p /etc/systemd/system/user@.service.d
  
    cat > /etc/systemd/system/user@.service.d/delegate.conf << 'INNER_EOF'
[Service]
Delegate=cpu cpuset io memory pids
INNER_EOF

    systemctl daemon-reload
  fi

EOF



# TODO:
#  scp -r /etc/openpanel root@$node_ip_address:/etc/openpanel
# sync conf from master to slave!

    # mount home dir on master
    if command -v sshfs &> /dev/null; then
    	:
    else
	    # Check for the package manager and install sshfs accordingly
	    if command -v apt-get &> /dev/null; then
	      apt-get install -y sshfs
	    elif command -v dnf &> /dev/null; then
	      dnf install -y sshfs
	    elif command -v yum &> /dev/null; then
	      yum install -y sshfs
	    else
	      echo "[✘] ERROR: Unable to setup the slave server. Contact support."
	      exit 1
	    fi
    fi
	sshfs -o IdentityFile=~/.ssh/$node_ip_address root@$node_ip_address:/home/$username /home/$username

fi
 
}

#Get CPU, DISK, INODES and RAM limits for the plan
get_plan_info_and_check_requirements() {
    log "Getting information from the database for plan $plan_name"
    # Fetch DISK, CPU, RAM, INODES, BANDWIDTH and NAME for the given plan_name from the MySQL table
    query="SELECT cpu, ram, disk_limit, inodes_limit, bandwidth, id FROM plans WHERE name = '$plan_name'"
    
    # Execute the MySQL query and store the results in variables
    cpu_ram_info=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$query" -sN)
    
    # Check if the query was successful
    if [ $? -ne 0 ]; then
        echo "[✘] ERROR: Unable to fetch plan information from the database."
        exit 1
    fi
    
    # Check if any results were returned
    if [ -z "$cpu_ram_info" ]; then
        echo "[✘] ERROR: Plan with name $plan_name not found. Unable to fetch CPU/RAM limits information from the database."
        exit 1
    fi
    
    # Extract DISK, CPU, RAM, INODES, BANDWIDTH and NAME,values from the query result
    cpu=$(echo "$cpu_ram_info" | awk '{print $1}')
    ram=$(echo "$cpu_ram_info" | awk '{print $2}')
    disk_limit=$(echo "$cpu_ram_info" | awk '{print $3}' | sed 's/ //;s/B//')
    # 4. is GB in disk_limit
    inodes=$(echo "$cpu_ram_info" | awk '{print $5}')
    bandwidth=$(echo "$cpu_ram_info" | awk '{print $6}')
    plan_id=$(echo "$cpu_ram_info" | awk '{print $7}')
        
    # Get the maximum available CPU cores on the server
    if [ -n "$node_ip_address" ]; then
        # TODO: Use a custom user or configure SSH instead of using root
        max_available_cores=$(ssh $key_flag "root@$node_ip_address" "nproc")
    else
        max_available_cores=$(nproc)
    fi
    
    # Compare the specified CPU cores with the maximum available cores
    if [ "$cpu" -gt "$max_available_cores" ]; then
        echo "[✘] ERROR: CPU cores ($cpu) limit on the plan exceed the maximum available cores on the server ($max_available_cores). Cannot create user."
        exit 1
    fi
    
    

    # Get the maximum available RAM on the server in GB
    if [ -n "$node_ip_address" ]; then
	max_available_ram_gb=$(ssh $key_flag "root@$node_ip_address" "free -g | awk '/Mem:/{print \$2}'")
    else
        max_available_ram_gb=$(free -g | awk '/^Mem:/{print $2}')
    fi    
    numram="${ram%"g"}"

    
    # Compare the specified RAM with the maximum available RAM
    if [ "$numram" -gt "$max_available_ram_gb" ]; then
        echo "[✘] ERROR: RAM ($ram GB) limit on the plan exceeds the maximum available RAM on the server ($max_available_ram_gb GB). Cannot create user."
        exit 1
    fi
}




setup_ssh_key(){
     if [ -n "$node_ip_address" ]; then
	log "Setting ssh key.."
 
 	public_key=$(ssh-keygen -y -f "$key")
ssh $key_flag root@$node_ip_address << EOF
  mkdir -p /home/$username/.ssh/ > /dev/null 2>&1
  touch  /home/$username/.ssh/authorized_keys > /dev/null 2>&1
  chown $username -R /home/$username/.ssh > /dev/null 2>&1
  if ! grep -q "$public_key" /home/$username/.ssh/authorized_keys; then
    echo "$public_key" >> /home/$username/.ssh/authorized_keys
  fi
  
EOF

mkdir ~/.ssh  > /dev/null 2>&1
chmod 600 ~/.ssh/config  > /dev/null 2>&1
   
cp $key ~/.ssh/$node_ip_address && chmod 600 ~/.ssh/$node_ip_address

mkdir -p ~/.ssh/cm_socket  > /dev/null 2>&1
   
echo "Host $username
    HostName $node_ip_address
    User $username
    IdentityFile ~/.ssh/$node_ip_address
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    ControlPath ~/.ssh/cm_socket/%r@%h:%p
    ControlMaster auto
    ControlPersist 30s
" >> ~/.ssh/config


	                ssh "$username" "exit" &> /dev/null
	                if [ $? -eq 0 ]; then
	                    log "SSH connection successfully established"
		     	else
			    echo "ERROR: Failed to establish SSH connection to the newly created user."
	      		    exit 1
	     		fi
 
      fi
}



validate_ssh_login(){
   if [ -n "$node_ip_address" ]; then
   	log "Validating SSH connection to the server $node_ip_address"
    	if [ "$DEBUG" = true ]; then
	        if [ -f "$key" ]; then	
	            if [ "$(stat -c %a "$key")" -eq 600 ]; then	                
	                ssh -i "$key" "$node_ip_address" "exit" &> /dev/null
	                if [ $? -eq 0 ]; then
	                    log "SSH connection successfully established"
	                else
	                    echo 'ERROR: SSH connection failed to $node_ip_address'
		     	    exit 1
	                fi
	            else
	                log "SSH key permissions are incorrect. Correcting permissions to 600."
	                chmod 600 "$key"
	            fi
	        else
	            echo 'ERROR: Provided ssh key path: "$key" does not exist.'
	     	# TODO: GENERATE OTHERWISE!
		#ssh-keygen -t rsa -b 4096 -f ~/.ssh/${username}context
		# ssh-copy-id $username@$hostname+IPPPPP      
	     	    exit 1
	        fi
 

	fi
    fi
}








# DEBUG
print_debug_info_before_starting_creation() {
    if [ "$DEBUG" = true ]; then
   	 echo "--------------------------------------------------------------"
	if [ -n "$reseller" ]; then
	        echo "Reseller user information:"
	        echo "- Reseller:             $reseller" 
	        echo "- Existing accounts:    $current_accounts/$max_accounts" 	 
	        echo "--------------------------------------------------------------"

	fi
	
	if [ -n "$node_ip_address" ]; then
	        echo "Data for connecting to the Node server:"
	        echo "- IP address:           $node_ip_address" 
	        echo "- Hostname:             $hostname" 	 
	        echo "- SSH user:             root" 	
	        echo "- SSH key path:         $key" 		
	        echo "--------------------------------------------------------------"

	fi
        echo "Selected plan limits from database:"
        echo "- plan id:           $plan_id" 
        echo "- plan name:         $plan_name"
        echo "- cpu limit:         $cpu"
        echo "- memory limit:      $ram"
	
	if [ "$disk_limit" -eq 0 ]; then
        	echo "- storage:           unlimited"
	else
		echo "- storage:           $disk_limit GB"
	fi
 
 	if [ "$inodes" -eq 0 ]; then
        	echo "- inodes:            unlimited"
	else
		echo "- inodes:            $inodes"
	fi
 
        echo "- port speed:        $bandwidth"
	echo "--------------------------------------------------------------"
    fi
}


create_local_user() {
	log "Creating user $username"
	useradd -m -d /home/$username $username
 	user_id=$(id -u $username)	
	if [ $? -ne 0 ]; then
		echo "Error: Failed creating linux user $username on master server."
  		exit 1
	fi
 
 	# https://github.com/stefanpejcic/OpenPanel/issues/367
  	#echo "$username:$password" | chpasswd

}

create_remote_user() {
	local provided_id=$1
        if [ -n "$provided_id" ]; then
		id_flag="-u $provided_id"
 	else
		id_flag=""
 	fi
  	
   	if [ -n "$node_ip_address" ]; then
                    log "Creating user $username on server $node_ip_address"
                    ssh $key_flag "root@$node_ip_address" "useradd -m -s /bin/bash -d /home/$username $id_flag $username" #-s /bin/bash needed for sourcing 
		    user_id=$(ssh $key_flag "root@$node_ip_address" "id -u $username")
			if [ $? -ne 0 ]; then
			    echo "Error: Failed creating linux user $username on node: $node_ip_address"
			    exit 1
			fi
	fi
 

}

set_user_quota(){

    if [ "$disk_limit" -ne 0 ]; then
    	storage_in_blocks=$((disk_limit * 1024000))
        log "Setting storage size of ${disk_limit}GB and $inodes inodes for the user"
      	setquota -u $username $storage_in_blocks $storage_in_blocks $inodes $inodes /
    else
    	log "Setting unlimited storage and inodes for the user"
      	setquota -u $username 0 0 0 0 /
    fi

}

# CREATE THE USER
create_user_set_quota_and_password() {
 	create_local_user
       	create_remote_user $user_id
	set_user_quota
}


docker_compose() {
	local arm_link="https://github.com/docker/compose/releases/download/v2.36.0/docker-compose-linux-aarch64"
 	local x86_link="https://github.com/docker/compose/releases/download/v2.36.0/docker-compose-linux-x86_64"
   	if [ -n "$node_ip_address" ]; then
    		# TODO: CHECK ARMCPU!
	    	log "Configuring Docker Compose for user $username on node $node_ip_address"
		ssh $key_flag root@$node_ip_address "su - $username -c '
		DOCKER_CONFIG=\${DOCKER_CONFIG:-/home/$username/.docker}
		mkdir -p /home/$username/.docker/cli-plugins
		curl -sSL $x86_link -o /home/$username/.docker/cli-plugins/docker-compose
		chmod +x /home/$username/.docker/cli-plugins/docker-compose
		'"
	else	
 		architecture=$(lscpu | grep Architecture | awk '{print $2}')
	    	log "Configuring Docker Compose for user $username"
      		if [ "$architecture" == "aarch64" ]; then
			log "Setting compose for ARM CPU (/etc/openpanel/docker/docker-compose-linux-aarch64)"
			system_wide_compose_file="/etc/openpanel/docker/docker-compose-linux-aarch64"
	      		if [ ! -f "$system_wide_compose_file" ]; then
				curl -sSL $arm_link -o $system_wide_compose_file
			fi
  		else 
    			log "Setting compose for x86_64 CPU (/etc/openpanel/docker/docker-compose-linux-x86_64)"
			system_wide_compose_file="/etc/openpanel/docker/docker-compose-linux-x86_64"
	   
	      		if [ ! -f "$system_wide_compose_file" ]; then
				curl -sSL $x86_link -o $system_wide_compose_file
			fi
     		fi
     		chmod +x $system_wide_compose_file
      		mkdir -p /home/$username/.docker/cli-plugins
	        ln -sf $system_wide_compose_file /home/$username/.docker/cli-plugins/docker-compose
	fi
}



docker_rootless() {

log "Configuring Docker in Rootless mode"

mkdir -p /home/$username/docker-data /home/$username/.config/docker > /dev/null 2>&1
cp /etc/openpanel/docker/daemon/rootless.json /home/$username/.config/docker/daemon.json
sed -i "s/USERNAME/$username/g" /home/$username/.config/docker/daemon.json
		
mkdir -p /home/$username/bin > /dev/null 2>&1
chmod 755 -R /home/$username/ >/dev/null 2>&1
sed -i '1i export PATH=/home/'"$username"'/bin:$PATH' /home/"$username"/.bashrc


   	if [ -n "$node_ip_address" ]; then
log "Setting AppArmor profile.."
ssh $key_flag root@$node_ip_address <<'EOF1'

cat > "/etc/apparmor.d/home.$username.bin.rootlesskit" <<'EOT1'
abi <abi/4.0>,
include <tunables/global>

  /home/$username/bin/rootlesskit flags=(unconfined) {
    userns,
    include if exists <local/home.$username.bin.rootlesskit>
  }
EOT1

# Generate the filename for the profile
filename=$(echo "/home/$username/bin/rootlesskit" | sed -e 's@^/@@' -e 's@/@.@g')

# Create the rootlesskit profile for the user directly
cat > "/home/$username/${filename}" <<'EOT2'
abi <abi/4.0>,
include <tunables/global>

  "/home/$username/bin/rootlesskit" flags=(unconfined) {
    userns,
    include if exists <local/${filename}>
  }
EOT2

# Move the generated file to the AppArmor directory
mv "/home/$username/${filename}" "/etc/apparmor.d/${filename}"

EOF1




log "Restarting services.."


		ssh $key_flag root@$node_ip_address "
		# Restart apparmor service
		systemctl restart apparmor.service >/dev/null 2>&1
		
		# Enable lingering for the user to keep their session alive across reboots
		loginctl enable-linger $username >/dev/null 2>&1
		
		# Create necessary directories and set permissions
		mkdir -p /home/$username/.docker/run >/dev/null 2>&1
		chmod 700 /home/$username/.docker/run >/dev/null 2>&1
		
		# Set the appropriate permissions for the user home directory
		chmod 755 -R /home/$username/ >/dev/null 2>&1
		chown -R $username:$username /home/$username/ >/dev/null 2>&1
  		"



  log "Downloading https://get.docker.com/rootless"

ssh $key_flag root@$node_ip_address "
    su - $username -c 'bash -l -c \"
        cd /home/$username/bin
        wget -O /home/$username/bin/dockerd-rootless-setuptool.sh https://get.docker.com/rootless > /dev/null 2>&1
        chmod +x /home/$username/bin/dockerd-rootless-setuptool.sh
        /home/$username/bin/dockerd-rootless-setuptool.sh install > /dev/null 2>&1

	echo \\\"export XDG_RUNTIME_DIR=/home/$username/.docker/run\\\" >> ~/.bashrc
        echo \\\"export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/\$(id -u)/bus\\\" >> ~/.bashrc
        
    \"'
"

#         echo \\\"export DOCKER_HOST=unix:///home/$username/.docker/run/docker.sock\\\" >> ~/.bashrc
        
  log "Configuring Docker service.."

ssh $username "
    mkdir -p ~/.config/systemd/user/
    cat > ~/.config/systemd/user/docker.service <<EOF
[Unit]
Description=Docker Application Container Engine (Rootless)
After=network.target

[Service]
Environment=PATH=/home/$username/bin:$PATH
Environment=DOCKER_HOST=unix:///home/$username/.docker/run/docker.sock
ExecStart=/home/$username/bin/dockerd-rootless.sh -H unix:///home/$username/.docker/run/docker.sock

Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=default.target
EOF
"




  log "Starting user services.."

ssh $key_flag root@$node_ip_address "
    machinectl shell $username@ /bin/bash -c '

        echo \"XDG_RUNTIME_DIR=\$XDG_RUNTIME_DIR\"
        echo \"DBUS_SESSION_BUS_ADDRESS=\$DBUS_SESSION_BUS_ADDRESS\"
        echo \"PATH=\$PATH\"
        echo \"DOCKER_HOST=\$DOCKER_HOST\"
	
	systemctl --user daemon-reload > /dev/null 2>&1
        systemctl --user enable docker > /dev/null 2>&1
        systemctl --user start docker > /dev/null 2>&1

	systemctl --user status > /dev/null 2>&1
    ' 2>/dev/null
"

#fork/exec /proc/self/exe: operation not permitted

# we should check with: systemctl --user status


	else

		
cat <<EOT | tee "/etc/apparmor.d/home.$username.bin.rootlesskit" > /dev/null 2>&1
# ref: https://ubuntu.com/blog/ubuntu-23-10-restricted-unprivileged-user-namespaces
abi <abi/4.0>,
include <tunables/global>

/home/$username/bin/rootlesskit flags=(unconfined) {
userns,

# Site-specific additions and overrides. See local/README for details.
include if exists <local/home.$username.bin.rootlesskit>
}
EOT

filename=$(echo $HOME/bin/rootlesskit | sed -e s@^/@@ -e s@/@.@g)

cat <<EOF > ~/${filename} 2>/dev/null
abi <abi/4.0>,
include <tunables/global>

"$HOME/bin/rootlesskit" flags=(unconfined) {
userns,

include if exists <local/${filename}>
}
EOF

  
  		mv ~/${filename} /etc/apparmor.d/${filename} > /dev/null 2>&1

		systemctl restart apparmor.service >/dev/null 2>&1
		loginctl enable-linger $username >/dev/null 2>&1
		mkdir -p /home/$username/.docker/run /home/$username/bin/ /home/$username/bin/.config/systemd/user/ >/dev/null 2>&1
		chmod 700 /home/$username/.docker/run >/dev/null 2>&1
		chmod 755 -R /home/$username/ >/dev/null 2>&1
		chown -R $username:$username /home/$username/ >/dev/null 2>&1

      		system_wide_rootless_script="/etc/openpanel/docker/dockerd-rootless-setuptool.sh"
      		if [ ! -f "$system_wide_rootless_script" ]; then
			curl -sSL https://get.docker.com/rootless -o $system_wide_rootless_script
   			chmod +x $system_wide_rootless_script
		fi
  
  		ln -sf $system_wide_rootless_script /home/$username/bin/dockerd-rootless-setuptool.sh

machinectl shell $username@ /bin/bash -c "
    # Setup environment for rootless Docker
    source ~/.bashrc

    /home/$username/bin/dockerd-rootless-setuptool.sh install >/dev/null 2>&1

    echo 'export XDG_RUNTIME_DIR=/home/$username/.docker/run' >> ~/.bashrc
    echo 'export PATH=/home/$username/bin:\$PATH' >> ~/.bashrc
    echo 'export DOCKER_HOST=unix:///home/$username/.docker/run/docker.sock' >> ~/.bashrc

    source ~/.bashrc
    mkdir -p ~/.config/systemd/user/

    cat <<EOF > ~/.config/systemd/user/docker.service
[Unit]
Description=Docker Application Container Engine (Rootless)
After=network.target

[Service]
Environment=PATH=/home/$username/bin:\$PATH
Environment=DOCKER_HOST=unix://%t/docker.sock
ExecStart=/home/$username/bin/dockerd-rootless.sh
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload > /dev/null 2>&1
    systemctl --user restart docker > /dev/null 2>&1
" 2>/dev/null
	fi
}



run_docker() {

# TODO:
# check ports on remote server!
#
    # added in 0.2.3 to set fixed ports for mysql and ssh services of the user!
    find_available_ports() {

	last_user=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -se "SELECT server FROM users ORDER BY id DESC LIMIT 1;")

	if [[ -n "$last_user" ]]; then
		env_file="/home/$last_user/.env"
	
		if [[ ! -f "$env_file" ]]; then
		    ###echo "Warning: .env file is missing for existing users, port generation may be a bit off - make sure to check ports assigned to user afterwards."
      		    min_port="32768"
		fi
		
		highest_port=$(grep -E '^[A-Z_]+_PORT=' "$env_file" | grep -oE '[0-9]+' | sort -nr | head -n 1)
		
		if [[ -z "$highest_port" ]]; then
			###echo "ERROR: No ports found in the .env file."
   			min_port="32768"
      		else
			min_port=${highest_port}
		fi
	else
		# no users yet! 
      		min_port="32768"
	fi      
      
      
      local found_ports=()
                  
        if [ -n "$node_ip_address" ]; then
            # TODO: Use a custom user or configure SSH instead of using root
            ssh $key_flag "root@$node_ip_address" '
		declare -a found_ports=()
		for ((i=1; i<=7; i++)); do
		    port=$((min_port + i))
		    found_ports+=("$port")
		done
		        
              echo "${found_ports[@]}"
            '
        else
		declare -a found_ports=()
		for ((i=1; i<=7; i++)); do
		    port=$((min_port + i))
		    found_ports+=("$port")
		done
            echo "${found_ports[@]}"
        fi



    }

	    
	validate_port() {
	    local port="$1"
	    port="$(echo "$1" | tr -d '[:space:]')"  # Trim spaces
	
	    if [[ "$port" =~ ^[0-9]+$ ]]; then
	        return 0  # Port is valid
	    else
	        echo "DEBUG: Invalid port detected: $port"
	        echo "DEBUG: Available ports: $AVAILABLE_PORTS"
	        return 1  # Port is invalid
	    fi
	}

    # Find available ports
    log "Checking available ports to use for the docker container"
    AVAILABLE_PORTS=$(find_available_ports)

    # Split the ports into variables
	FIRST_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $1}')
	SECOND_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $2}')
	THIRD_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $3}')
	FOURTH_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $4}')
	FIFTH_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $5}')
        SIXTH_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $6}')
	SEVENTH_NEXT_AVAILABLE=$(echo $AVAILABLE_PORTS | awk '{print $7}')

    # todo: better validation!
    if validate_port "$FIRST_NEXT_AVAILABLE" && validate_port "$SECOND_NEXT_AVAILABLE" && validate_port "$THIRD_NEXT_AVAILABLE" && validate_port "$FOURTH_NEXT_AVAILABLE" && validate_port "$FIFTH_NEXT_AVAILABLE" && validate_port "$SIXTH_NEXT_AVAILABLE" && validate_port "$SEVENTH_NEXT_AVAILABLE"; then
	port_1="$FIRST_NEXT_AVAILABLE:22"
	port_2="$SECOND_NEXT_AVAILABLE:3306"
	port_3="$THIRD_NEXT_AVAILABLE:7681"
	port_4="$FOURTH_NEXT_AVAILABLE:80"
	port_5="127.0.0.1:$FIFTH_NEXT_AVAILABLE:80"
        port_6="127.0.0.1:$SIXTH_NEXT_AVAILABLE:443"
	port_7="127.0.0.1:$SEVENTH_NEXT_AVAILABLE:80"
    else
	port_1=""
	port_2=""
	port_3=""
	port_4=""
	port_5=""
	port_6=""
 	port_7=""
    fi


cp /etc/openpanel/docker/compose/1.0/docker-compose.yml /home/$username/docker-compose.yml

if [ ! -f /home/$username/docker-compose.yml ]; then
  echo "ERROR: /home/$username/docker-compose.yml file not created. Make sure that the /etc/openpanel/ is updated and contains valid templates."
  exit 1
fi

# TEMPLATE: https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/1.0/.env

postgres_password=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9')
mysql_root_password=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9')
pg_admin_password=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9')

: '
    log "Using:"
    log ""
    log "USERNAME: $username"
    log "USER_ID: $user_id"
    log "CONTEXT: $username"
    log "TOTAL_CPU: $cpu"
    log "TOTAL_RAM: $ram"
    log "HTTP_PORT: $port_5"
    log "HTTPS_PORT: $port_6"
    log "HOSTNAME: $hostname"
    log "UNUSED_1_PORT: $port_1"
    log "UNUSED_2_PORT: $port_3"
    log "PMA_PORT: $port_4"
    log "MYSQL_PORT: $port_2"
    log "DEFAULT_PHP_VERSION: $default_php_version"
    log "POSTGRES_PASSWORD: $postgres_password"
    log "MYSQL_ROOT_PASSWORD: $mysql_root_password"
'

if [ -z "$username" ] || [ -z "$user_id" ] || [ -z "$cpu" ] || [ -z "$ram" ] || [ -z "$port_5" ] || [ -z "$port_6" ] || [ -z "$port_7" ] || [ -z "$port_1" ] || [ -z "$port_3" ] || [ -z "$port_4" ] || [ -z "$port_2" ] || [ -z "$default_php_version" ] || [ -z "$postgres_password" ] || [ -z "$mysql_root_password" ]; then
   echo "ERROR: One or more required variables are not set."
   exit 1
fi

cp /etc/openpanel/docker/compose/1.0/.env /home/$username/.env

sed -i -e "s|USERNAME=\"[^\"]*\"|USERNAME=\"$username\"|g" \
    -e "s|USER_ID=\"[^\"]*\"|USER_ID=\"$user_id\"|g" \
    -e "s|CONTEXT=\"[^\"]*\"|CONTEXT=\"$username\"|g" \
    -e "s|TOTAL_CPU=\"[^\"]*\"|TOTAL_CPU=\"$cpu\"|g" \
    -e "s|TOTAL_RAM=\"[^\"]*\"|TOTAL_RAM=\"$ram\"|g" \
    -e "s|^HTTP_PORT=\"[^\"]*\"|HTTP_PORT=\"$port_5\"|g" \
    -e "s|HTTPS_PORT=\"[^\"]*\"|HTTPS_PORT=\"$port_6\"|g" \
    -e "s|UNUSED_1_PORT=\"[^\"]*\"|UNUSED_1_PORT=\"127.0.0.1:$port_1\"|g" \
    -e "s|UNUSED_2_PORT=\"[^\"]*\"|UNUSED_2_PORT=\"$port_3\"|g" \
    -e "s|PMA_PORT=\"[^\"]*\"|PMA_PORT=\"$port_4\"|g" \
    -e "s|POSTGRES_PASSWORD=\"[^\"]*\"|POSTGRES_PASSWORD=\"$postgres_password\"|g" \
    -e "s|PGADMIN_PW=\"[^\"]*\"|PGADMIN_PW=\"$pg_admin_password\"|g" \
    -e "s|OPENSEARCH_INITIAL_ADMIN_PASSWORD=\"[^\"]*\"|OPENSEARCH_INITIAL_ADMIN_PASSWORD=\"$pg_admin_password\"|g" \
    -e "s|PGADMIN_MAIL=\"[^\"]*\"|PGADMIN_MAIL=\"$email\"|g" \
    -e "s|MYSQL_PORT=\"[^\"]*\"|MYSQL_PORT=\"127.0.0.1:$port_2\"|g" \
    -e "s|DEFAULT_PHP_VERSION=\"[^\"]*\"|DEFAULT_PHP_VERSION=\"$default_php_version\"|g" \
    -e "s|MYSQL_ROOT_PASSWORD=\"[^\"]*\"|MYSQL_ROOT_PASSWORD=\"$mysql_root_password\"|g" \
    -e "s|PROXY_HTTP_PORT=\"[^\"]*\"|#PROXY_HTTP_PORT=\"$port_7\"|g" \
    "/home/$username/.env"

if [[ -n "$webserver" ]]; then
    # Check for varnish+nginx or varnish+apache, and extract the web server part
    if [[ "$webserver" =~ ^varnish\+([a-zA-Z]+)$ ]]; then
        webserver="${BASH_REMATCH[1]}"  # Extract the part after varnish+
        log "Setting varnish caching and $webserver as webserver for the user.."
        sed -i -e "s|WEB_SERVER=\"[^\"]*\"|WEB_SERVER=\"$webserver\"|g" \
	-e "s|^#PROXY_HTTP_PORT|PROXY_HTTP_PORT|g" "/home/$username/.env"
    elif [[ "$webserver" =~ ^(nginx|apache|openresty)$ ]]; then
        log "Setting $webserver as webserver for the user.."
        sed -i -e "s|WEB_SERVER=\"[^\"]*\"|WEB_SERVER=\"$webserver\"|g" "/home/$username/.env"
        VARNISH=false
    else
        log "Warning: invalid webserver type selected: $webserver. Must be 'nginx', 'apache', 'openresty', 'varnish+nginx', 'varnish+apache' or 'varnish+openresty'. Using the default instead.."
    fi
fi


if [[ -n "$sql_type" ]]; then
    if [[ "$sql_type" =~ ^(mysql|mariadb)$ ]]; then
        log "Setting $sql_type as MySQL server type for the user.."
        sed -i -e "s|MYSQL_TYPE=\"[^\"]*\"|MYSQL_TYPE=\"$sql_type\"|g" \
            "/home/$username/.env"
    else
        log "Warning: Invalid SQL server type selected: $sql_type. Must be 'mysql' or 'mariadb'. Using the default instead.."
    fi
fi


mkdir -p /home/$username/sockets/{mysqld,postgres,redis,memcached}
cp /etc/openpanel/mysql/user.cnf /home/${username}/custom.cnf
cp /etc/openpanel/nginx/user-nginx.conf /home/$username/nginx.conf  # added in 1.2.2
cp /etc/openpanel/openresty/nginx.conf /home/$username/openresty.conf
cp /etc/openpanel/apache/httpd.conf /home/$username/httpd.conf
cp /etc/openpanel/varnish/default.vcl /home/$username/default.vcl
cp /etc/openpanel/ofelia/users.ini /home/$username/crons.ini  > /dev/null 2>&1 # added in 1.2.1
cp /etc/openpanel/backups/backup.env /home/$username/backup.env  > /dev/null 2>&1 # added in 1.1.7
cp /etc/openpanel/mysql/phpmyadmin/pma.php /home/$username/pma.php  > /dev/null 2>&1 # added in 1.4.4 for autologin to pma

cp -r /etc/openpanel/php/ini /home/${username}/php.ini
chown -R $username:$username /home/$username/sockets
chmod 777 /home/$username/sockets/


if [ ! -f "/home/$username/.env" ]; then
   echo "ERROR: Failed to create .env file! Make sure that the /etc/openpanel/ is updated and contains valid templates."
   exit 1
fi

echo "[client]
user=root
password="$mysql_root_password"
" > /home/$username/my.cnf

if [ ! -f "/home/$username/my.cnf" ]; then
   echo "WARNING: Failed to create my.cnf file! Make sure that the /etc/openpanel/ is updated and contains valid templates."
fi

}





copy_skeleton_files() {
    log "Creating configuration files for the newly created user"
    
	rm -rf /etc/openpanel/skeleton/domains > /dev/null 2>&1 #todo remove from 1.0.0!
        cp -r /etc/openpanel/skeleton/ /etc/openpanel/openpanel/core/users/$username/  > /dev/null 2>&1

	if [ -n "$node_ip_address" ]; then
		# adding dedicated ip to be used in caddy and dns files
		local json_file="/etc/openpanel/openpanel/core/users/$username/ip.json"
		echo "{ \"ip\": \"$node_ip_address\" }" > "$json_file"
	fi
 
        opencli php-available_versions $username  > /dev/null 2>&1 &
}



get_php_version() {
    default_php_version=$(grep '^DEFAULT_PHP_VERSION=' /etc/openpanel/docker/compose/1.0/.env | sed -E 's/^DEFAULT_PHP_VERSION="?([^"]*)"?/\1/')

    if [ -z "$default_php_version" ]; then
      if [ "$DEBUG" = true ]; then
        echo "Default PHP version not found in .env file, using the fallback default version.."
      fi
      default_php_version="8.4"
    fi
}




start_panel_service() {
	# from 0.2.5 panel service is not started until acc is created
	log "Checking if OpenPanel service is already running, or starting it in background.."
        nohup sh -c "cd /root && docker compose up -d openpanel" </dev/null >nohup.out 2>nohup.err &
}


create_context() {

    if [ -n "$node_ip_address" ]; then

	docker context create $username \
	  --docker "host=ssh://$username" \
	  --description "$username"
   else
   	docker context create $username \
			  --docker "host=unix:///hostfs/run/user/${user_id}/docker.sock" \
		   	  --description "$username"
   fi
}

test_compose_command_for_user() {

    # Check if Docker Compose is working
    if ! docker --context=$username compose version >/dev/null 2>&1; then
        echo "[✘] Error: Docker Compose is not working in this context. User creation failed."
	hard_cleanup # remove data!
        exit 1
    fi
   
}

save_user_to_db() {
    log "Saving new user to database"
    
     # Insert data into MySQL database
	if [ -n "$reseller" ]; then
   	    mysql_query="INSERT INTO users (username, password, owner, email, plan_id, server) VALUES ('$username', '$hashed_password', '$reseller', '$email', '$plan_id', '$username');"
	else
	    mysql_query="INSERT INTO users (username, password, email, plan_id, server) VALUES ('$username', '$hashed_password', '$email', '$plan_id', '$username');"
	fi    
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$mysql_query"
    
    if [ $? -eq 0 ]; then
        echo "[✔] Successfully added user $username with password: $password"
    else
        echo "[✘] Error: Data insertion failed."
	hard_cleanup # remove data!
        exit 1
    fi

}

create_backup_dirs_for_each_index() {
    log "Creating backup jobs directories for the user"
    for dir in /etc/openpanel/openadmin/config/backups/index/*/; do
      mkdir -p "${dir}${username}"
    done
}




send_email_to_new_user() {
    if $SEND_EMAIL; then
        echo "Sending email to $email with login information"
            # Check if the provided email is valid
            if [[ $email =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                . "$SEND_EMAIL_FILE"
                email_notification "New OpenPanel account information" "OpenPanel URL: $login_url | username: $username  | password: $password"
		# todo: check nodeip, send it in email!
            else
                echo "$email is not a valid email address. Login infomration can not be sent to the user."
            fi       
    fi
}


reload_user_quotas() {
    quotacheck -avm >/dev/null 2>&1
    repquota -u / > /etc/openpanel/openpanel/core/users/repquota 
}



generate_user_password_hash() {
    if [ "$password" = "generate" ]; then
        password=$(openssl rand -base64 12)
        log "Generated password: $password" 
    fi

    hashed_password=$("/usr/local/admin/venv/bin/python3" -c "from werkzeug.security import generate_password_hash; print(generate_password_hash('$password'))")
}


collect_stats() {
  local file="/etc/openpanel/openpanel/core/users/$username/docker_usage.txt"
  local timestamp=$(date '+%Y-%m-%d-%H-%M-%S')
  local data='{"BlockIO":"0B / 0B","CPUPerc":"0 %","Container":"0","ID":"","MemPerc":"0 %","MemUsage":"0MiB / 0MiB","Name":"","NetIO":"0 / 0","PIDs":0}'
  echo "$timestamp $data" > "$file"
}


##########################################################
########################## MAIN ##########################
##########################################################
(
flock -n 200 || { echo "[✘] Error: A user creation process is already running."; echo "Please wait for it to complete before starting a new one. Exiting."; exit 1; }
get_hostname_of_master                       # later can be overwritten if slave
check_username_is_valid                      # validate username first
validate_password_in_lists $password         # compare with weakpass dictionaries
get_slave_if_set                             # check if slave should be used and test connection
get_existing_users_count                     # list users from db
get_plan_info_and_check_requirements         # list plan from db and check available resources
check_if_reseller                            # if reseller, check limits
print_debug_info_before_starting_creation    # print debug info
validate_ssh_login                           # test ssh logins for cluster member
create_user_set_quota_and_password           # create user
sshfs_mounts                                 # mount /home/user
setup_ssh_key                                # set key for the user
docker_rootless                              # install 
docker_compose                               # magic happens here
create_context                               # on master
test_compose_command_for_user
get_php_version                              # must be before run_docker !
run_docker                                   # run docker container
reload_user_quotas                           # refresh their quotas
generate_user_password_hash
copy_skeleton_files                          # get webserver, php version and mysql type for user
create_backup_dirs_for_each_index            # added in 0.3.1 so that new users immediately show with 0 backups in :2087/backups#restore
start_panel_service                          # start user panel if not running
save_user_to_db                              # save user to mysql db
collect_stats                                # must be after insert in db
send_email_to_new_user                       # added in 0.3.2 to optionally send login info to new user
)200>/var/lock/openpanel_user_add.lock

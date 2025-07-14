#!/bin/bash
################################################################################
# Script Name: ftp/add.sh
# Description: Create FTP sub-user for openpanel user.
# Usage: opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 22.05.2024
# Last Modified: 13.07.2025
# Company: openpanel.co
# Copyright (c) openpanel.co
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

if [ "$#" -lt 4 ] || [ "$#" -gt 5 ]; then
    echo "Usage: opencli ftp-add <new_username> <new_password> '<directory>' <openpanel_username> [--debug]"
    exit 1
fi

username="${1,,}"
password="$2"
directory="$3"
openpanel_username="$4"
DEBUG=false  # Default value for DEBUG


# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done




check_and_start_ftp_server(){
	if [ -n "$(docker ps -q -f name=openadmin_ftp)" ]; then
	    :
	else
	    cd /root && docker --context default compose up -d ftp_env_generator >/dev/null 2>&1
	    cd /root && docker --context default compose up -d openadmin_ftp >/dev/null 2>&1
	fi
}



get_docker_context_for_user(){
    context=$(mysql -e "SELECT server FROM users WHERE username='$openpanel_username';" -N)   
    if [ -z "$context" ]; then
        echo "ERROR: No context found for user '$openpanel_username'. Aborting!"
        exit 1
    fi    
}

# Function to read users from users.list files and create them
create_user() {
	get_docker_context_for_user
 
	real_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/"
	relative_path="${directory##/var/www/html/}"
	new_directory="${real_path}${relative_path}"
 
	GID=$(grep "$openpanel_username" /hostfs/etc/group | cut -d: -f3)
	GROUP=$(docker exec openadmin_ftp sh -c  "getent group \"$GID\" | cut -d: -f1")
	if [ -n "$GROUP" ]; then
	    GROUP_OPT="-g $GROUP"
	elif [ -n "$GID" ]; then
	    addgroup -g "$GID" "$openpanel_username"
	    GROUP_OPT="-g $openpanel_username"
	    chmod +rx "/home/$openpanel_username"
	    chmod +rx "/home/$openpanel_username/docker-data"
	    chmod +rx "/home/$openpanel_username/docker-data/volumes"
	    chmod +rx "/home/$openpanel_username/docker-data/volumes/${openpanel_username}_html_data"
	    chmod +rx "/home/$openpanel_username/docker-data/volumes/${openpanel_username}_html_data/_data"
	fi

	PYTHON_PATH=$(which python3 || echo "/usr/local/bin/python")
	
	HASHED_PASS=$($PYTHON_PATH -W ignore -c "import crypt, random, string; salt = ''.join(random.choices(string.ascii_letters + string.digits, k=16)); print(crypt.crypt('$password', '\$6\$' + salt))")

 	# Create user without password
	docker exec openadmin_ftp sh -c "adduser -h '${new_directory}' -s /sbin/nologin ${GROUP_OPT} --disabled-password --gecos '' '${username}'" >/dev/null 2>&1


	# Set the hashed password
	if docker exec openadmin_ftp sh -c "usermod -p '$HASHED_PASS' '$username'"; then
	    mkdir -p "/hostfs$new_directory"
	    chown "${openpanel_username}:${openpanel_username}" "/hostfs$new_directory"
	    chmod +rx "/hostfs/home/$openpanel_username"
	    chmod +rx "/hostfs/home/$openpanel_username/docker-data"
	    chmod +rx "/hostfs/home/$openpanel_username/docker-data/volumes"
	    chmod +rx "/hostfs/home/$openpanel_username/docker-data/volumes/${openpanel_username}_html_data"
	    chmod +rx "/hostfs/home/$openpanel_username/docker-data/volumes/${openpanel_username}_html_data/_data"
	
	    USER_UID=$(docker exec openadmin_ftp id -u "$username")
	    USER_GID=$(docker exec openadmin_ftp id -g "$username")
	
	    echo "$username|$HASHED_PASS|$directory|$USER_UID|$USER_GID" >> "/etc/openpanel/ftp/users/${openpanel_username}/users.list"
	    echo "Success: FTP user '$username' created successfully (UID: $USER_UID, GID: $USER_GID)."
	else
	    if [ "$DEBUG" = true ]; then
	        echo "ERROR: Failed to create FTP user with command:"     
	        echo ""
	        echo "docker exec openadmin_ftp sh -c 'adduser -h $new_directory -s /sbin/nologin $username'"
	        echo ""
	        echo "Run the command manually to check for errors."
	    else
	        echo "ERROR: Failed to create FTP user. To debug run this command on terminal: opencli ftp-add $username $password '$new_directory' $openpanel_username --debug"  
	    fi
	    exit 1
	fi
}


validate_data() {
	# user.openpanel_username
	if [[ ! $username == *".${openpanel_username}" ]]; then
	    echo "ERROR: FTP username must end with openpanel username, example: '$username.$openpanel_username'"
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
	

	if [[ "$directory" != /var/www/html/* ]]; then
	echo "ERROR: Invalid path. It must start with /var/www/html/"
	exit 1
	fi
	
	if [[ "$directory" == *".."* ]]; then
	echo "ERROR: Path traversal detected (.. is not allowed)."
	exit 1
	fi
	
	if [[ "$directory" == *"//"* ]]; then
	echo "ERROR: Double slashes are not allowed in the path."
	exit 1
	fi
	
	if [[ "$directory" == *"|"* ]]; then
	echo "ERROR: The path cannot contain the '|' character."
	exit 1
	fi


	# Check if password length is at least 8 characters
	if [ ${#password} -lt 8 ]; then
	    echo "ERROR: Password is too short. It must be at least 8 characters long."
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
	
	# Check if password contains at least one uppercase letter
	if ! [[ $password =~ [A-Z] ]]; then
	    echo "ERROR: Password must contain at least one uppercase letter."
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
	
	# Check if password contains at least one lowercase letter
	if ! [[ $password =~ [a-z] ]]; then
	    echo "ERROR: Password must contain at least one lowercase letter."
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
	
	# Check if password contains at least one digit
	if ! [[ $password =~ [0-9] ]]; then
	    echo "ERROR: Password must contain at least one digit."
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
	
	# Check if password contains at least one special character
	if ! [[ $password =~ [[:punct:]] ]]; then
	    echo "ERROR: Password must contain at least one special character."
	    echo "       docs: https://openpanel.com/docs/articles/accounts/forbidden-usernames/#ftp"
	    exit 1
	fi
}

# check if ftp user exists
check_user_exists() {
	if grep -Fq "$username|" "/etc/openpanel/ftp/users/${openpanel_username}/users.list"; then
	    echo "ERROR: FTP User '$username' already exists."
	    exit 1
	fi
}

make_dirs() {
	mkdir -p "/etc/openpanel/ftp/users/${openpanel_username}"
	touch "/etc/openpanel/ftp/users/${openpanel_username}/users.list"
}



# main
validate_data                 # check username, path and password
make_dirs		      # create dir for user and users file
check_user_exists             # check user exists
check_and_start_ftp_server    # start ftpserver if not running
create_user                   # create new user
exit 0

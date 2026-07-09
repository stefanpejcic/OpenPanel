#!/bin/bash
################################################################################
# Script Name: ftp/add.sh
# Description: Create FTP sub-user for openpanel user.
# Usage: opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 22.05.2024
# Last Modified: 08.07.2026
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

if [ "$#" -lt 4 ] || [ "$#" -gt 5 ]; then
    echo "Usage: opencli ftp-add <new_username> <new_password> '<directory>' <openpanel_username> [--debug]"
    exit 1
fi

username="${1,,}"
password="$2"
directory="$3"
openpanel_username="$4"
DEBUG=false  # Default value for DEBUG

source /usr/local/opencli/lib/password_strength.sh


# Parse optional flags to enable debug mode when needed!
for arg in "$@"; do
    case $arg in
        --debug) DEBUG=true ;;
        *) ;;
    esac
done




check_and_start_ftp_server(){
	if [ -z "$(docker ps -q -f name=openadmin_ftp)" ]; then
	    cd /root && docker --context default compose up -d openadmin_ftp >/dev/null 2>&1
	fi
}


# Function to read users from users.list files and create them
create_user() {
    real_path="/home/${context}/docker-data/volumes/${context}_html_data/_data/"
    relative_path="${directory##/var/www/html/}"
    new_directory="${real_path}${relative_path}"

    GID=$(stat -c '%u' "/home/$context")
    if [[ ! "$GID" =~ ^[0-9]+$ ]]; then
        echo "ERROR: Could not determine GID for '$context'."
        exit 1
    fi

    EXISTING_GROUP=$(docker exec openadmin_ftp sh -c "getent group '$GID' | cut -d: -f1")
    if [[ -z "$EXISTING_GROUP" ]]; then
        docker exec openadmin_ftp addgroup -g "$GID" "$context"
    fi

    mkdir -p "$new_directory"
    chmod +rx "/home/$context"
    chmod +rx "/home/$context/docker-data"
    chmod +rx "/home/$context/docker-data/volumes"
    chmod +rx "/home/$context/docker-data/volumes/${context}_html_data"
    chmod +rx "/home/$context/docker-data/volumes/${context}_html_data/_data"

    PYTHON_PATH=$(which python3 || echo "/usr/local/bin/python3")

    HASHED_PASS=$($PYTHON_PATH -c "
import sys, hashlib, os, base64
pw = sys.argv[1]
salt = base64.b64encode(os.urandom(12)).decode()[:16]
import crypt
print(crypt.crypt(pw, '\$6\$' + salt))
" "$password" 2>/dev/null || $PYTHON_PATH -c "
import sys, os, hashlib
pw = sys.argv[1].encode()
salt = os.urandom(8)
import hashlib
h = hashlib.sha512(pw + salt).hexdigest()
print('\$6\$' + salt.hex() + '\$' + h)
" "$password")

    docker exec openadmin_ftp useradd -d "${new_directory}" -s /sbin/nologin -g "$context" -M "$username" --badname

    if docker exec openadmin_ftp sh -c "usermod -p '$HASHED_PASS' '$username'"; then
        mkdir -p "$new_directory"
        chown -R "$GID:$GID" "$new_directory"
        chmod -R 2775 "$new_directory"
        chmod +rx "/home/$context"
        chmod +rx "/home/$context/docker-data"
        chmod +rx "/home/$context/docker-data/volumes"
        chmod +rx "/home/$context/docker-data/volumes/${context}_html_data"
        chmod +rx "/home/$context/docker-data/volumes/${context}_html_data/_data"

        USER_UID=$(docker exec openadmin_ftp id -u "$username")
        USER_GID=$(docker exec openadmin_ftp id -g "$username")

        echo "$username|$HASHED_PASS|$directory|$USER_UID|$USER_GID" >> "/etc/openpanel/ftp/users/${context}/users.list"

        nohup opencli sentinel --action=ftp_create --title="FTP account created" --message="New FTP account has been created for OpenPanel user: '$openpanel_username'. Directory: $directory | UID: $USER_UID | GID: $USER_GID" >/dev/null 2>&1 &
        disown

        echo "Success: FTP user '$username' created successfully (UID: $USER_UID, GID: $USER_GID)."
    else
        if [ "$DEBUG" = true ]; then
            echo "ERROR: Failed to create FTP user with command:"
            echo "docker exec openadmin_ftp useradd -d $new_directory -s /sbin/nologin -g $context $username"
        else
            echo "ERROR: Failed to create FTP user. To debug, run: opencli ftp-add $username $password '$new_directory' $context --debug"
        fi
        exit 1
    fi
}






validate_data() {
	# user.openpanel_username
	if [[ $username != *@* ]]; then
	    echo "ERROR: FTP username must be in format: username@domain"
	    exit 1
	fi

    local_part="${username%%@*}"
    domain_part="${username##*@}"

    if [[ ! "$local_part" =~ ^[a-z0-9_-]+$ ]]; then
        echo "ERROR: FTP username before the domain part may only contain a-z, 0-9, _ and -"
        exit 1
    fi

    if [[ ! "$domain_part" =~ ^[a-z0-9._-]+$ ]]; then
        echo "ERROR: Invalid domain provided in FTP username."
        exit 1
    fi

	if [[ ! "$openpanel_username" =~ ^[a-z0-9_-]+$ ]]; then
	    echo "ERROR: Invalid OpenPanel username."
	    exit 1
	fi

	context=$(mysql -N -e "SELECT u.server FROM users u WHERE u.username='${openpanel_username}' AND EXISTS (SELECT 1 FROM domains d WHERE d.domain_url = '${domain_part}' AND d.user_id = u.id);")
    if [ -z "$context" ]; then
        echo "ERROR: No context found for user '$openpanel_username' - or does not own the domain name. Aborting!"
        exit 1
    fi

	if [[ "$directory" != /var/www/html/* ]]; then
		echo "ERROR: Invalid path. It must start with /var/www/html/"
		exit 1
	fi

	if [[ "$directory" =~ [*?\[\]] ]]; then
	    echo "ERROR: Path cannot contain glob characters."
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

	# Check against the admin-configured strength threshold
	require_password_strength "$password"
}

# check if ftp user exists
check_user_exists() {
	if grep -Fq "$username|" "/etc/openpanel/ftp/users/${context}/users.list"; then
	    echo "ERROR: FTP User '$username' already exists."
	    exit 1
	fi
}

make_dirs() {
	mkdir -p "/etc/openpanel/ftp/users/${context}"
	touch "/etc/openpanel/ftp/users/${context}/users.list"
}



# main
validate_data                 # check username, path and password
make_dirs		              # create dir for user and users file
check_user_exists             # check user exists
check_and_start_ftp_server    # start ftpserver if not running
create_user                   # create new user
exit 0

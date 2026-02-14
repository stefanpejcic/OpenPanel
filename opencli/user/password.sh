#!/bin/bash
################################################################################
# Script Name: user/password.sh
# Description: Reset password for a user.
# Usage: opencli user-password <USERNAME> <NEW_PASSWORD | random>
# Docs: https://docs.openpanel.com
# Author: Stefan Pejcic
# Created: 30.11.2023
# Last Modified: 13.02.2026
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
# Variables
username=$1
new_password=$2
random_flag=false


# ======================================================================
# Validations
if [ $# -ne 2 ]; then
    echo "Usage: opencli user-password <USERNAME> <NEW_PASSWORD | random>"
    exit 1
fi


# ======================================================================
# Helpers
generate_random_password() {
    tr -dc 'a-zA-Z0-9' < /dev/urandom | head -c 12
}

determine_python_path() {
    if [ -x /usr/local/admin/venv/bin/python3 ]; then
        PYTHON_EXEC=/usr/local/admin/venv/bin/python3
    elif command -v python3 &>/dev/null; then
        PYTHON_EXEC=python3
    else
        echo "Warning: No Python 3 interpreter found. Please install Python 3 or check the virtual environment."
        exit 1
    fi
}

generate_and_hash_password() {

    if [ "$new_password" == "random" ]; then
        new_password=$(generate_random_password)
        random_flag=true
    fi

hashed_password=$(
"$PYTHON_EXEC" - "$new_password" << 'EOF'
import sys
from werkzeug.security import generate_password_hash
print(generate_password_hash(sys.argv[1]))
EOF
)

# added in 1.6.8 to handle ' and " in passwords
escaped_hash=$(printf "%s" "$hashed_password" | sed "s/'/''/g") 
}

save_to_database() {
    mysql_query="UPDATE users SET password='$escaped_hash' WHERE username='$username';"
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$mysql_query"
    
    if [ $? -eq 0 ]; then
        delete_sessions_query="DELETE FROM active_sessions WHERE user_id=(SELECT id FROM users WHERE username='$username');"
        mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$delete_sessions_query"
        if [ $? -ne 0 ]; then
            echo "WARNING: Failed to terminate existing sessions for the user."
        fi

        echo "Successfully changed password for user $username$([ "$random_flag" = true ] && echo ", new random generated password is: $new_password")"
    else
        echo "Error: Data insertion failed."
        exit 1
    fi
}


# ======================================================================
# Main
determine_python_path
generate_and_hash_password
source /usr/local/opencli/db.sh
save_to_database

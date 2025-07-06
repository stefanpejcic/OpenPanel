#!/bin/bash
################################################################################
# Script Name: user/change_email.sh
# Description: Change email for user
# Usage: opencli user-email <USERNAME> <NEW_EMAIL>
# Docs: https://docs.openpanel.com
# Author: Radovan Jecmenica
# Created: 06.12.2023
# Last Modified: 04.07.2025
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
# Source the database connection script
source /usr/local/opencli/db.sh

# Function to change email in the database
change_email_in_db() {
    USERNAME=$1
    NEW_EMAIL=$2
    
    # Update the username in the database with the suspended prefix
    mysql_query="UPDATE users SET email='$NEW_EMAIL' WHERE username='$USERNAME';"
    
    mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$mysql_query"
}

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <USERNAME> <NEW_EMAIL>"
  exit 1
fi

# Extract arguments
USERNAME=$1
NEW_EMAIL=$2

# Call the function to change email in the database
change_email_in_db "$USERNAME" "$NEW_EMAIL"

# Check if the function executed successfully
if [ $? -eq 0 ]; then
  echo "Email for user $USERNAME updated to $NEW_EMAIL."
else
  echo "Error: Failed to update email for user $USERNAME."
fi

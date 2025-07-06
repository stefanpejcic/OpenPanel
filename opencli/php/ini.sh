#!/bin/bash
################################################################################
# Script Name: php/ini.sh
# Description: View or change php.ini values of any php version for a user.
# Usage: opencli php-ini <username> <action> <setting> [value]
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


# Check if username, action, setting, and value arguments are provided
if [ $# -lt 3 ]; then
    echo "Usage: openli php-ini <username> <action> <setting> [value]"
    exit 1
fi

# Assigning arguments to variables
USERNAME=$1
ACTION=$2
SETTING=$3
VALUE=$4

# Get PHP versions on the host machine
PHP_VERSIONS=$(opencli php-enabled_php_versions $USERNAME)


get_context_for_user() {
     source /usr/local/opencli/db.sh
        username_query="SELECT server FROM users WHERE username = '$USERNAME'"
        context=$(mysql -D "$mysql_database" -e "$username_query" -sN)
        if [ -z "$context" ]; then
            context=$USERNAME
        fi
}

get_context_for_user

# Loop through PHP versions based on the action
for VERSION in $PHP_VERSIONS; do
    # Strip "php" prefix from the version
    VERSION=$(echo "$VERSION" | sed 's/php//')

    # Construct PHP ini file path within the Docker container
    PHP_INI_PATH="/etc/php/$VERSION/fpm/php.ini"

    # Check if PHP ini file exists
    if docker --context $context exec -it $USERNAME [ -f "$PHP_INI_PATH" ]; then
        # Handle different actions: get or set
        case "$ACTION" in
            "get")
                # Display the current setting value
                CURRENT_VALUE=$(docker --context $context exec -it $USERNAME awk -F '=' '/^'"$SETTING"'[ \t]*=/ {gsub(/[ \t]+/, "", $2); print $2}' "$PHP_INI_PATH")
                echo "Current $SETTING setting in $PHP_INI_PATH for PHP version $VERSION: $CURRENT_VALUE"
                ;;
            "set")
                # Set the new value for the setting
                if [ -z "$VALUE" ]; then
                    echo "Error: Value is required for setting $SETTING in $PHP_INI_PATH for PHP version $VERSION"
                else
                    docker --context $context exec -it $USERNAME sed -i "s/;\{0,1\}$SETTING =.*/$SETTING = $VALUE/" "$PHP_INI_PATH"
                    echo "Setting $SETTING updated to $VALUE in $PHP_INI_PATH for PHP version $VERSION"

                    # Check the status of PHP-FPM service
                    SERVICE_STATUS=$(docker --context $context exec -it $USERNAME service php$VERSION-fpm status)

                    # If PHP-FPM service is active, restart it
                    if [[ $SERVICE_STATUS == *"active"* ]]; then
                        docker --context $context exec -it $USERNAME service php$VERSION-fpm restart
                        echo "PHP-FPM service restarted for PHP version $VERSION"
                    fi
                fi
                ;;
            *)
                echo "Error: Unknown action. Supported actions: get, set"
                exit 1
                ;;
        esac
    else
        echo "Error: PHP ini file not found for $USERNAME and PHP version $VERSION"
    fi
done

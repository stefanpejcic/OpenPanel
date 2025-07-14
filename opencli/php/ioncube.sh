#!/bin/bash
################################################################################
# Script Name: php/ioncube.sh
# Description: Enable IonCube Loader for every installed PHP version
# Usage: opencli php-ioncube <username>
# Author: Stefan Pejcic
# Created: 26.07.2024
# Last Modified: 13.07.2025
# Company: openpanel.com
# Copyright (c) Stefan Pejcic
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



# Check if the correct number of arguments are provided
if [ "$#" -ne 1 ]; then
  echo "Usage: opencli php-ioncube <username>"
  exit 1
fi

container_name="$1"


get_context_for_user() {
     source /usr/local/opencli/db.sh
        username_query="SELECT server FROM users WHERE username = '$container_name'"
        context=$(mysql -D "$mysql_database" -e "$username_query" -sN)
        if [ -z "$context" ]; then
            context=$container_name
        fi
}


get_context_for_user

# Check if the file exists and read the custom link if available
if [ -f /etc/openpanel/php/ioncube.txt ]; then
    custom_link=$(cat /etc/openapanel/php/inocube.txt)
    
    # Check if the content is a valid URL (starting with http or https)
    if [[ "$custom_link" =~ ^https?://.+ ]]; then
        download_link="$custom_link"
        download_message="Downloading ionCube loader extensions from custom link"
    else
        download_link="https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz"
        echo "Invalid URL in inocube.txt. Falling back to default: $download_link"
        download_message="Downloading latest ionCube loader extensions from default link"
    fi
else
    download_link="https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz"
    download_message="Downloading latest ionCube loader extensions from official website"
fi


file_name="ioncube_loaders.tar.gz"

# Step 1: Download the ionCube loaders tarball
echo "$download_message: $download_link"
docker --context $context exec -it "$container_name" bash -c "cd /tmp && wget -O $file_name $download_link > /dev/null 2>&1"

# Step 2: Uncompress the downloaded tarball
docker --context $context exec -it "$container_name" bash -c "cd /tmp && tar -xzf $file_name"

# Step 3: Copy the ionCube loader files to the appropriate directories
docker --context $context exec -it "$container_name" bash -c 'for dir in /usr/lib/php/20*/; do cp -r /tmp/ioncube/ioncube_loader_lin_*.so "$dir"; done'


echo "Listing installed PHP versions for user $container_name "
echo ""
# List php versions
php_versions=$(docker --context $context  exec -it "$container_name" update-alternatives --list php | awk -F'/' '{print $NF}' | grep -v 'default' | tr -d '\r')


# Process each PHP version
for php_version in $php_versions; do

  # Strip the 'php' part to get the version number
  php_version_number=$(echo "$php_version" | sed 's/php//')

  echo "### CHECKING PHP VERSION $php_version_number"
  # Check if already enabled
  if docker --context $context  exec -it "$container_name" bash -c "$php_version -m | grep -qi ioncube"; then
      echo "ionCube Loader is already enabled for $php_version"
      continue
  fi

  # Check if the ionCube loader file exists for this PHP version
  ioncube_file="/tmp/ioncube/ioncube_loader_lin_${php_version_number}.so"
  if docker --context $context exec -it "$container_name" test -f "$ioncube_file"; then
    echo "IonCube Loader extension is available for PHP version: $php_version_number - enabling.."
    # Function to add zend_extension line if it doesn't exist
    add_zend_extension_if_not_exists() {
      local ini_file="$1"
      local zend_extension_line="zend_extension=ioncube_loader_lin_${php_version_number}.so"
      
      # Check if zend_extension line exists, if not, append it
      if ! docker --context $context exec -it "$container_name" grep -q "^zend_extension=.*ioncube_loader_lin_${php_version_number}.so" "$ini_file"; then
        echo "$zend_extension_line" | docker --context $context exec -i "$container_name" tee -a "$ini_file" > /dev/null
      else
        docker --context $context exec -it "$container_name" sed -i "s|^zend_extension=.*ioncube_loader_lin_${php_version_number}.so|$zend_extension_line|" "$ini_file"
      fi
    }

    # Update CLI and FPM php.ini files
    add_zend_extension_if_not_exists "/etc/php/$php_version_number/cli/php.ini"
    add_zend_extension_if_not_exists "/etc/php/$php_version_number/fpm/php.ini"


    # Check if the PHP-FPM service is active
    service_status=$(docker --context $context exec -it "$container_name" sh -c "service php${php_version_number}-fpm status")
  
    if echo "$service_status" | grep -q "active (running)"; then
      echo "Restarting PHP-FPM service for PHP version $php_version_number"
      docker --context $context exec -it "$container_name" sh -c "service php${php_version_number}-fpm restart"
    fi

  else
    echo "ERROR: IonCube Loader extension ($ioncube_file) is not currently available for PHP version: $php_version_number"
    echo "       Please check manually on ioncube website if extension is available for your version: https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz"  
  fi
  
done
echo ""
echo "DONE"

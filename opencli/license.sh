#!/bin/bash
################################################################################
# Script Name: license.sh
# Description: Manage OpenPanel Enterprise license.
# Usage: opencli license verify 
# Author: Stefan Pejcic
# Created: 01.11.2023
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


CONFIG_FILE_PATH='/etc/openpanel/openpanel/conf/openpanel.config'
WHMCS_URL="https://my.openpanel.com/modules/servers/licensing/verify.php"

GREEN='\033[0;32m'
RED='\033[0;31m'
RESET='\033[0m'


# IP SERVERS
SCRIPT_PATH="/usr/local/admin/core/scripts/ip_servers.sh"
if [ -f "$SCRIPT_PATH" ]; then
    source "$SCRIPT_PATH"
else
    IP_SERVER_1=IP_SERVER_2=IP_SERVER_3="https://ip.openpanel.com"
fi


# Display usage information
usage() {
    echo "Usage: opencli license [options]"
    echo ""
    echo "Commands:"
    echo "  key                                           View current license key."
    echo "  enterprise-XXXXXXXXXX                         Save the license key."
    echo "  verify                                        Verify the license key."
    echo "  info                                          Display information about the license owner and expiration."
    echo "  delete                                        Delete the license key and downgrade OpenPanel to Community edition."
    exit 1
}



        # Check if --json flag is present
        if [[ " $@ " =~ " --json " ]]; then
          JSON="yes"
        else
          JSON="no"
        fi

        # Check if --json flag is present
        if [[ " $@ " =~ " --no-restart " ]]; then
          NO_RESTART="yes"
        else
          NO_RESTART="no"
        fi

# open conf file
read_config() {
    config=$(awk -F '=' '/\[LICENSE\]/{flag=1; next} /\[/{flag=0} flag{gsub(/^[ \t]+|[ \t]+$/, "", $1); gsub(/^[ \t]+|[ \t]+$/, "", $2); print $1 "=" $2}' $CONFIG_FILE_PATH)
    echo "$config"
}

# read key from the main conf file
get_license_key() {
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [ -z "$license_key" ]; then
        # Check if --json flag is present
        if [[ " $@ " =~ " --json " ]]; then
          license_key="No License Key"
        else
          license_key="${RED}No License Key${RESET}"
        fi
    else
        # Check if --json flag is present
        if [[ " $@ " =~ " --json " ]]; then
          echo "$license_key"
        else
          echo -e "${GREEN}$license_key${RESET}"
        fi
    fi
    
}



# dummy verification for terminal only to check on WHMCS directly
#
# OpenAdmin uses a different method
#
get_license_key_and_verify_on_my_openpanel() {
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [ -z "$license_key" ]; then
        echo -e "${RED}No License Key. Please add the key first: opencli config update key XXXXXXXXXX${RESET}"
        exit 0
    else
        ip_address=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)  # Get the public IP address
        check_token=$(openssl rand -hex 16)  # Generate a random token
        
        response=$(curl -sS -X POST -d "licensekey=$license_key&ip=$ip_address&check_token=$check_token" $WHMCS_URL)
        license_status=$(echo "$response" | grep -oP '(?<=<status>).*?(?=</status>)')
        
        if [ "$license_status" = "Active" ]; then

            # Check if --json flag is present
            if [[ " $@ " =~ " --json " ]]; then
                  echo "License is valid"
            else
                echo -e "${GREEN}License is valid${RESET}"
            fi

            # Check if --no-restart flag is present
            if [[ "$NO_RESTART" == "yes" ]]; then
                  echo "Please restart OpenAdmin service to enable new features."
		  exit 0
            else
                service admin restart > /dev/null
		echo "OpenPanel and OpenAdmin are restarted to apply Enterprise features."
            fi

        else
            # Check if --json flag is present
            if [[ " $@ " =~ " --json " ]]; then
                  echo "License is invalid"
            else
                echo -e "${RED}License is invalid${RESET}"
            fi
            exit 0
        fi
    fi
}





save_license_to_file() {
        new_key=$1
        if opencli config update key "$new_key" > /dev/null; then
            # Check if --json flag is present
            if [[ " $@ " =~ " --json " ]]; then
                echo "License key ${new_key} added."
            else
                echo -e "License key ${GREEN}${new_key}${RESET} added."
            fi

	    enable_emails_module  > /dev/null
	    docker --context default ps -q -f name=openpanel | grep . && docker --context default restart openpanel > /dev/null &

            # Check if --no-restart flag is present
            if [[ "$NO_RESTART" == "yes" ]]; then
                  echo "Please restart OpenAdmin service to enable new features."
		  exit 0
            else
	            service admin restart > /dev/null
	     	    echo "OpenPanel and OpenAdmin are restarted to apply Enterprise features."

            fi
        else
            # Check if --json flag is present
            if [[ " $@ " =~ " --json " ]]; then
                echo "License is valid, but failed to save the license key ${new_key}"
            else
                echo -e "${RED}License is valid, but failed to save the license key.${RESET}"
            fi
        fi
}



verify_license_first() {
    license_key=$1

        ip_address=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)
        check_token=$(openssl rand -hex 16)
        response=$(curl -sS -X POST -d "licensekey=$license_key&ip=$ip_address&check_token=$check_token" $WHMCS_URL)
        license_status=$(echo "$response" | grep -oP '(?<=<status>).*?(?=</status>)')
        
        if [ "$license_status" = "Active" ]; then
            save_license_to_file $new_key
        else

       if [[ "$JSON" == "yes" ]]; then
          echo "License is invalid"
        else
          echo -e "${RED}License is invalid${RESET}"
        fi
        exit 0
        fi
}



get_license_key_and_verify_on_my_openpanel_then_show_info() {
    config=$(read_config)
    license_key=$(echo "$config" | grep -i 'key' | cut -d'=' -f2)

    if [ -z "$license_key" ]; then

       if [[ "$JSON" == "yes" ]]; then
          echo "No License Key"
        else
          echo -e "${RED}No License Key. Please add the key first: opencli config update key XXXXXXXXXX${RESET}"
        fi
        exit 0
    else
        ip_address=$(curl --silent --max-time 2 -4 $IP_SERVER_1 || wget --timeout=2 -qO- $IP_SERVER_2 || curl --silent --max-time 2 -4 $IP_SERVER_3)  # Get the public IP address
        check_token=$(openssl rand -hex 16)  # Generate a random token
        
        response=$(curl -sS -X POST -d "licensekey=$license_key&ip=$ip_address&check_token=$check_token" $WHMCS_URL)
        license_status=$(echo "$response" | grep -oP '(?<=<status>).*?(?=</status>)')
        
        if [ "$license_status" = "Active" ]; then
            registered_name=$(echo "$response" | grep -oP '(?<=<registeredname>).*?(?=</registeredname>)')
            company_name=$(echo "$response" | grep -oP '(?<=<companyname>).*?(?=</companyname>)')
            email=$(echo "$response" | grep -oP '(?<=<email>).*?(?=</email>)')
            product_name=$(echo "$response" | grep -oP '(?<=<productname>).*?(?=</productname>)')
            reg_date=$(echo "$response" | grep -oP '(?<=<regdate>).*?(?=</regdate>)')
            next_due_date=$(echo "$response" | grep -oP '(?<=<nextduedate>).*?(?=</nextduedate>)')
            billing_cycle=$(echo "$response" | grep -oP '(?<=<billingcycle>).*?(?=</billingcycle>)')
            valid_ip=$(echo "$response" | grep -oP '(?<=<validip>).*?(?=</validip>)')

            # Check if --json flag is present
            if [[ "$JSON" == "yes" ]]; then          
                echo '{"Owner": "'"$registered_name"'","Company Name": "'"$company_name"'","Email": "'"$email"'","License Type": "'"$product_name"'","Registration Date": "'"$reg_date"'","Next Due Date": "'"$next_due_date"'","Billing Cycle": "'"$billing_cycle"'","Valid IP": "'"$valid_ip"'"}'
            else
                echo "Owner: $registered_name"
                echo "Company Name: $company_name"
                echo "Email: $email"
                echo "License Type: $product_name"
                echo "Registration Date: $reg_date"
                echo "Next Due Date: $next_due_date"
                echo "Billing Cycle: $billing_cycle"
                echo "Valid IP: $valid_ip"
            fi


        else


            # Check if --json flag is present
            if [[ "$JSON" == "yes" ]]; then
              echo "License is invalid"
            else
              echo -e "${RED}License is invalid${RESET}"
            fi
            exit 0
        fi
    fi
}






enable_emails_module() {
	    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2)
    	if echo "$enabled_modules" | grep -q 'emails'; then
	        :
	    else
	        new_modules="${enabled_modules},emails"
	        sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$CONFIG_FILE_PATH"
		fi  
}


disable_emails_module() {
	    enabled_modules=$(grep '^enabled_modules=' "$CONFIG_FILE_PATH" | cut -d'=' -f2)
        if echo "$enabled_modules" | grep -q 'emails'; then           
            new_modules=$(echo "$enabled_modules" | sed 's/,emails//g; s/emails,//g; s/^emails$//g')
            sed -i "s/^enabled_modules=.*/enabled_modules=${new_modules}/" "$CONFIG_FILE_PATH"
        else
            :
        fi
}


case "$1" in
    "key")
        # display the key
        license_key=$(get_license_key "$@")
        echo $license_key
        ;;
    "info")
        # display license info from whmcs
        get_license_key_and_verify_on_my_openpanel_then_show_info "$@" 
        ;;
    "verify")
        # check license on whmcs
        get_license_key_and_verify_on_my_openpanel "$@"
        ;;
    "delete")
        # remove the key and reload admin
        opencli config update key "" > /dev/null                        # delete key
        disable_emails_module                                           # remove email features from user panel
        rm -rf /etc/openpanel/openpanel/core/users/*/data.json          # purge cached users data to remove pro features
        service admin restart                                           # restart admin to remove pro features
        ;;
    "enterprise"*)
        # Update the license key "enterprise-"
        new_key=$1
        verify_license_first $new_key                                   # verify on whmcs, then save to file if valid
        exit 0        
        ;;
    *)
        echo -e "${RED}Invalid command.${RESET}"
        usage
        exit 1
        ;;
esac

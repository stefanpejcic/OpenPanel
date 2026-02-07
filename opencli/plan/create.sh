#!/bin/bash
################################################################################
# Script Name: plan/create.sh
# Description: Create a new hosting plan (Package) and set its limits.
# Usage: opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT>
# Example: opencli plan-create name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G
# Author: Radovan Jecmenica
# Created: 06.11.2023
# Last Modified: 06.02.2026
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
# Usage
help() {
    echo "Usage: opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT>"
    echo
    echo "Arguments:"
    echo "  name            - Name of the plan (string, no spaces)."
    echo "  description     - Plan description (string, use quotes for multiple words)."
    echo "  feature_set     - Name of the feature set used for users on this plan"
    echo "  email_limit     - Max number of email accounts (integer, 0 for unlimited)."
    echo "  max_email_quota - Max size per email account (intiger followed by B|k|M|G|T, 0 for unlimited)"    
    echo "  ftp_limit       - Max number of FTP accounts (integer, 0 for unlimited)."
    echo "  domains_limit   - Max number of domains (integer, 0 for unlimited)."
    echo "  websites_limit  - Max number of websites (integer, 0 for unlimited)."
    echo "  disk_limit      - Disk space limit in GB (integer)."
    echo "  inodes_limit    - Max number of inodes (integer, 0 = unlimited)."
    echo "  db_limit        - Max number of databases (integer, 0 for unlimited)."
    echo "  cpu             - CPU core limit (integer)."
    echo "  ram             - RAM limit in GB (integer)."
    echo "  bandwidth       - Port speed in Mbit/s (integer)."
    echo
    echo "Example:"
    echo "  opencli plan-create name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G"
    echo
    exit 1
}


# ======================================================================
# Validations

# TODO: update to 13 after 1.8.X
if [ "$#" -lt 12 ]; then
    help
fi

check_cpu_cores() {
  local available_cores=$(nproc) 
  #TODO: for slave servers, remove the limit
  if [ "$cpu" -gt "$available_cores" ]; then
    echo "Error: Insufficient CPU cores. Required: ${cpu}, Available: ${available_cores}"
    exit 1
  fi
}

check_available_ram() {
  #TODO: for slave servers, remove the limit
  local available_ram=$(free -m | awk '/^Mem:/ {printf "%d\n", ($2+512)/1024 }')

  if [ "$ram" -gt "$available_ram" ]; then
    echo "Error: Insufficient RAM. Required: ${ram}GB, Available: ${available_ram}GB"
    exit 1
  fi
}

check_plan_exists() {
    local name="$1"
    local sql="SELECT id FROM plans WHERE name='${name}';"
    local existing_plan=$(mysql --defaults-extra-file=$config_file -D "$mysql_database" -N -B -e "$sql")
    if [ -n "$existing_plan" ]; then
      echo "Plan name '${name}' already exists. Please choose another name."
      exit 1
    fi
}

validate_fields_first() {
    local ftp_limit="$1"
    local emails_limit="$2"
    local domains_limit="$3"
    local websites_limit="$4"
    local disk_limit="$5"
    local inodes_limit="$6"
    local db_limit="${7}"
    local cpu="$8"
    local ram="$9"
    local bandwidth="${10}"
    max_email_quota="${11}"

    is_integer() {
        [[ "$1" =~ ^-?[0-9]+$ ]]
    }

    is_number() {
        [[ "$1" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]
    }
    
    for var_name in ftp_limit emails_limit domains_limit websites_limit disk_limit inodes_limit db_limit bandwidth; do
        value="${!var_name}"
        if ! is_integer "$value"; then
            echo "Error: $var_name must be a number (integer only)"
            exit 1
        fi
    done

    for var_name in max_email_quota; do
        value="${!var_name}"

        if [[ "$value" =~ ^([0-9]+([.][0-9]+)?)([BkMGT]?)$ ]]; then
            number="${BASH_REMATCH[1]}"
            unit="${BASH_REMATCH[3]}"
    
            if [[ -z "$unit" && "$number" != "0" ]]; then
                value="${number}G"
            fi
    
            max_email_quota="$value"
        else
            echo "Error: $max_email_quota must be a number with optional unit (B|k|M|G|T)"
            exit 1
        fi
    done

    for var_name in cpu ram; do
        value="${!var_name}"
        if ! is_number "$value"; then
            echo "Error: $var_name must be a number (integer or float)"
            exit 1
        fi
    done
}


# ======================================================================
# Functions

source /usr/local/opencli/db.sh

insert_plan() {
  local name="$1"
  local description="$2"
  local email_limit="$3"
  local ftp_limit="$4"
  local domains_limit="$5"
  local websites_limit="$6"
  local disk_limit="$7"
  local inodes_limit="$8"
  local db_limit="${9}"
  local cpu="${10}"
  local ram="${11}"
  local bandwidth="${12}"
  local feature_set="${13}"
  local max_email_quota="${14}"

  disk_limit="${disk_limit} GB"

  if [ "$inodes_limit" -lt 0 ]; then
    inodes_limit=0
  fi

  ram="${ram}g"

  # Insert the plan into the 'plans' table
  local sql="INSERT INTO plans (name, description, email_limit, ftp_limit, domains_limit, websites_limit, disk_limit, inodes_limit, db_limit, cpu, ram, bandwidth, feature_set, max_email_quota) VALUES ('$name', '$description', $email_limit, $ftp_limit, $domains_limit, $websites_limit, '$disk_limit', $inodes_limit, $db_limit, $cpu, '$ram', $bandwidth, '$feature_set', '$max_email_quota');"

  mysql --defaults-extra-file=$config_file -D "$mysql_database" -e "$sql"
  if [ $? -eq 0 ]; then
        # if reseller, grant them access to the new plan id
        if [[ "$CHECK_PLAN_ID" == true ]]; then
            plan_id=$(check_plan_exists "$name")
            if [ -n "$plan_id" ]; then
                tmpfile=$(mktemp)

                jq --arg plan "$plan_id" \
                   '.allowed_plans |= (if index($plan) then . else . + [$plan] end)' \
                   "$reseller_file" > "$tmpfile" && mv "$tmpfile" "$reseller_file"
            else
                echo "Warning: failed to retrieve plan ID from the database and assign it to the reseller user."
            fi
        fi

    echo "Plan ${name} created successfully."
  else
    echo "Failed to create plan: ${name}"
  fi
}

check_reseller_user() {
    reseller_file="/etc/openpanel/openadmin/resellers/${reseller_user}.json"

    if [[ -z "$reseller_user" ]]; then
        return 1
    fi

    if [[ ! -f "$reseller_file" ]]; then
        echo "Error: Configuration file does not exist: $reseller_file"
        exit 1 #return 1
    fi

    if jq empty "$reseller_file" >/dev/null 2>&1; then
        CHECK_PLAN_ID=true
    else
        echo "Error: Invalid JSON in $reseller_file"
        exit 1 #return 1
    fi
}




# ======================================================================
# Parse flags
name=""
description=""
email_limit=""
ftp_limit=""
domains_limit=""
websites_limit=""
disk_limit=""
inodes_limit=""
db_limit=""
cpu=""
ram=""
bandwidth=""
feature_set="default"
max_email_quota="0" # TODO: temporary, for 1.7.42 before making it mandatory in 1.8.X

for arg in "$@"; do
  case $arg in
    reseller=*)
      reseller_user="${arg#*=}"
      ;;      
    name=*)
      name="${arg#*=}"
      ;;
    description=*)
      description="${arg#*=}"
      ;;
    emails=*)
      email_limit="${arg#*=}"
      ;;
    ftp=*)
      ftp_limit="${arg#*=}"
      ;;
    domains=*)
      domains_limit="${arg#*=}"
      ;;
    websites=*)
      websites_limit="${arg#*=}"
      ;;
    disk=*)
      disk_limit="${arg#*=}"
      ;;
    inodes=*)
      inodes_limit="${arg#*=}"
      ;;
    databases=*)
      db_limit="${arg#*=}"
      ;;
    cpu=*)
      cpu="${arg#*=}"
      ;;
    ram=*)
      ram="${arg#*=}"
      ;;
    bandwidth=*)
      bandwidth="${arg#*=}"
      ;;
    feature_set=*)
      feature_set="${arg#*=}"
      ;;
    max_email_quota=*)
      max_email_quota="${arg#*=}"
      ;;      
    *)
      echo "Unknown argument: $arg"
      help
      ;;
  esac
done




# ======================================================================
# Main
check_cpu_cores "$cpu"
check_available_ram "$ram"
check_plan_exists "${name}"
validate_fields_first "$ftp_limit" "$email_limit" "$domains_limit" "$websites_limit" "$disk_limit" "$inodes_limit" "$db_limit" "$cpu" "$ram" "$bandwidth" "$max_email_quota"
check_reseller_user # if no file or invalid json, abort
insert_plan "$name" "$description" "$email_limit" "$ftp_limit" "$domains_limit" "$websites_limit" "$disk_limit" "$inodes_limit" "$db_limit" "$cpu" "$ram" "$bandwidth" "$feature_set" "$max_email_quota"

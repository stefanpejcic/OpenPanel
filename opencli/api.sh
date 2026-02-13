#!/bin/bash
################################################################################
# Script Name: api/list.sh
# Description: On/Off OpenAdmin API access and list available API endpoints.
# Usage: opencli api <status|on|off|list>
# Author: Stefan Pejcic
# Created: 04.09.2024
# Last Modified: 12.02.2026
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
# Helpers

show_usage() {
    cat <<EOF

Usage: opencli api {status|on|off|list}
Docs:  https://dev.openpanel.com/cli/api.html

Examples:
  opencli api status              Display current OpenAdmin API status.
  opencli api on                  Enable OpenAdmin API access.
  opencli api off                 Disable OpenAdmin API access.
  opencli api list                List available API endpoints with examples.

EOF
}

# ======================================================================
# Validators

if [[ -z "$1" ]]; then
  show_usage
  exit 1
fi

license_key=$(grep "^key=" "/etc/openpanel/openpanel/conf/openpanel.config" | cut -d'=' -f2-)
[ -n "$license_key" ] || {
    echo "Error: OpenPanel Community edition does not support API access. Please consider purchasing the Enterprise version that has remote API access and integrations with billing softwares such as WHMCS and FOSSBilling."
    source "/usr/local/opencli/enterprise.sh"
    echo "$ENTERPRISE_LINK"
    exit 1
}


# ======================================================================
# Main
case "$1" in
  status)
    opencli config get api
    ;;
  on|off)
    opencli config update api "$1"
    if systemctl is-active --quiet admin; then
        systemctl restart admin 2>&1
    fi
    ;;
  list)
    if [ "$2" = "--save" ]; then
      # can only be run BEFORE pyarmor encoding
      python3 /usr/local/admin/modules/api/generate.py --save
    else
      python3 /usr/local/admin/modules/api/generate.py
    fi
    ;;
  *)
    echo "Invalid option: $1"
    show_usage
    exit 1
    ;;
esac

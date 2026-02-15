#!/bin/bash
################################################################################
# Script Name: server/logrotate.sh
# Description: Configures logrotate for caddy, openpanel, syslog.
# Usage: opencli server-logrotate
# Author: Stefan Pejcic
# Created: 16.01.2024
# Last Modified: 14.02.2026
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

CONFIG_FILE="/etc/openpanel/openpanel/conf/openpanel.config"

get_config() {
    grep -E "^$1=" "$CONFIG_FILE" 2>/dev/null | cut -d= -f2 | tr -d ' ' || echo "$2"
}

install_logrotate() {
    command -v logrotate &>/dev/null && return

    echo "logrotate is not installed. Installing..."

    if command -v apt-get &>/dev/null; then
        sudo apt-get update -qq && sudo apt-get install -y -qq logrotate
    elif command -v yum &>/dev/null; then
        sudo yum install -y -q logrotate
    elif command -v dnf &>/dev/null; then
        sudo dnf install -y -q logrotate
    else
        echo "No supported package manager found."
        exit 1
    fi

    command -v logrotate &>/dev/null || {
        echo "logrotate installation failed."
        exit 1
    }
}

install_logrotate

logrotate_enable=$(get_config logrotate_enable yes)
logrotate_size_limit=$(get_config logrotate_size_limit 100m)
logrotate_retention=$(get_config logrotate_retention 10)
logrotate_keep_days=$(get_config logrotate_keep_days 30)

[ "$logrotate_enable" != "yes" ] && {
    echo "Log rotation is not enabled."
    exit 0
}

# ---------- CADDY ----------
cat > /etc/logrotate.d/caddy-logs <<EOF
/var/log/caddy/domlogs/coraza_waf/*.log
/var/log/caddy/domlogs/domlogs/*/access.log {
    size $logrotate_size_limit
    rotate $logrotate_retention
    daily
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    create 640 root adm
    postrotate
        docker exec caddy bash -c "caddy validate && caddy reload" >/dev/null 2>&1
    endscript
    maxage $logrotate_keep_days
}
EOF

logrotate -f /etc/logrotate.d/caddy-logs
echo "Caddy log rotation configured."

# ---------- SYSLOG ----------
cat > /etc/logrotate.d/syslog <<EOF
/var/log/syslog {
    su root adm
    size $logrotate_size_limit
    rotate $logrotate_retention
    weekly
    missingok
    notifempty
    compress
    delaycompress
    postrotate
        systemctl reload rsyslog >/dev/null 2>&1 || true
    endscript
    maxage $logrotate_keep_days
}
EOF

logrotate -f /etc/logrotate.d/syslog
echo "Syslog rotation configured."

# ---------- OPENPANEL ----------
cat > /etc/logrotate.d/openpanel <<EOF
/var/log/openpanel/**/*.log {
    su root adm
    size $logrotate_size_limit
    rotate $logrotate_retention
    weekly
    missingok
    notifempty
    compress
    delaycompress
    copytruncate
    create 640 root adm
    maxage $logrotate_keep_days
}
EOF

logrotate -f /etc/logrotate.d/openpanel

echo "OpenPanel log rotation configured."

#!/bin/bash

# lfd only reads log files like /var/log/auth.log and /var/log/secure, on journald-only systems (Ubuntu 26.04, Debian 12 and newer)
# they don't exist so it never blocked anything. Install rsyslog where the firewall is installed and restart lfd so it opens them.
if command -v csf >/dev/null 2>&1 && ! systemctl is-active --quiet rsyslog; then
    if ! command -v rsyslogd >/dev/null 2>&1; then
        echo "Installing rsyslog so the firewall (lfd) has log files to read..."
        if command -v apt-get >/dev/null 2>&1; then
            DEBIAN_FRONTEND=noninteractive apt-get install -y rsyslog >/dev/null 2>&1
        elif command -v dnf >/dev/null 2>&1; then
            dnf install -y rsyslog >/dev/null 2>&1
        elif command -v yum >/dev/null 2>&1; then
            yum install -y rsyslog >/dev/null 2>&1
        fi
    fi
    if command -v rsyslogd >/dev/null 2>&1 && systemctl enable --now rsyslog >/dev/null 2>&1; then
        sleep 2
        systemctl restart lfd >/dev/null 2>&1
        echo "rsyslog enabled and lfd restarted, it now blocks brute force logins"
    else
        echo "Could not install or start rsyslog, lfd has no log files to read. Install it by hand: apt-get install -y rsyslog && systemctl enable --now rsyslog && systemctl restart lfd"
    fi
fi

# the Modern theme moved to caesarmind, installed copies still link the old stefanpejcic repo in their readme.txt
modern_readme=/etc/openpanel/openpanel/templates/modern/readme.txt
if [ -f "$modern_readme" ] && grep -q 'github.com/stefanpejcic/modern-openpanel-theme' "$modern_readme"; then
    sed -i -e 's#^Theme URI:.*#Theme URI: https://github.com/caesarmind/modern-openpanel-theme#' \
           -e 's#^Author URI:.*#Author URI: https://github.com/caesarmind#' "$modern_readme"
    echo "Modern theme readme now points at github.com/caesarmind"
fi

# hidden_docker_apps lists the one-click Docker apps the admin took off the Auto Installer, existing configs don't have the key yet
op_config=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$op_config" ] && ! grep -q '^hidden_docker_apps=' "$op_config"; then
    if grep -q '^applications_dashboard_items=' "$op_config"; then
        sed -i '/^applications_dashboard_items=/a hidden_docker_apps=' "$op_config"
    elif grep -q '^\[PANEL\]' "$op_config"; then
        sed -i '/^\[PANEL\]/a hidden_docker_apps=' "$op_config"
    else
        printf '\n[PANEL]\nhidden_docker_apps=\n' >> "$op_config"
    fi
    echo "Added hidden_docker_apps= to $op_config"
fi

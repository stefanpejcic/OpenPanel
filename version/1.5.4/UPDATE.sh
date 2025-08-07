#!/bin/bash

# https://github.com/shinsenter/php/discussions/301
sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /home/*/docker-compose.yml

sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /etc/openpanel/docker/compose/1.0/docker-compose.yml

echo "🔍 Checking if Apache is installed..."

# Detect package manager
if command -v apt-get >/dev/null 2>&1; then
    PKG_MGR="apt"
elif command -v yum >/dev/null 2>&1; then
    PKG_MGR="yum"
elif command -v dnf >/dev/null 2>&1; then
    PKG_MGR="dnf"
else
    echo "❌ No supported package manager found (apt, yum, dnf)"
    exit 1
fi

# Define Apache-related packages
APT_PACKAGES="apache2 apache2-bin apache2-data apache2-utils libapache2-mod-php libapache2-mod-php8.3"
YUM_PACKAGES="httpd httpd-tools mod_php"

# Check if Apache is installed
if command -v apache2 >/dev/null 2>&1 || command -v httpd >/dev/null 2>&1; then
    echo "✅ Apache is installed. Proceeding to remove it..."
else
    echo "✅ Apache is NOT installed. Nothing to remove."
    exit 0
fi

# Remove Apache
if [ "$PKG_MGR" = "apt" ]; then
    apt-get purge -y $APT_PACKAGES
    apt-get autoremove -y
elif [ "$PKG_MGR" = "yum" ] || [ "$PKG_MGR" = "dnf" ]; then
    $PKG_MGR remove -y $YUM_PACKAGES
    $PKG_MGR autoremove -y || true
fi

# Final check
echo "🔁 Verifying removal..."
if command -v apache2 >/dev/null 2>&1 || command -v httpd >/dev/null 2>&1; then
    echo "❌ Apache is STILL installed!"
    exit 1
else
    echo "✅ Apache has been successfully removed."
    docker --context=default restart caddy
    exit 0
fi

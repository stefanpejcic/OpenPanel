#!/bin/bash

# https://github.com/stefanpejcic/OpenPanel/issues/671#issuecomment-3242952264

echo "Adding support for custom SSL on OpenPanel GUI.."
wget -4 -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/fc1b6503cc62e3d61abab6473e42175e63fb7436/openpanel/service/service.config.py && docker restart openpanel

echo "Adding support for custom SSL on OpenAdmin GUI.."
wget -4 -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py


echo "Fixing 403 error on phpmyadmin autologin for users created before 1.4.5"
wget -4 -O /etc/openpanel/mysql/phpmyadmin/pma.php https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php

for path in /home/*/pma.php; do
    if [ -d "$path" ]; then
        rm -rf "$path"
        cp /etc/openpanel/mysql/phpmyadmin/pma.php "$path"
        echo "- Fix applied to: $path"
    fi
done

echo "Fixing syntax of Nginx VHosts template causing 400 error when Elementor is used on WP with pretty-permalinks.."
sed -i 's/index\.php\$is_args\?\$args/index.php?\$args/g' /etc/openpanel/nginx/vhosts/1.1/docker_nginx_domain.conf
sed -i 's/index\.php\$is_args\?\$args/index.php?\$args/g' /etc/openpanel/nginx/vhosts/1.1/docker_openresty_domain.conf

for userdir in /home/*/; do
    username=$(basename "$userdir")
    target_dir="$userdir/docker-data/volumes/${username}_webserver_data/_data"
    if [ -d "$target_dir" ]; then
        echo "Processing $target_dir..."
        for conf_file in "$target_dir"/*.conf; do
            [ -f "$conf_file" ] || continue
            echo "  Checking VHosts file: $conf_file"
            sed -i 's/index\.php\$is_args\?\$args/index.php?\$args/g' "$conf_file"
        done
    fi
done

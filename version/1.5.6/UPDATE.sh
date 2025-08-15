#!/bin/bash


# https://github.com/stefanpejcic/OpenPanel/issues/640
CONFIG_FILE="/etc/openpanel/mysql/container_my.cnf"
LINE="skip-ssl"
grep -qxF "$LINE" "$CONFIG_FILE" || echo "$LINE" >> "$CONFIG_FILE"

# https://my.openpanel.com/adminad/supporttickets.php?action=view&id=3937
mysql --database=panel -e "
ALTER TABLE domains CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE sites CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
"


# datatracker.ietf.org/doc/html/rfc1035#autoid-31
for file in /etc/openpanel/bind9/zone_template.txt /etc/openpanel/bind9/zone_template_ipv6.txt; do
    sed -i 's/\(@[[:space:]]\+IN[[:space:]]\+SOA[[:space:]]\+{ns1}\.\)[[:space:]]\+{ns2}\./\1 {rpemail}./' "$file"
done



# https://github.com/shinsenter/php/discussions/301
echo ""
echo "Applying fix for https://github.com/shinsenter/php/discussions/301"

for file in /home/*/docker-compose.yml; do
  cp "$file" "$file.154_UPDATE.bak"
  echo "Editing $file"

  sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' "$file"
done
echo ""

echo "Editing tempalte used for new users.."
sed -i '/command: >/{N;s/command: >\n *sh -c "php-fpm --allow-to-run-as-root"/command: php-fpm --allow-to-run-as-root/}' /etc/openpanel/docker/compose/1.0/docker-compose.yml



# LITESPEED

mkdir -p /etc/openpanel/openlitespeed
wget -O /etc/openpanel/nginx/vhosts/1.1/docker_openlitespeed_domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/docker_openlitespeed_domain.conf
wget -O /etc/openpanel/openlitespeed/httpd_config.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openlitespeed/httpd_config.conf
wget -O /etc/openpanel/openlitespeed/start.sh https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openlitespeed/start.sh
chmod +x /etc/openpanel/openlitespeed/start.sh


# for sse
wget -O /etc/openpanel/openadmin/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/service/service.config.py





#!/bin/bash

echo ""
echo "📥 Updating HTML templates for domains.."

wget -O /etc/openpanel/nginx/default_page.html https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/default_page.html
wget -O /etc/openpanel/nginx/suspended_user.html https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/suspended_user.html
wget -O /etc/openpanel/nginx/suspended_website.html https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/suspended_website.html


echo ""
echo "📥 Updating VirtualHost templates for domains.."

wget -O /etc/openpanel/nginx/vhosts/1.1/docker_nginx_domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/docker_nginx_domain.conf
wget -O /etc/openpanel/nginx/vhosts/1.1/docker_openresty_domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/docker_openresty_domain.conf
wget -O /etc/openpanel/nginx/vhosts/1.1/docker_apache_domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/nginx/vhosts/1.1/docker_apache_domain.conf


echo ""
echo "📥 Updating docker-compose.yml file for new domains.."

cp /etc/openpanel/docker/compose/1.0/docker-compose.yml /etc/openpanel/docker/compose/1.0/127-docker-compose.yml
wget -O /etc/openpanel/docker/compose/1.0/docker-compose.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/docker-compose.yml

#!/bin/bash

echo "Donwloading template for OpenResty"
wget -O /etc/openpanel/nginx/vhosts/1.1/docker_openresty_domain.conf https://github.com/stefanpejcic/openpanel-configuration/blob/main/nginx/vhosts/1.1/docker_openresty_domain.conf
wget -O /etc/openpanel/openresty/nginx.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openresty/nginx.conf


echo "Updating template: /etc/openpanel/varnish/default.vcl"
wget -O /etc/openpanel/varnish/default.vcl https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl




for dir in /home/*; do
    file="$dir/.env"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        cp /etc/openpanel/varnish/default.vcl file="$dir/default.vcl"
        echo "- Updated Varnish default.vcl template for user: $user"
        
        if ! grep -q 'CRONJOBS' "$file"; then
            sed -i '/BUSYBOX_RAM="0.1G"/a \
# CRONJOBS\nCRON_CPU="0.1"\nCRON_RAM="0.25G"' "$file"

            echo "- Updated $file for user: $user to add CRON limits"
        fi       

        cp /etc/openpanel/openresty/nginx.conf file="$dir/openresty.conf"
        echo "- Created openresty settings  template for user: $user" 
        # TODO:
        # edit .env
        # and edit docker-compose.yml
    fi

    file="$dir/docker-compose.yml"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        echo "Fixing permission issues in PHP containers.. You should restart services manually to re-apply changes."
        sed -i 's/- APP_USER=${CONTEXT:-root}/- APP_USER=root/g' $file
        sed -i 's/- APP_GROUP=${CONTEXT:-root}/- APP_GROUP=root/g' $file
    fi   
done


echo "Purging old openpanel/openpanel-ui images.."
all_images=$(docker --context default images --format "{{.Repository}} {{.ID}}" | grep "^openpanel/openpanel-ui" | awk '{print $2}')
used_images=$(docker --context default ps --format "{{.Image}}" | xargs -n1 docker inspect --format '{{.Id}}' 2>/dev/null | sort | uniq)
for img in $all_images; do
    if echo "$used_images" | grep -q "$img"; then
        echo "⏩ Skipping in-use image: $img"
    else
        echo "🗑️ Deleting unused image: $img"
        docker rmi "$img"
    fi
done

: '

1. check if enterprise license and mailserver running

2. install 
apt update && apt install msmtp msmtp-mta -y
chmod a+x /usr/bin/msmtp


3. for all existing users create `/etc/msmtprc`

```
# /etc/msmtprc
defaults
auth       off
tls        off
logfile    /var/log/msmtp.log

account    default
host       PUBLIC_IPV4_HERE
port       25
from       OPENPANEL_USERNAME@SERVER_HOSTNAME
```

4. update compsoe files:
Add 
      - /usr/bin/msmtp:/usr/bin/msmtp:ro
      - ./msmtprc:/etc/msmtprc:ro
for each php service!

5. edit ini files:

replace
```
; sendmail_path = 
```
with

```
sendmail_path = /usr/bin/msmtp -t
```

6. reload those php services!

'

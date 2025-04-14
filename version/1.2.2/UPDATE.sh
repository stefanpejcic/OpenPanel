#!/bin/bash

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
    fi

    file="$dir/docker-compose.yml"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        echo "Fixing permission issues in PHP containers.. You should restart services manually to re-apply changes."
        sed -i 's/- APP_USER=${CONTEXT:-root}/- APP_USER=root/g' $file
        sed -i 's/- APP_GROUP=${CONTEXT:-root}/- APP_GROUP=root/g' $file
    fi   
done


: '
apt update && apt install msmtp msmtp-mta -y
chmod a+x /usr/bin/msmtp

Add 
      - /usr/bin/msmtp:/usr/bin/msmtp:ro
for each php service!



create `/etc/msmtprc` for each user

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

edit ini files:

replace
```
; sendmail_path = 
```
with

```
sendmail_path = /usr/bin/msmtp -t
```

reload those php services!
'

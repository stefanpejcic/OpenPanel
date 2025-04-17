#!/bin/bash

echo "Downloading template for OpenResty"

wget -O /etc/openpanel/nginx/vhosts/1.1/docker_openresty_domain.conf https://github.com/stefanpejcic/openpanel-configuration/blob/main/nginx/vhosts/1.1/docker_openresty_domain.conf
mkdir -p /etc/openpanel/openresty/
wget -O /etc/openpanel/openresty/nginx.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openresty/nginx.conf


echo "Updating template: /etc/openpanel/varnish/default.vcl"
wget -O /etc/openpanel/varnish/default.vcl https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl


echo "Extending list of forbidden usernames: /etc/openpanel/openadmin/config/forbidden_usernames.txt "
wget -O /etc/openpanel/openadmin/config/forbidden_usernames.txt https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/forbidden_usernames.txt

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

        cp /etc/openpanel/openresty/nginx.conf $dir/openresty.conf
        echo "- Created openresty settings  template for user: $user" 

                # Check if the file already contains OPENRESTY config
                    if grep -q '^VARNISH_VERSION=' "$file" && \
                       grep -q '^VARNISH_CPU=' "$file" && \
                       grep -q '^VARNISH_RAM=' "$file" && \
                       ! grep -q '^OPENRESTY_VERSION=' "$file" && \
                       ! grep -q '^OPENRESTY_CPU=' "$file" && \
                       ! grep -q '^OPENRESTY_RAM=' "$file"; then
                    
                        awk '
                            BEGIN { inserted=0 }
                            {
                                print
                                if ($0 ~ /^VARNISH_RAM=/ && inserted == 0) {
                                    print "# OPENRESTY"
                                    print "OPENRESTY_VERSION=\"bullseye-fat\""
                                    print "OPENRESTY_CPU=\"0.5\""
                                    print "OPENRESTY_RAM=\"0.5G\""
                                    inserted=1
                                }
                            }
                        ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
                    
                        echo "- Inserted OPENRESTY config after VARNISH_RAM in $file for user: $user"
                    else
                        echo "- Skipped OPENRESTY config for $user (missing VARNISH or already has OPENRESTY)"
                    fi

    fi

    file="$dir/docker-compose.yml"
    user=$(basename "$dir")

    if [[ -f "$file" ]]; then
        cp $file $dir/122_docker-compose.yml
        echo "Fixing permission issues in PHP containers.. You should restart services manually to re-apply changes."
        sed -i 's/- APP_USER=${CONTEXT:-root}/- APP_USER=root/g' $file
        sed -i 's/- APP_GROUP=${CONTEXT:-root}/- APP_GROUP=root/g' $file


        # Check if 'openresty:' already exists
        if grep -q "^  openresty:" "$file"; then
            echo "✅ 'openresty' service already exists in $file."
            exit 0
        fi

# Define the openresty service block
read -r -d '' OPENRESTY_BLOCK <<'EOF'
  openresty:
    image: openresty/openresty:${OPENRESTY_VERSION:-bullseye-fat}
    container_name: openresty
    restart: always
    ports:
      - "${PROXY_HTTP_PORT:-${HTTP_PORT}}"
      - "${HTTPS_PORT}"
    working_dir: /var/www/html
    volumes:
      # - ./openresty.conf:/etc/openresty/nginx.conf:ro # TODO
      - webserver_data:/etc/nginx/conf.d
      - html_data:/var/www/html/                          # Website files
      - /etc/openpanel/nginx/certs/:/etc/nginx/ssl/       # SSLs
    deploy:
      resources:
        limits:
          cpus: "${OPENRESTY_CPU:-0.5}"
          memory: "${OPENRESTY_RAM:-0.5G}"
    networks:
      - www

EOF

INDENTED_BLOCK=$(echo "$OPENRESTY_BLOCK" | sed 's/^/  /')
NGINX_LINE=$(grep -n "^  nginx:" "$file" | cut -d: -f1)

if [ -z "$NGINX_LINE" ]; then
    echo "❌ 'nginx:' service not found. Cannot determine insertion point."
    exit 1
fi

# Insert the openresty block above the nginx line
awk -v insert_line="$NGINX_LINE" -v new_block="$INDENTED_BLOCK" '
NR == insert_line { print new_block }
{ print }
' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"

        echo "✅ 'openresty' service added in file $file"
       
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

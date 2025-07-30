#!/bin/bash

#wget -O /etc/openpanel/ftp/start_vsftpd.sh https://raw.githubusercontent.com/stefanpejcic/OpenPanel-FTP/refs/heads/master/start_vsftpd.sh


touch /root/openpanel_restart_needed
wget -O /etc/openpanel/openpanel/service/service.config.py https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/service/service.config.py




set -e

file="/root/docker-compose.yml"
backup="${file}.152_bak"

# Backup the original file
cp "$file" "$backup"

# New FTP block
read -r -d '' ftp_block << 'EOF'
  # FTP
  openadmin_ftp:
    build:
      context: /etc/openpanel/ftp/    
    container_name: openadmin_ftp
    restart: always
    ports:
      - "21:21"
      - "${FTP_PORT_RANGE}:21000-21010"
    volumes:
      - /home/:/home/
      - /etc/openpanel/ftp/vsftpd.conf:/etc/vsftpd/vsftpd.conf
      - /etc/openpanel/ftp/start_vsftpd.sh:/bin/start_vsftpd.sh
      - /etc/openpanel/ftp/vsftpd.chroot_list:/etc/vsftpd.chroot_list
      - /etc/openpanel/ftp/all.users:/etc/openpanel/ftp/all.users
      - /etc/openpanel/ftp/users/:/etc/openpanel/ftp/users/
      - /etc/openpanel/caddy/ssl/:/etc/openpanel/caddy/ssl/:ro #ssl
    networks:
      - openpanel_network
    deploy:
      resources:
        limits:
          cpus: "${FTP_CPUS:-0.5}"
          memory: "${FTP_RAM:-0.5G}"
          pids: 100
EOF

# New OpenPanel block
read -r -d '' openpanel_block << 'EOF'
  # OpenPanel service running on port 2083
  openpanel:
    image: openpanel/openpanel-ui:${VERSION}
    container_name: openpanel
    depends_on:
      - openpanel_mysql
      - openpanel_redis
      #- clamav
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    ports:
      - "${PORT}:2083/tcp"
    volumes:
      - /etc/bind:/etc/bind
      - /lib/modules:/lib/modules:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /hostfs/run/user/:/hostfs/run/user/
      - /etc/passwd:/hostfs/etc/passwd:ro
      - /etc/group:/hostfs/etc/group:ro
      - /home:/home
      - /usr/local/admin:/usr/local/admin
      - /usr/local/bin/opencli:/usr/local/bin/opencli
      - /usr/local/opencli:/usr/local/opencli
      - /var/log/caddy:/var/log/caddy
      - /var/log/openpanel:/var/log/openpanel
      - /var/lib/csf/:/var/lib/csf/
      - /etc/openpanel/mysql/container_my.cnf:/etc/my.cnf
      - /etc/openpanel/:/etc/openpanel/
      - /var/run/docker.sock:/var/run/docker.sock
      - mysql:/var/lib/mysql
      - /usr/bin/docker:/usr/bin/docker
      - /root/.ssh/:/root/.ssh/:ro
      - /root/.docker/:/root/.docker/
      - /root/openpanel_restart_needed:/root/openpanel_restart_needed #flag
      - /root/.env:/root/.env
      - /root/docker-compose.yml:/root/docker-compose.yml
      - ${REDIS_SOCKET}:/tmp/redis/
      - /etc/openpanel/openpanel/custom_code/:/templates/custom_code/
      - /etc/openpanel/openpanel/custom_code/custom.css:/static/css/custom.css
      - /etc/openpanel/openpanel/custom_code/custom.js:/static/js/custom.js
      - /etc/openpanel/openpanel/conf/knowledge_base_articles.json:/etc/openpanel/openpanel/conf/knowledge_base_articles.json
      - /etc/openpanel/openpanel/translations/:/etc/openpanel/openpanel/translations/
      - /usr/local/mail/openmail/:/usr/local/mail/openmail/
      - /etc/localtime:/etc/localtime:ro
      - /etc/openpanel/modules/:/modules/plugins/
    restart: always
    privileged: true
    networks:
      - openpanel_network
    deploy:
      resources:
        limits:
          cpus: "${OPENPANEL_CPUS:-1.0}"
          memory: "${OPENPANEL_RAM:-1.0G}"
          pids: 300
EOF

# Replace blocks
awk -v ftp_block="$ftp_block" -v openpanel_block="$openpanel_block" '
  BEGIN { replacing = 0 }
  /^  # FTP$/ {
    print ftp_block
    replacing = 1
    next
  }
  /^  # OpenPanel service running on port 2083$/ {
    print openpanel_block
    replacing = 2
    next
  }
  /^[^ ]/ {
    replacing = 0
  }
  replacing == 1 || replacing == 2 {
    next
  }
  {
    print
  }
' "$backup" > "$file"

# Restart containers
cd /root

echo "Stopping openpanel..."
docker --context=default compose down openpanel
echo "Starting openpanel..."
docker --context=default compose up openpanel -d

echo "Done!"

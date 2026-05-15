#!/bin/bash

OPENMAIL_DIR="/usr/local/mail/openmail"

wget -q -O "/etc/openpanel/caddy/check.conf" https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/check.conf

# create symlinks for user files
for env_file in /home/*/.env; do
    [ -e "$env_file" ] || continue
    USERNAME=$(basename "$(dirname "$env_file")")
    HOME_DIR="/home/$USERNAME"
    TARGET_LINK="$HOME_DIR/files"
    SOURCE_DIR="$HOME_DIR/docker-data/volumes/${USERNAME}_html_data/_data/"
    if [ ! -e "$TARGET_LINK" ]; then
        ln -sfn "$SOURCE_DIR" "$TARGET_LINK"
    fi
done

if [ -d "$OPENMAIL_DIR" ]; then

	# 1. edit openpanel contianer to NOT use :ro on the files
    sed -i 's|$OPENMAIL_DIR/:$OPENMAIL_DIR/:ro|$OPENMAIL_DIR/:$OPENMAIL_DIR/|g' "/root/docker-compose.yml"

    # 2. download files for mail quotas
    if [ ! -d "$OPENMAIL_DIR/openpanel" ]; then
    	mkdir -p $OPENMAIL_DIR/openpanel
    	wget -O $OPENMAIL_DIR/openpanel/utils.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/utils.sh
    	wget -O $OPENMAIL_DIR/openpanel/accounts.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/accounts.sh
    fi
	
	# 3. download roundcube plugins for webmail autologon
    if [ ! -d "$OPENMAIL_DIR/roundcube-plugins" ]; then
    	mkdir -p $OPENMAIL_DIR/roundcube-plugins
    	wget -q -O "$OPENMAIL_DIR/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php
        wget -q -O "$OPENMAIL_DIR/roundcube-plugins/autologin.py" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/autologin.py
    fi

    # 4.edit mailserver to overwrite our files!
	if [ -f "$OPENMAIL_DIR/compose.yml" ]; then
	    COMPOSE="$OPENMAIL_DIR/compose.yml"
	    
	    MISSING_ACCOUNTS=true
	    MISSING_UTILS=true
	    
	    grep -qF './openpanel/accounts.sh:/usr/local/bin/helpers/accounts.sh:ro' "$COMPOSE" && MISSING_ACCOUNTS=false
	    grep -qF './openpanel/utils.sh:/usr/local/bin/helpers/utils.sh:ro' "$COMPOSE" && MISSING_UTILS=false
	    
	    if $MISSING_ACCOUNTS || $MISSING_UTILS; then
	        MAILSERVER_LINE=$(grep -n '^\s\{2\}mailserver:' "$COMPOSE" | cut -d: -f1)
	        NEXT_SERVICE_LINE=$(awk "NR > $MAILSERVER_LINE && /^  [a-z]/ { print NR; exit }" "$COMPOSE")
	        
	        VOLUMES_LINE=$(awk "NR > $MAILSERVER_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*volumes:/ { print NR; exit }" "$COMPOSE")
	        
	        if [ -z "$VOLUMES_LINE" ]; then
	            echo "ERROR: Could not find volumes: section under mailserver service" >&2
	            exit 1
	        fi
	        
	        LAST_VOLUME_LINE=$(awk "NR > $VOLUMES_LINE && NR < ${NEXT_SERVICE_LINE:-9999} && /^\s*- / { last=NR } END { print last+0 }" "$COMPOSE")
	        
	        if [ "$LAST_VOLUME_LINE" -eq 0 ]; then
	            echo "ERROR: Could not find any volume entries under mailserver" >&2
	            exit 1
	        fi
	        
	        INSERT=""
	        $MISSING_ACCOUNTS && INSERT="${INSERT}      - ./openpanel/accounts.sh:/usr/local/bin/helpers/accounts.sh:ro\n"
	        $MISSING_UTILS    && INSERT="${INSERT}      - ./openpanel/utils.sh:/usr/local/bin/helpers/utils.sh:ro\n"
	        
	        sed -i "${LAST_VOLUME_LINE}a\\$(printf '%b' "$INSERT" | head -c -1)" "$COMPOSE"
	        
	        echo "Added volume mounts to mailserver in $COMPOSE"

	        echo "Restarting mailserver to use the overwrites for storage.."
	 		cd $OPENMAIL_DIR && docker --context=default compsoe down mailserver && docker --context=default compsoe up -d mailserver
	 		echo "done"
	    fi
	fi
fi







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

# Download roundcube plugins

if [ -d "$OPENMAIL_DIR" ]; then

	# 1. edit openpanel contianer to NOT use :ro on the files
    sed -i 's|$OPENMAIL_DIR/:$OPENMAIL_DIR/:ro|$OPENMAIL_DIR/:$OPENMAIL_DIR/|g' "/root/docker-compose.yml"

    # 2. download files
    if [ ! -d "$OPENMAIL_DIR/openpanel" ]; then
    	mkdir -p $OPENMAIL_DIR/openpanel
    	wget -O $OPENMAIL_DIR/openpanel/utils.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/utils.sh
    	wget -O $OPENMAIL_DIR/openpanel/accounts.sh https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/openpanel/accounts.sh
    fi

    if [ ! -d "$OPENMAIL_DIR/roundcube-plugins" ]; then
    	mkdir -p $OPENMAIL_DIR/roundcube-plugins
    	wget -q -O "$OPENMAIL_DIR/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/dovecot_impersonate/dovecot_impersonate.php
        wget -q -O "$OPENMAIL_DIR/roundcube-plugins/autologin.py" https://raw.githubusercontent.com/stefanpejcic/OpenMail/refs/heads/main/roundcube-plugins/autologin.py
    fi

    # 3.edit mailserver to overwrite our files!
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






# ── RADOVAN

COMPOSE_FILE="/usr/local/mail/openmail/compose.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
    echo "ERROR: $COMPOSE_FILE not found."
    exit 1
fi

cp "$COMPOSE_FILE" "${COMPOSE_FILE}.bak.$(date +%Y%m%d%H%M%S)"

python3 - "$COMPOSE_FILE" <<'PYEOF'
import sys

compose_path = sys.argv[1]

with open(compose_path, "r") as f:
    lines = f.readlines()

def find_service_range(lines, service_name):
    start = None
    end = None
    for i, line in enumerate(lines):
        if line.rstrip() == f"  {service_name}:":
            start = i
        elif start is not None:
            if len(line) > 2 and line[:2] == '  ' and line[2] != ' ' and line.rstrip().endswith(':'):
                end = i
                break
            elif line[0] != ' ' and line.strip():
                end = i
                break
    if start is not None and end is None:
        end = len(lines)
    return start, end

rc_start, rc_end = find_service_range(lines, "roundcube")

if rc_start is None:
    print("ERROR: 'roundcube' service not found in compose.yml")
    sys.exit(1)

changed = 0

VOL_IMPERSONATE = "      - ./roundcube-plugins/dovecot_impersonate:/var/www/html/plugins/dovecot_impersonate:ro\n"
VOL_AUTOLOGIN   = "      - ./roundcube-plugins/autologin.php:/var/www/html/public_html/autologin.php:ro\n"

rc_block = lines[rc_start:rc_end]
has_volumes_section = any(line.rstrip() == "    volumes:" for line in rc_block)
has_vol_impersonate = any("dovecot_impersonate:/var/www/html/plugins/dovecot_impersonate" in line for line in rc_block)
has_vol_autologin   = any("autologin.php:/var/www/html/public_html/autologin.php" in line for line in rc_block)

if not has_volumes_section:
    anchor_keywords = ["    ports:", "    networks:", "    command:"]
    insert_at = rc_end
    for i, line in enumerate(rc_block):
        if any(line.startswith(kw) for kw in anchor_keywords):
            insert_at = rc_start + i
            break

    new_lines = ["    volumes:\n"]
    if not has_vol_impersonate:
        new_lines.append(VOL_IMPERSONATE)
    if not has_vol_autologin:
        new_lines.append(VOL_AUTOLOGIN)

    lines = lines[:insert_at] + new_lines + lines[insert_at:]
    changed += 1

else:
    rc_block = lines[rc_start:rc_end]
    last_vol_idx = None
    in_vol = False
    for i, line in enumerate(rc_block):
        if line.rstrip() == "    volumes:":
            in_vol = True
            continue
        if in_vol:
            if line.startswith("      - "):
                last_vol_idx = rc_start + i
            else:
                break

    insert_at = (last_vol_idx + 1) if last_vol_idx is not None else rc_end
    to_add = []
    if not has_vol_impersonate:
        to_add.append(VOL_IMPERSONATE)
    if not has_vol_autologin:
        to_add.append(VOL_AUTOLOGIN)

    if to_add:
        lines = lines[:insert_at] + to_add + lines[insert_at:]
        changed += 1

for i, line in enumerate(lines):
    if "ROUNDCUBEMAIL_PLUGINS=" in line and "dovecot_impersonate" not in line:
        lines[i] = line.rstrip() + ",dovecot_impersonate\n"
        changed += 1
        break

if changed:
    with open(compose_path, "w") as f:
        f.writelines(lines)
    print("OK: compose.yml updated successfully.")
else:
    print("OK: compose.yml already up to date. No changes needed.")
PYEOF





#  Check and patch backup service networks in compose.yml ───────────────

python3 - "$COMPOSE_FILE" <<'PYEOF2'
import sys

compose_path = sys.argv[1]

with open(compose_path, "r") as f:
    lines = f.readlines()

def find_service_range(lines, service_name):
    start = None
    end = None
    for i, line in enumerate(lines):
        if line.rstrip() == f"  {service_name}:":
            start = i
        elif start is not None:
            if len(line) > 2 and line[:2] == '  ' and line[2] != ' ' and line.rstrip().endswith(':'):
                end = i
                break
            elif line[0] != ' ' and line.strip():
                end = i
                break
    if start is not None and end is None:
        end = len(lines)
    return start, end

backup_start, backup_end = find_service_range(lines, "backup")

if backup_start is None:
    print("ERROR: 'backup' service not found in compose.yml")
    sys.exit(1)

backup_block = lines[backup_start:backup_end]

has_networks_section = any(line.rstrip() == "    networks:" for line in backup_block)
has_www = any(line.strip() == "- www" for line in backup_block)

if has_www:
    print("OK: 'www' network already present in backup service. No changes needed.")
    sys.exit(0)

if has_networks_section:
    # Find the last network entry and append after it
    last_net_idx = None
    in_networks = False
    for i, line in enumerate(backup_block):
        if line.rstrip() == "    networks:":
            in_networks = True
            continue
        if in_networks:
            if line.startswith("      - "):
                last_net_idx = backup_start + i
            else:
                break

    insert_at = (last_net_idx + 1) if last_net_idx is not None else backup_end
    lines = lines[:insert_at] + ["      - www\n"] + lines[insert_at:]
else:
    # No networks section at all — add before restart or at end of block
    anchor_keywords = ["    restart:", "    deploy:"]
    insert_at = backup_end
    for i, line in enumerate(backup_block):
        if any(line.startswith(kw) for kw in anchor_keywords):
            insert_at = backup_start + i
            break
    lines = lines[:insert_at] + ["    networks:\n", "      - www\n"] + lines[insert_at:]

with open(compose_path, "w") as f:
    f.writelines(lines)
print("OK: 'www' network added to backup service.")
PYEOF2

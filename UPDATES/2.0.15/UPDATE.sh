#!/bin/bash

# cloudflare-only snippet now also allows the server's own ips, rebuild it so wp-cron and loopback requests stop getting 403
CF_SNIPPET="/etc/openpanel/caddy/templates/cloudflare.only"
if [ -f "$CF_SNIPPET" ] && ! grep -q "server ip" "$CF_SNIPPET"; then
    echo "Regenerating $CF_SNIPPET..."
    opencli domains-cloudflare
    if podman ps --format '{{.Names}}' | grep -qx caddy; then
        echo "Reloading Caddy..."
        podman exec caddy caddy reload --config /etc/caddy/Caddyfile >/dev/null 2>&1
    fi
fi

OPENPANEL_CONFIG="/etc/openpanel/openpanel/conf/openpanel.config"

# acccess_logs_per_page typo fixed in 2.0.15, keep whatever value was set
if [ -f "$OPENPANEL_CONFIG" ] && grep -q "^acccess_logs_per_page=" "$OPENPANEL_CONFIG"; then
    echo "Renaming acccess_logs_per_page to access_logs_per_page in $OPENPANEL_CONFIG..."
    sed -i 's/^acccess_logs_per_page=/access_logs_per_page=/' "$OPENPANEL_CONFIG"
fi

# servers installed with the template's default mail_security_token get their own
if [ -f "$OPENPANEL_CONFIG" ] && grep -q "^mail_security_token=78bd237db2da$" "$OPENPANEL_CONFIG"; then
    echo "Generating a new mail_security_token..."
    sed -i "s|^mail_security_token=.*|mail_security_token=$(openssl rand -hex 16)|" "$OPENPANEL_CONFIG"
fi

# the panel no longer creates mcp_tokens on the fly, new installs get it from the initial sql
source /usr/local/opencli/db.sh
echo "Making sure the mcp_tokens table exists..."
mariadb --defaults-extra-file="$config_file" -D "$mysql_database" -e "
CREATE TABLE IF NOT EXISTS mcp_tokens (
  id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id int NOT NULL,
  name varchar(100) NOT NULL,
  token_prefix varchar(20) NOT NULL,
  token_hash char(64) NOT NULL UNIQUE,
  read_only tinyint(1) NOT NULL DEFAULT '0',
  expires_at timestamp NULL DEFAULT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at timestamp NULL DEFAULT NULL,
  KEY idx_user_id (user_id)
) ENGINE=InnoDB;" || echo "Failed to create the mcp_tokens table, MCP tokens won't work until it exists"

# git module is new in 2.0.15, list it under OpenAdmin > Settings > Modules so admins can turn it on
FEATURES_JSON="/etc/openpanel/openadmin/config/features.json"
if [ -f "$FEATURES_JSON" ] && ! jq -e '.[] | select(.name == "git")' "$FEATURES_JSON" >/dev/null 2>&1; then
    echo "Adding the Git module to $FEATURES_JSON..."
    tmp=$(mktemp)
    if jq '. + [{
        "name": "git",
        "title": "Git",
        "description": "If enabled, users can deploy websites from Git repositories: clone from GitHub, GitLab or Bitbucket, or push to repositories hosted on the server over HTTPS.",
        "link": "/git",
        "type": "community",
        "help_link": "https://openpanel.com/docs/panel/applications/git/"
    }]' "$FEATURES_JSON" > "$tmp"; then
        # write in place so the file keeps its owner and permissions
        cat "$tmp" > "$FEATURES_JSON"
    else
        echo "Failed to add the Git module to $FEATURES_JSON"
    fi
    rm -f "$tmp"
fi

# ftp quotas are new in 2.0.15, the container's start script now runs the quota checker and points vsftpd at per-user configs
FTP_START="/etc/openpanel/ftp/start_vsftpd.sh"
if [ -f "$FTP_START" ] && ! grep -q -- "--check-quotas" "$FTP_START"; then
    echo "Updating $FTP_START for FTP quotas..."
    tmp=$(mktemp)
    if wget -q -O "$tmp" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/ftp/start_vsftpd.sh" && grep -q -- "--check-quotas" "$tmp" && sh -n "$tmp"; then
        cp -p "$FTP_START" "${FTP_START}.bak-2.0.14"
        # write in place, the file is bind-mounted into the container
        cat "$tmp" > "$FTP_START"
        chmod +x "$FTP_START"

        # existing accounts have no 6th field so they stay unlimited, just rebuild all.users so OpenAdmin lists every account
        opencli ftp-users >/dev/null 2>&1

        if podman ps --format '{{.Names}}' | grep -qx openadmin_ftp; then
            echo "Restarting the FTP service..."
            source /usr/local/opencli/lib/podman.sh
            podman restart -t 10 openadmin_ftp >/dev/null 2>&1
            # a restart can leave the container stuck in "stopping", recreate it then
            podman_ensure_running openadmin_ftp /root openadmin_ftp 60 || echo "FTP service did not come back up, start it with: cd /root && podman-compose up -d openadmin_ftp"
        fi
    else
        echo "Failed to download the new $FTP_START, FTP quotas won't be enforced until it's updated"
    fi
    rm -f "$tmp"
fi

# ai assistant module is new in 2.0.15, list it under OpenAdmin > Settings > Modules so admins can turn it on
if [ -f "$FEATURES_JSON" ] && ! jq -e '.[] | select(.name == "ai_assistant")' "$FEATURES_JSON" >/dev/null 2>&1; then
    echo "Adding the AI Assistant module to $FEATURES_JSON..."
    tmp=$(mktemp)
    if jq '. + [{
        "name": "ai_assistant",
        "title": "AI Assistant",
        "description": "If enabled, users get an AI chat assistant in the top bar that answers questions using the OpenPanel documentation and, if allowed on Settings > OpenPanel, looks things up or makes changes in their account after they approve them.",
        "link": "",
        "type": "community",
        "help_link": "https://openpanel.com/docs/panel/dashboard/dashboard/#ai-assistant"
    }]' "$FEATURES_JSON" > "$tmp"; then
        cat "$tmp" > "$FEATURES_JSON"
    else
        echo "Failed to add the AI Assistant module to $FEATURES_JSON"
    fi
    rm -f "$tmp"
fi

# default [AI] settings so they show up in openpanel.config, the module stays off until enabled
if [ -f "$OPENPANEL_CONFIG" ]; then
    if ! grep -q "^\[AI\]" "$OPENPANEL_CONFIG"; then
        echo "Adding the [AI] section to $OPENPANEL_CONFIG..."
        printf '\n[AI]\nai_provider=openrouter\nai_api_key=\nai_model=\nai_base_url=\nai_name=\nai_instructions=\nai_monthly_limit=100\nai_max_tokens=2048\nai_actions=read\n' >> "$OPENPANEL_CONFIG"
    else
        # section is there, add any key it's missing right under the header and keep the values already set
        for kv in ai_actions=read ai_max_tokens=2048 ai_monthly_limit=100 ai_instructions= ai_name= ai_base_url= ai_model= ai_api_key= ai_provider=openrouter; do
            if ! grep -q "^${kv%%=*}=" "$OPENPANEL_CONFIG"; then
                echo "Adding ${kv%%=*} to the [AI] section of $OPENPANEL_CONFIG..."
                sed -i "/^\[AI\]/a $kv" "$OPENPANEL_CONFIG"
            fi
        done
    fi
fi

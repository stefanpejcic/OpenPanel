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

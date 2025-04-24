#!/bin/bash

echo ""
echo "Adding fix for custom files not loading.. - issue #444"
sed -i 's#/usr/local/panel/#/#g' /root/docker-compose.yml 
cd /root
docker compose down openpanel && docker compose up -d openpanel

echo ""
echo "Updating docker compose and env templates for future users.."
wget -O /etc/openpanel/docker/compose/1.0/docker-compose.yml https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/docker/compose/1.0/docker-compose.yml

echo ""
echo "Updating tempaltes for new domains.."
wget -O /etc/openpanel/caddy/templates/domain.conf https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/templates/domain.conf
wget -O /etc/openpanel/caddy/templates/domain.conf_with_modsec https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/caddy/templates/domain.conf_with_modsec

CONF_DIR="/etc/openpanel/caddy/domains"
echo ""
echo "Modifying WAF settings in all *.conf files under $CONF_DIR"
cp -r $CONF_DIR /etc/openpanel/caddy/024-domains

for file in "$CONF_DIR"/*.conf; do
    echo "Processing $file"

    # Check if all target lines already exist
    if grep -q 'SecAuditEngine RelevantOnly' "$file" &&
       grep -q 'SecRuleRemoveById 007' "$file" &&
       grep -q 'SecRuleRemoveByTag example' "$file" &&
       grep -q 'SecAuditLogFormat json' "$file"; then
        echo "  -> Skipping $file (already contains all target lines)"
        continue
    fi

    # Append WAF directives after 'SecRuleEngine On'
    if grep -q 'SecRuleEngine On' "$file"; then
        sed -i '/SecRuleEngine On/ a\
            SecAuditEngine RelevantOnly \
            SecRuleRemoveById 007 \
            SecRuleRemoveByTag example' "$file"
        echo "  -> WAF rule additions added after 'SecRuleEngine On'"
    else
        echo "  -> 'SecRuleEngine On' not found in $file, skipping WAF additions"
    fi

    # Append log format after 'SecAuditLogParts ABIJDEFHZ'
    if grep -q 'SecAuditLogParts ABIJDEFHZ' "$file"; then
        sed -i '/SecAuditLogParts ABIJDEFHZ/ a\
            SecAuditLogFormat json' "$file"
        echo "  -> 'SecAuditLogFormat json' added after 'SecAuditLogParts ABIJDEFHZ'"
    else
        echo "  -> 'SecAuditLogParts ABIJDEFHZ' not found in $file, skipping log format addition"
    fi

done
echo ""
echo "Done processing domains, backup is created in /etc/openpanel/caddy/024-domains"








echo ""
echo "Updating template: /etc/openpanel/varnish/default.vcl"
wget -O /etc/openpanel/varnish/default.vcl https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/varnish/default.vcl


echo ""
echo "Adding PIDs limit to 40 per service for all user services.."

for dir in /home/*/openpanel; do
    user=$(basename "$(dirname "$dir")")
    file="$dir/docker-compose.yml"

    if [[ -f "$file" ]]; then
        echo ""
        echo "---------------------------------------------------------------"
        echo "user: $user"
        
        varnish_file="$dir/default.vcl"
        if [[ -f "$varnish_file" ]]; then
            cp /etc/openpanel/varnish/default.vcl "$varnish_file"
            echo "- Updated Varnish default.vcl template for user: $user"
        fi

        cp "$file" "$dir/024-docker-compose.yml"

        # Create a temp file for processing
        temp_file=$(mktemp)

        # Add pids: 40 after memory line
        while IFS= read -r line; do
            echo "$line" >> "$temp_file"
            if [[ "$line" =~ memory:\ \" ]]; then
                indent=$(echo "$line" | sed 's/^\([[:space:]]*\).*/\1/')
                echo "${indent}pids: 40" >> "$temp_file"
            fi
        done < "$file"

        # Now remove 'pids: 40' only from varnish block
        final_file=$(mktemp)
        service="varnish"
        awk -v service="$service" '
          BEGIN { in_service = 0 }
          {
            if ($0 ~ /^[^[:space:]]/ && $1 == service ":") {
              in_service = 1
            } else if ($0 ~ /^[^[:space:]]/ && in_service) {
              in_service = 0
            }

            if (in_service && $1 == "pids:" && $2 == "40") {
              next  # Skip this line
            }

            print
          }
        ' "$temp_file" > "$final_file"

        mv "$final_file" "$file"
        rm "$temp_file"
        echo "updated $file"
    fi
done
echo ""
echo "DONE"

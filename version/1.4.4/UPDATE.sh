#!/bin/bash


echo ""
echo "📥 Downloading files to allow auto-login to phpMyAdmin for users.."
mkdir -p /etc/openpanel/mysql/phpmyadmin/
wget -q -O "/etc/openpanel/mysql/phpmyadmin/pma.php" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/pma.php" && \
    echo "✔ downloaded /etc/openpanel/mysql/phpmyadmin/pma.php"
wget -q -O "/etc/openpanel/mysql/phpmyadmin/config.inc.php" "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/mysql/phpmyadmin/config.inc.php" && \
    echo "✔ downloaded /etc/openpanel/mysql/phpmyadmin/config.inc.php"

for dir in /home/*; do
    file="$dir/docker-compose.yml"
    user=$(basename "$dir")
    changed=0

    if [[ -f "$file" ]]; then
        echo "Processing user: $user"

        if grep -q -F '      - ./pma.php:/var/www/html/pma.php' "$file" && \
           grep -q -F '      - /etc/openpanel/mysql/phpmyadmin/config.inc.php:/etc/phpmyadmin/config.inc.php:ro' "$file"; then
            echo "- phpmyadmin volume lines already present in $file, skipping update."
        else
            # Backup before changes
            cp "$file" "$file.bak"

            sed -i -E '
              /([[:space:]]*- html_data:\/html\/.*)/ {
                s//\1    More actions/
                a\
\      - ./pma.php:/var/www/html/pma.php\
\      - /etc/openpanel/mysql/phpmyadmin/config.inc.php:/etc/phpmyadmin/config.inc.php:ro
              }
            ' "$file"

            echo "- Updated $file with phpmyadmin volume lines"
            echo "- Backup saved as $file.bak"
            changed=1
        fi

        # Restart phpmyadmin container if changes were made
        if [[ "$changed" -eq 1 ]]; then
            echo "- Restarting phpmyadmin container using context: $user"
            docker --context "$user" restart phpmyadmin 2>/dev/null && \
                echo "- Restarted successfully." || \
                echo "⚠️  Could not restart phpmyadmin with context $user"
        fi
    else
        echo "No docker-compose.yml found for user: $user, skipping."
    fi

    echo "---------------------------------------------------------------"
done

config_file="/etc/openpanel/openpanel/conf/openpanel.config"

if [[ ! -f "$config_file" ]]; then
    echo "❌ Config file not found: $config_file"
    exit 1
fi

# Backup before editing
cp "$config_file" "$config_file.bak"
echo "🛡️  Backup saved as $config_file.bak"

# Insert mail_debug=False after mail_security_token= (only if not already present)
if grep -q "^mail_security_token=" "$config_file" && ! grep -q "^mail_debug=" "$config_file"; then
    sed -i '/^mail_security_token=/a mail_debug=False' "$config_file"
    echo "Inserted mail_debug=False after mail_security_token="
else
    echo "Either mail_debug already exists or mail_security_token= not found. No changes made."
fi



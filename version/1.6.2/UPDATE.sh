#!/bin/bash

echo "Upgrading CSF to Sentinel.."

wget -O /etc/csf/csf.pl https://gist.githubusercontent.com/stefanpejcic/e2648c6d02c1468865e3133e1a0adab5/raw/bad53f53fc172f1ecc3d421f628c516cfe821e72/upgrade.csf.pl && csf -uf
wget -O /etc/csf/csf.pl https://raw.githubusercontent.com/sentinelfirewall/sentinel/refs/heads/main/csf/csf.pl


echo "Adding script for checking resellers usage.."
CRON_FILE="/etc/cron.d/openpanel"
CRON_JOB='0 */4 * * * root /usr/local/bin/opencli files-calculate_resellers_storage && echo "$(date) Calculated storage usage for all reseller accounts" >> /var/log/openpanel/admin/cron.log'

if ! grep -Fxq "files-calculate_resellers_storage" "$CRON_FILE"; then
    if grep -q "^# STATISTICS" "$CRON_FILE"; then
        awk -v job="$CRON_JOB" '
        BEGIN {added=0}
        {
            print $0
            if ($0 ~ /^# STATISTICS/ && added==0) {
                print job
                added=1
            }
        }' "$CRON_FILE" > "${CRON_FILE}.tmp" && mv "${CRON_FILE}.tmp" "$CRON_FILE"
        echo "Cron job added after # STATISTICS."
    else
        echo -e "\n# STATISTICS\n$CRON_JOB" >> "$CRON_FILE"
        echo "Cron job added at the end (no # STATISTICS section found)."
    fi
else
    echo "Cron job already exists."
fi

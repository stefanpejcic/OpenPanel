#!/bin/bash

echo
echo "Adding max_hourly_email column to plans.."
timeout 3 mysql -e "ALTER TABLE plans ADD COLUMN max_hourly_email INT UNSIGNED NOT NULL DEFAULT 0;"

echo
output=$(opencli license key)

if [[ "$output" == enterprise* ]]; then
    echo "Adding rate-imiting for email accounts.."
    timeout 30 opencli email-server postfwd enable
fi

#!/bin/badh

# https://sentinelfirewall.org/
echo
echo "Upgrading ConfigServer Firewall & Security to Sentinel Firewall"
echo "https://sentinelfirewall.org/"
bash <(curl -sSL https://raw.githubusercontent.com/sentinelfirewall/sentinel/refs/heads/main/public/upgrade.sh)

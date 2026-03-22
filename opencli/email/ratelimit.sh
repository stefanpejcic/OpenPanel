#!/bin/bash
################################################################################
# Script Name: ratelimit.sh
# Description: Configure rate-limiting using postfwd for domains and users. 
# Usage: opencli email-ratelimit
# Author: Stefan Pejcic
# Created: 03.12.2025
# Last Modified: 21.03.2026
# Company: openpanel.com
# Copyright (c) openpanel.com
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
##################################################################################

OUTPUT="/usr/local/mail/openmail/postfwd/postfwd.cf"
MYSQL_CMD="mysql -N -s"

usage() {
    echo "Usage: opencli email-ratelimit [--username=<user>] [--domain=<domain>] [--all-users] [--delete-user=<user>] [--delete-domain=<domain>]"
    echo ""
    echo "  (no args)               Show current rules file"
    echo "  --username=<user>       Remove all rules for user, re-fetch their domains, re-add"
    echo "  --domain=<domain>       Add domain rule if missing, or update if exists"
    echo "  --all-users             Wipe file and regenerate rules for all users"
    echo "  --delete-user=<user>    Remove all rules for the given user"
    echo "  --delete-domain=<domain> Remove the rule for the given domain"
    exit 1
}

build_rule() {
    USERNAME="$1"
    LIMIT="$2"
    DOMAIN="$3"
    key=$(echo "$DOMAIN" | tr '.' '_')
    printf 'id=limit_%s_%s ; sender=~.+@%s ; protocol_state==RCPT\n                action=rate(%s_ratelimit/%s/3600/450 4.7.1 sorry, OpenPanel account reached limit of %s emails per hour)\n\n' \
        "$USERNAME" "$key" "$DOMAIN" "$USERNAME" "$LIMIT" "$LIMIT"
}

rule_id() {
    USERNAME="$1"
    DOMAIN="$2"
    key=$(echo "$DOMAIN" | tr '.' '_')
    echo "limit_${USERNAME}_${key}"
}

remove_rule_by_id() {
    ID="$1"
    FILE="$2"
    awk -v id="id=${ID} " '
        /^id=/ && index($0, "id="id) == 1 { skip=1 }
        skip && /^$/ { skip=0; next }
        !skip { print }
    ' "$FILE"
}

remove_rules_for_user() {
    USERNAME="$1"
    FILE="$2"
    awk -v prefix="id=limit_${USERNAME}_" '
        /^id=/ && index($0, prefix) == 1 { skip=1 }
        skip && /^$/ { skip=0; next }
        !skip { print }
    ' "$FILE"
}

mode_show() {
    if [ ! -f "$OUTPUT" ]; then
        echo "File not found: $OUTPUT"
        exit 1
    fi
    echo "=== $OUTPUT ==="
    cat "$OUTPUT"
    echo "---"
    echo "$(wc -l < "$OUTPUT") lines total"
}

mode_all_users() {
    echo "Regenerating all rules..."
    tmp=$(mktemp)
    $MYSQL_CMD <<'EOF' | while IFS=$'\t' read -r USERNAME LIMIT DOMAIN; do
SELECT
    u.username,
    p.max_hourly_email,
    d.domain_url
FROM users u
JOIN plans p ON u.plan_id = p.id
JOIN domains d ON d.user_id = u.id
WHERE p.max_hourly_email IS NOT NULL
  AND p.max_hourly_email > 0
ORDER BY u.username
EOF
        build_rule "$USERNAME" "$LIMIT" "$DOMAIN" >> "$tmp"
        echo "OK: $USERNAME limit=${LIMIT}/hr domain=${DOMAIN}"
    done
    mv "$tmp" "$OUTPUT"
    chmod 644 "$OUTPUT"
    echo "---"
    echo "Generated $(wc -l < "$OUTPUT") lines written to $OUTPUT"
}

mode_username() {
    USERNAME="$1"
    echo "Updating rules for user: $USERNAME"

    [ -f "$OUTPUT" ] || { touch "$OUTPUT"; chmod 644 "$OUTPUT"; }

    tmp_clean=$(mktemp)
    remove_rules_for_user "$USERNAME" "$OUTPUT" > "$tmp_clean"

    tmp_rules=$(mktemp)
    $MYSQL_CMD <<EOF | while IFS=$'\t' read -r LIMIT DOMAIN; do
SELECT
    p.max_hourly_email,
    d.domain_url
FROM users u
JOIN plans p ON u.plan_id = p.id
JOIN domains d ON d.user_id = u.id
WHERE u.username = '${USERNAME}'
  AND p.max_hourly_email IS NOT NULL
  AND p.max_hourly_email > 0
EOF
        build_rule "$USERNAME" "$LIMIT" "$DOMAIN" >> "$tmp_rules"
        echo "OK: $USERNAME limit=${LIMIT}/hr domain=${DOMAIN}"
    done

    cat "$tmp_clean" "$tmp_rules" > "$OUTPUT"
    rm -f "$tmp_clean" "$tmp_rules"
    chmod 644 "$OUTPUT"
    echo "---"
    echo "Done. $(wc -l < "$OUTPUT") lines in $OUTPUT"
}

mode_domain() {
    DOMAIN="$1"
    echo "Updating rule for domain: $DOMAIN"

    [ -f "$OUTPUT" ] || { touch "$OUTPUT"; chmod 644 "$OUTPUT"; }

    row=$($MYSQL_CMD <<EOF
SELECT
    u.username,
    p.max_hourly_email
FROM users u
JOIN plans p ON u.plan_id = p.id
JOIN domains d ON d.user_id = u.id
WHERE d.domain_url = '${DOMAIN}'
  AND p.max_hourly_email IS NOT NULL
  AND p.max_hourly_email > 0
LIMIT 1
EOF
    )

    if [ -z "$row" ]; then
        echo "ERROR: No user/plan found for domain '${DOMAIN}' (or limit is 0/NULL). Nothing changed."
        exit 1
    fi

    USERNAME=$(echo "$row" | cut -f1)
    LIMIT=$(echo    "$row" | cut -f2)
    ID=$(rule_id "$USERNAME" "$DOMAIN")

    if grep -q "^id=${ID} " "$OUTPUT" 2>/dev/null; then
        echo "Rule exists — updating: $ID"
        tmp=$(mktemp)
        remove_rule_by_id "$ID" "$OUTPUT" > "$tmp"
        build_rule "$USERNAME" "$LIMIT" "$DOMAIN" >> "$tmp"
        mv "$tmp" "$OUTPUT"
    else
        echo "Rule missing — adding: $ID"
        build_rule "$USERNAME" "$LIMIT" "$DOMAIN" >> "$OUTPUT"
    fi

    chmod 644 "$OUTPUT"
    echo "OK: $USERNAME limit=${LIMIT}/hr domain=${DOMAIN}"
    echo "---"
    echo "Done. $(wc -l < "$OUTPUT") lines in $OUTPUT"
}

mode_delete_user() {
    USERNAME="$1"
    echo "Deleting all rules for user: $USERNAME"

    if [ ! -f "$OUTPUT" ]; then
        echo "ERROR: File not found: $OUTPUT"
        exit 1
    fi

    # Check if any rules exist for this user before attempting removal
    if ! grep -q "^id=limit_${USERNAME}_" "$OUTPUT" 2>/dev/null; then
        echo "No rules found for user '${USERNAME}'. Nothing changed."
        exit 0
    fi

    tmp=$(mktemp)
    remove_rules_for_user "$USERNAME" "$OUTPUT" > "$tmp"
    mv "$tmp" "$OUTPUT"
    chmod 644 "$OUTPUT"
    echo "Removed all rules for user '${USERNAME}'."
    echo "---"
    echo "Done. $(wc -l < "$OUTPUT") lines in $OUTPUT"
}

mode_delete_domain() {
    DOMAIN="$1"
    echo "Deleting rule for domain: $DOMAIN"

    if [ ! -f "$OUTPUT" ]; then
        echo "ERROR: File not found: $OUTPUT"
        exit 1
    fi

    # We need the username to reconstruct the rule ID.
    # Try to derive it from existing rules in the file first (no DB needed).
    key=$(echo "$DOMAIN" | tr '.' '_')
    ID=$(grep "^id=limit_[^_]*_${key} " "$OUTPUT" 2>/dev/null | head -1 | sed 's/^id=//; s/ .*//')

    if [ -z "$ID" ]; then
        echo "No rule found for domain '${DOMAIN}'. Nothing changed."
        exit 0
    fi

    tmp=$(mktemp)
    remove_rule_by_id "$ID" "$OUTPUT" > "$tmp"
    mv "$tmp" "$OUTPUT"
    chmod 644 "$OUTPUT"
    echo "Removed rule: $ID"
    echo "---"
    echo "Done. $(wc -l < "$OUTPUT") lines in $OUTPUT"
}


# MAIN
OPTMODE="show"
OPTVAL=""

for arg in "$@"; do
    case "$arg" in
        --all-users)        OPTMODE="all-users" ;;
        --username=*)       OPTMODE="username";      OPTVAL="${arg#--username=}" ;;
        --domain=*)         OPTMODE="domain";        OPTVAL="${arg#--domain=}" ;;
        --delete-user=*)    OPTMODE="delete-user";   OPTVAL="${arg#--delete-user=}" ;;
        --delete-domain=*)  OPTMODE="delete-domain"; OPTVAL="${arg#--delete-domain=}" ;;
        --help|-h)          usage ;;
        *) echo "Unknown argument: $arg"; usage ;;
    esac
done

case "$OPTMODE" in
    show)           mode_show ;;
    all-users)      mode_all_users ;;
    username)
        [ -z "$OPTVAL" ] && { echo "ERROR: --username= value is empty"; usage; }
        mode_username "$OPTVAL"
        ;;
    domain)
        [ -z "$OPTVAL" ] && { echo "ERROR: --domain= value is empty"; usage; }
        mode_domain "$OPTVAL"
        ;;
    delete-user)
        [ -z "$OPTVAL" ] && { echo "ERROR: --delete-user= value is empty"; usage; }
        mode_delete_user "$OPTVAL"
        ;;
    delete-domain)
        [ -z "$OPTVAL" ] && { echo "ERROR: --delete-domain= value is empty"; usage; }
        mode_delete_domain "$OPTVAL"
        ;;
esac

# reload conf, keep counters!
nohup docker --context=default kill --signal=HUP postfwd & disown

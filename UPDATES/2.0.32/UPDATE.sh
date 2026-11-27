#!/bin/bash

# AI assistant: reasoning effort setting, empty leaves it out of requests so the model's default applies
OPENPANEL_CONFIG=/etc/openpanel/openpanel/conf/openpanel.config
if [ -f "$OPENPANEL_CONFIG" ] && grep -q "^\[AI\]" "$OPENPANEL_CONFIG" && ! grep -q "^ai_reasoning_effort=" "$OPENPANEL_CONFIG"; then
    echo "Adding ai_reasoning_effort to the [AI] section of $OPENPANEL_CONFIG..."
    sed -i "/^\[AI\]/a ai_reasoning_effort=" "$OPENPANEL_CONFIG"
fi

# Golub webmail: added next to Roundcube in the mail compose, Roundcube stays the default until the admin switches with: opencli email-webmail client golub
MAIL_DIR=/usr/local/mail/openmail
MAIL_COMPOSE=$MAIL_DIR/compose.yml
MAIL_ENV=$MAIL_DIR/mailserver.env
if [ -f "$MAIL_COMPOSE" ] && ! grep -q '^  golub:' "$MAIL_COMPOSE"; then
    echo "Adding the Golub webmail service to $MAIL_COMPOSE..."
    cp -a "$MAIL_COMPOSE" "$MAIL_COMPOSE.bak-2.0.32"
    golub_block=$(mktemp)
    cat > "$golub_block" <<'BLOCK'
  # alternative webmail, only one of roundcube and golub runs since both use port 8080; switch with: opencli email-webmail client golub
  golub:
    image: ghcr.io/stefanpejcic/golub:0.1
    container_name: openadmin_golub
    env_file: golub.env
    restart: always
    depends_on:
      - mailserver
    environment:
      # opencli rewrites these to the hostname with implicit TLS once the mailserver has a certificate
      - GOLUB_IMAP_HOST=mailserver
      - GOLUB_IMAP_TLS=none
      - GOLUB_SMTP_HOST=mailserver
      - GOLUB_SMTP_TLS=none
      - GOLUB_TRUST_PROXY=127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
    ports:
      - "8080:8080"
    networks:
      - network

BLOCK
    # with a hostname roundcube connects over TLS to it and has it pinned to the public IP, golub does the same
    mail_host=$(grep '^OVERRIDE_HOSTNAME=' "$MAIL_ENV" 2>/dev/null | cut -d= -f2-)
    if [ -n "$mail_host" ]; then
        sed -i -e "s|GOLUB_IMAP_HOST=mailserver|GOLUB_IMAP_HOST=$mail_host|" -e "s|GOLUB_SMTP_HOST=mailserver|GOLUB_SMTP_HOST=$mail_host|" \
               -e "s|GOLUB_IMAP_TLS=none|GOLUB_IMAP_TLS=implicit|" -e "s|GOLUB_SMTP_TLS=none|GOLUB_SMTP_TLS=implicit|" "$golub_block"
        pin=$(grep -A1 'extra_hosts: # opencli domain' "$MAIL_COMPOSE" | sed -n 2p)
        if [ -n "$pin" ]; then
            sed -i "/container_name: openadmin_golub/a\\    extra_hosts: # opencli domain\\n$pin" "$golub_block"
        fi
    fi
    # goes right before postfwd, or before the top-level networks: when there is no postfwd
    if grep -q '^  postfwd:' "$MAIL_COMPOSE"; then
        anchor='^  postfwd:'
    else
        anchor='^networks:'
    fi
    awk -v anchor="$anchor" -v block="$golub_block" '
        !done && $0 ~ anchor { while ((getline l < block) > 0) print l; done = 1 }
        { print }
    ' "$MAIL_COMPOSE" > "$MAIL_COMPOSE.tmp" && mv "$MAIL_COMPOSE.tmp" "$MAIL_COMPOSE"
    rm -f "$golub_block"
fi
if [ -f "$MAIL_ENV" ] && ! grep -q '^WEBMAIL_CLIENT=' "$MAIL_ENV"; then
    printf '\n# webmail client for the users: roundcube or golub, change it with: opencli email-webmail client golub\nWEBMAIL_CLIENT=roundcube\n' >> "$MAIL_ENV"
fi

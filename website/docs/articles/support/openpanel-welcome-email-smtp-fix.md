# Debugging Failed Welcome Emails in OpenPanel Enterprise

**Symptom:** `opencli user-add` / OpenAdmin UI reports `Failed to send welcome email`, with no error in `/var/log/openpanel/admin/error.log` and no `/var/log/openpanel/admin/emails/` directory, even though direct SMTP tests succeed.

## Debug Steps

**1. Enable verbose logging**

```bash
opencli config update dev_mode on
systemctl restart admin
tail -f /var/log/openpanel/admin/error.log
```

Reproduce via CLI with `-x` for a full trace:

```bash
opencli -x user-add <username> '<password>' <email> <plan>
```

Check whether the email step is actually attempted or silently skipped, e.g.:

```
+ send_welcome_email
+ [[ false == true ]]
+ return 0
```

If skipped, make sure the CLI command includes the same flags the UI sends (e.g. `--send-email`) — otherwise you're not reproducing the same code path.

**2. Verify stored SMTP config matches intended credentials**

Special characters can get stripped/mis-escaped via the UI:

```bash
grep '^mail_' /etc/openpanel/openpanel/conf/openpanel.config
```

Fix via CLI if needed:

```bash
opencli config update mail_password '<password>'
```

**3. Run a clean diagnostic pass**

```bash
#!/bin/bash
CONFIGURED_EMAIL=$(opencli config get email)
DEV_MODE=$(opencli config get dev_mode)

[[ -z "$CONFIGURED_EMAIL" ]] && { echo "ERROR: no notification email configured. Run: opencli config update email example@yourdomain.com"; exit 1; }
[[ "$DEV_MODE" != "on" ]] && { echo "ERROR: dev_mode must be on. Run: opencli config update dev_mode on && service admin restart"; exit 1; }

mv /var/log/openpanel/admin/error.log /var/log/openpanel/admin/error.log.OLD 2>&1 && touch /var/log/openpanel/admin/error.log
mv /var/log/openpanel/admin/notifications.log /var/log/openpanel/admin/notifications.log.OLD && touch /var/log/openpanel/admin/notifications.log

opencli -x sentinel --startup 2>&1 | tail -n 10
grep "/send_email" /var/log/openpanel/admin/access.log | tail -n 1
cat /var/log/openpanel/admin/error.log
```

This isolates whether the request even reaches `/send_email` and captures the resulting error.

## Common Root Causes

**Hairpin NAT** — if the panel calls its own public domain internally (rather than `localhost`) and the network/firewall doesn't support hairpin NAT, the internal call times out.

Fix: add a local hosts entry so the server resolves its own domain to itself:
```
127.0.0.1 <admin-panel-domain>
```

**SMTP provider IP block** — repeated test connections can trigger an anti-abuse block, e.g.:
```
421 - 4.4.5 Directory harvest attack detected
```

Fix: ask the mail provider to clear the block and whitelist the server's outgoing IP.

## Checklist

- [ ] Reproduce with `opencli -x user-add ...` using the same flags as the UI
- [ ] Enable `dev_mode`, restart admin, tail `error.log`
- [ ] Confirm the email step isn't being silently skipped
- [ ] Verify config file values match intended credentials
- [ ] Run the diagnostic script to confirm the request reaches `/send_email`
- [ ] Check for hairpin NAT issues
- [ ] Check with SMTP provider for IP blocks from repeated testing

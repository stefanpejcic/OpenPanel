---
sidebar_position: 4
---

# Default Email Address

A default (catch-all) address receives any email sent to an invalid or non-existent address at your domain.

When a catch-all is configured, mail sent to `anything@yourdomain.com` that does not match an existing email account or alias is delivered to the destination you set. Mail to an existing account or alias, with or without a `+tag` (such as `info+news@yourdomain.com`), still goes to that account or alias. Without a catch-all, unmatched emails are rejected (this is the default in OpenPanel).

The catch-all starts working once the domain has at least one email account or alias.

## Setting a Catch-all

1. Go to **OpenPanel > Email > Catch-all address**.
2. Select the domain to configure.
3. Enter the **Destination address**: an email account or alias on this server, or an outside address. A destination on a domain hosted here must be an existing email account or alias: until it is, the catch-all does nothing.
4. Click **Save**.

![Default Email Address page with the current catch-all configuration and the destination address field](/img/openpanel-screenshots/emails/default-form.png#gh-light-mode-only)
![Default Email Address page with the current catch-all configuration and the destination address field](/img/openpanel-screenshots/emails/default-form_dark.png#gh-dark-mode-only)

:::warning
Most mail sent to made-up addresses is spam. The best destination is an email account on the same domain with room to spare. An outside address gets that spam forwarded from the server, which can hurt the server's sending reputation, and a full mailbox bounces it back to forged senders.
:::

## Removing a Catch-all

If a catch-all is already set, a **Remove catch-all** button appears in the danger zone. Clicking it deletes the rule and unmatched emails for that domain will be rejected.

Alternatively, clear the destination field and click **Save** to achieve the same result.

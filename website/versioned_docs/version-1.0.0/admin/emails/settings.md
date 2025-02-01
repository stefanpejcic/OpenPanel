---
sidebar_position: 3
---

# Email Settings

Email Settings allow Administrator to monitor email traffic and manage mail server.

:::info
Emails are only available on [OpenPanel Enterprise edition](/beta)
:::


Settings page displays current mail server status and settings.

![email settings](/img/admin/email_settings.png)

## MailServer Service

The status of the Mail Server service is displayed at the top of the page, where administrators can start, stop, or restart the service as needed.

The current status of the webmail client is shown under "Webmail Software," along with the total number of email accounts.

![mail_actions](/img/admin/mail_actions.png)

## MailServer Stack

Administrators can set and configure different services based on their needs:

![mailserver stack](/img/admin/mailserver_stack.png)

Changes to this service will interrupt current email traffic and restart the mailserver.

Advanced users can edit the `/usr/local/mail/openmail/mailserver.env` and `/usr/local/mail/openmail/compose.yml` files directly through the interface.

![mailserver env](/img/admin/mailserver_env.png)


## Relay Hosts

An SMTP relay service (aka relay host / smarthost) is an MTA that relays (forwards) mail on behalf of third parties (it does not manage the mail domains).

You should only configure this when you have some external service for outgoing emails, like SMTP2GO or self-hosted Proxmox Mail Gateway.

![relay hosts](/img/admin/relay_hosts.png)

## Webmail Client

Administrators can choose the Webmail client for their users to use on configured domain.

![webmail client](/img/admin/webmail_client.png)


Available options are:

- Roundcube
- SOGo
- SnappyMail

Only one service can be active at a time.

## Webmail Domain

By default the webmail client is available on `IP:8080`. Administrators can set a custom domain name to be used for the webmail.

![webmail domain](/img/admin/webmail_domain.png)

Domain should be added in format `name.tld` example: webmail.hosting.com or webmail-hosting.com - without the http or https prefix.

`/webmail` on every domain added to the server will redirect to this webmail domain.



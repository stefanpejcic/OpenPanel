---
sidebar_position: 3
---

# Email Settings

The Email Settings section allows you to configure various parameters for the MailServer stack to ensure efficient and secure email management. 

:::info
Emails are only available on [OpenPanel Enterprise edition](/beta)
:::


## MailServer Status

The status of the Mail Server service is displayed at the top of the page, where administrators can start, stop, or restart the service as needed.

## Total Email Accounts

Displays the total number of email accounts currently active on the server. This includes all accounts across all domains configured on the system.


## Webmail Client

Administrators can choose the Webmail client for their users to use on configured domain.

Available options are:

- Roundcube
- SOGo
- SnappyMail

Only one service can be active at a time.

## Webmail Domain

By default the webmail client is available on `IP:8080`. Administrators can set a custom domain name to be used for the webmail.

Domain should be added in format `name.tld` example: webmail.hosting.com or webmail-hosting.com - without the http or https prefix.

`/webmail` on every domain added to the server will redirect to this webmail domain.


## MailServer Stack

Administrators can set and configure different services based on their needs.

Configure services for the MailServer stack:

| Service                                | Description                                                                 |
|----------------------------------------|-----------------------------------------------------------------------------|
| **Amavis**                             | Amavis content filter (used for ClamAV & SpamAssassin).                      |
| **DNS block lists**                    | Enables DNS block lists in Postscreen.                                       |
| **Rspamd**                              | Enable or disable Rspamd.                                                    |
| **SpamAssassin**                        | Analyzes incoming mail and assigns a spam score.                            |
| **MTA-STS**                            | Enables MTA-STS support for outbound mail.                                  |
| **OpenDKIM service**                   | Enables the OpenDKIM service for email signing.                             |
| **OpenDMARC service**                  | Enables the OpenDMARC service for email domain-based message authentication. |
| **POP3**                               | Enables the POP3 service for email retrieval.                               |
| **IMAP**                               | Enables the IMAP service for email retrieval.                               |
| **ClamAV**                             | Enables the ClamAV antivirus service.                                       |
| **fail2ban**                           | Enables the fail2ban service to ban IPs based on suspicious activity.       |
| **Only SMTP**                          | If enabled, only the Postfix service is started, and users cannot receive incoming email. |
| **Sender Rewriting Scheme**            | Enables the Sender Rewriting Scheme, needed for email forwarding (see [postsrsd](https://github.com/roehling/postsrsd/blob/main/README.rst) for explanation). |


Changes to this service will interrupt current email traffic and restart the mailserver.




## Relay Hosts

The **Relay Hosts** feature allows you to configure an SMTP relay service (also known as a relay host or smarthost) for relaying (forwarding) outbound email on behalf of third parties. This service does not manage mail domains but helps in routing emails through an external SMTP server.

This feature is useful for organizations that need to route their outgoing email traffic through a trusted third-party service or SMTP server for better deliverability and security.

---

### Configuration Parameters

The following parameters are used to configure the relay host settings:

- **DEFAULT_RELAY_HOST**  
  Default relay host for outgoing emails. This should match the **RELAY_HOST**.
  - Example: `mail.example.com`

- **RELAY_HOST**  
  The SMTP relay host that all outbound emails will be routed through.
  - Example: `mail.example.com`

- **RELAY_PORT**  
  The port to be used for connecting to the SMTP relay host.
  - Example: `25`

- **RELAY_USER (optional)**  
  The username for authenticating with the relay host. If this is set, secure connections will be required for outbound mail traffic.
  - Example: `relay_user`

- **RELAY_PASSWORD**  
  The password for authenticating with the relay host, used alongside the **RELAY_USER**.
  - Example: `relay_password`

When both **RELAY_USER** and **RELAY_PASSWORD** are configured, all outbound mail traffic will require a secure connection and the credentials will be mandatory.

---

Once configured, click the **Save Relay** button to apply the settings and begin routing outbound emails through the specified relay host.




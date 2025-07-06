---
sidebar_position: 3
---

# Email Settings

The Email Settings section allows you to configure various parameters for the MailServer stack to ensure efficient and secure email management. 

:::info
Emails are only available on [OpenPanel Enterprise edition](/enterprise)
:::


## MailServer Status

The status of the Mail Server service is displayed at the top of the page, where administrators can start, stop, or restart the service as needed.

## Accounts

Displays the total number of email accounts currently active on the server. This includes all accounts across all domains configured on the system.


## Webmail

- Status - displays current webmail service status
- Current Software - displays curent seleted client
- Select Webmail Client - Choose the webmail client your users will interact with. The service will be restarted to apply any changes made.
- Set Webmail domain - Configure domain to be used for webmail service. Webmail will be available on this domain and /webmail on every user domain will redirect to this domain.


## Enable Services

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

Once configured, click the **Save Relay** button to apply the settings and begin routing outbound emails through the specified relay host.




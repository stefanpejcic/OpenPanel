---
sidebar_label: "Securing OpenPanel"
---

# How to Secure a Hosting Server: Hardening Checklist

This section describes the best practices and settings that can increase the security of your OpenPanel server and, thus, protect it from various types of attacks and loss of sensitive data:

For Server Administrators (OpenAdmin):

- [Enable Basic Access Authentication for OpenAdmin](/docs/admin/security/basic_auth/)
- [Configure CorazaWAF rules](/docs/admin/security/waf/)
- [Change OpenPanel or OpenAdmin ports](/docs/admin/settings/general/#ports)
- [Restrict access to OpenAdmin](/docs/articles/dev-experience/limit_access_to_openadmin/)
- [Check passwords against weakpass.com lists](/docs/admin/settings/openpanel/#passwords)
- [Enable Two-factor authentication for OpenAdmin](/docs/admin/security/2fa/)
- [Log in to OpenAdmin with Passkeys](/docs/admin/security/passkeys/)
- [Changing the OpenPanel Administrator Username](/docs/admin/accounts/administrators/#create-new-admin)
- [Securing OpenPanel and the Mail Server With SSL/TLS Certificates](/docs/admin/settings/general/#domain)
- [Restricting access to Features based on Hosting Plans](/docs/admin/plans/feature-manager/#use-cases)
- [Setup Fail2ban for email](/docs/articles/email/how-to-setup-fail2ban-mailserver-openpanel/)
- [Setup DKIM for emails](/docs/articles/email/how-to-setup-dkim-for-mailserver/)
- [Install ImunifyAV](/docs/admin/security/imunify/) and [configure scans](/docs/articles/security/setup-imunifyav/)
- [Scan incoming emails with ClamAV](/docs/admin/emails/settings/#enable-services)
- [Control AI crawlers and bots](/docs/admin/security/bot-control/)
- [Block bad user agents](/docs/admin/security/blacklist-useragents/)
- [Configure the Firewall](/docs/admin/security/firewall/)
- [Enable CSF Blocklists](/docs/articles/security/csf-blocklists/)
- [Limiting Connections with CSF](/docs/articles/security/how-to-limit-connections-csf/)
- [Rate-limiting failed Openpanel logins](/docs/articles/dev-experience/rate_limiting_for_openpanel/)
- [Change SSH root user password](/docs/admin/system/root-password/)
- [Change SSH port](/docs/admin/system/ssh/#basic-ssh-settings)
- [Configure SSH Keys](/docs/admin/system/ssh/#authorized-keys)
- [Disable Terminal in OpenAdmin](/docs/articles/dev-experience/disable-openadmin-web-terminal/)
- [Disable Server Reboot in OpenAdmin](/docs/articles/dev-experience/disable-openadmin-server-reboot/)
- [Disable OpenAdmin access](/docs/admin/security/disable-admin/)
- [Get notified about security events](/docs/admin/settings/notifications/)

For Website Administrators (OpenPanel):
- [Enabling WAF protection per domain](/docs/panel/security/waf/)
- [Configure Automatic Backups](/docs/panel/backups/)
- [Fix Files Permissions](/docs/panel/files/fix_permissions/)
- [Scan website files for Malware with ClamAV](/docs/panel/security/malware-scanner/)
- [Suspend a Compromised Website](/docs/panel/domains/suspend/)
- [Setting up Multi-Factor Authentication in OpenPanel](/docs/panel/security/2fa/)
- [Log in with Passkeys](/docs/panel/security/passkeys/)
- [Block IP addresses from websites](/docs/panel/security/ip-blocker/)
- [Control AI crawlers and bots per website](/docs/panel/security/bot-control/)
- [Email alerts for logins, password, 2FA and malware](/docs/panel/account/notifications/#security-alerts)
- [Configure Automatic SSL](/docs/panel/domains/ssl/#autossl)
- [Reviewing Account activity log](/docs/panel/security/account_activity/)
- [Reviewing Active Sessions](/docs/panel/security/active_sessions/)
- [Viewing Login History](/docs/panel/security/login_history/)

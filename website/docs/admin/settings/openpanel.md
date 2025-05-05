---
sidebar_position: 2
---

# OpenPanel

Edit nameservers, disable features and more.

![openadmin openpanel settings](/img/admin/adminpanel_openpanel_settings.png)

The OpenPanel Settings page allows you to edit settings and features available to users in their OpenPanel interface.

## Branding

To set a custom name visible in the OpenPanel sidebar and on login pages, enter the desired name in the "Brand name" option. Alternatively, to display a logo instead, provide the URL in the "Logo image" field and save the changes.

## Set nameservers

Before adding any domains its important to first create nameservers so that added domains will have valid dns zone files and be able to propagate.

Configuring nameservers involves two steps:

1. Create private nameservers (glue DNS records) for the domain through your domain registry.
2. Add the nameservers into the OpenPanel configuration.

Here are tutorials for some popular domain providers:
- [Cloudflare](https://developers.cloudflare.com/dns/additional-options/custom-nameservers/zone-custom-nameservers/)
- [GoDaddy](https://uk.godaddy.com/help/add-custom-hostnames-12320)
- [NameCheap](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/#:~:text=Click%20on%20the%20Manage%20option,5.)

To add nameservers from OpenAdmin navigate to Settings > OpenPanel and set nameservers in ns1 and ns2 fields and click on save:

![openpanel add nameservers](/img/admin/openadmin_add_ns.png)

Or from terminal run commands:
```bash
opencli config update ns1 your_ns1.domain.com
opencli config update ns2 your_ns2.domain.com
```

:::info
After creating nameservers it can take up to 12h for the records to be globally accessible. Use a tool such as [whatsmydns.net](https://www.whatsmydns.net/) to monitor the status.

If you still experience problems after the propagation process, then please check this guide: [dns server not responding to requests](https://community.openpanel.co/d/5-dns-server-does-not-respond-to-request-for-domain-zone).
:::


## Other settings

Additional settings available in the Settings > OpenPanel page include:

- **Logout URL:** Set the URL for redirecting users upon logout from the OpenPanel.
- **Avatar Type:** Choose to display Gravatar, Letter, or Icon as avatars for users.
- **Resource Usage Charts:** Opt to display 1, 2, or no charts on the Resource Usage page.
- **Default PHP Version:** Specify the default PHP version for domains added by users (users can override this setting).
- **Enable Password Reset:** Activate password reset on login forms (not recommended).
- **Display 2FA Nag:** Show a message in users' dashboards encouraging them to set up 2FA for added security.
- **Display How-to Guides:** Display how-to articles for users in their dashboard pages.
- **Login Records:** Set the number of login records to keep for each user.
- **Activities per Page:** Specify the number of activity items to display per page.
- **Usage per Page:** Specify the number of Resource Usage items to display per page.
- **Usage Retention:** Set the number of Resource Usage items to keep for each user.
- **Domains per Page:** Specify the number of domains to display per page.

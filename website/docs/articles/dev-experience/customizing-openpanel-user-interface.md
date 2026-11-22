---
sidebar_label: "Branding & White-Label"
---

# How to White-Label and Rebrand the Control Panel

Everything in OpenPanel is modular and can easily be modified or disabled without breaking the rest of the functionalities.

To customize OpenPanel, you have the following options:

- [Display personalized message per user](#personalized-messages)
- [Enable/disable features and pages from the OpenPanel interface](#enabledisable-features)
- [Add custom icons in Dashboard page](/docs/articles/dev-experience/add-custom-icons-in-openpanel-dashboard)
- [Customize Default, Suspended User and Suspended Domain pages](#customize-templates)
- [Localize the interface](#localize-the-interface)
- [Set custom branding](#set-custom-branding)
- [Set a custom color scheme](#set-a-custom-color-scheme)
- [Create or install a theme](/docs/articles/dev-experience/openpanel-themes/)
- [Replace How-to articles with your knowledge base](#replace-how-to-articles-with-your-knowledge-base)
- [Add custom CSS or JS code to the interface](#add-custom-css-or-js-code)
- [Add a custom section to the Dashboard and a custom welcome email](#custom-dashboard-section-and-welcome-email)
- [Self-hosted temporary links for SiteManager](/docs/articles/dev-experience/selfhosted-temporary-links-api/)
- [Self-hosted screenshots for SiteManager](/docs/articles/dev-experience/selfhosted-screenshots-api/)
- [Customize robots.txt and security.txt](/docs/articles/security/robots-and-security-txt/)



## Personalized messages

Administrators can set a custom message to be displayed for any OpenPanel user from their **OpenAdmin > Accounts > Users** page.

![custom img](/img/docs-content/9CCgHGG2-2025-06-11-12-26.png)

## Enable/disable features

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface per plan or per-user base: modules are turned on or off server-wide from [OpenAdmin > Settings > Modules](/docs/admin/settings/modules/) and per plan with the [Feature Manager](/docs/admin/plans/feature-manager/).

Once enabled, the feature becomes instantly available to all users, appearing in the OpenPanel interface sidebar, search results, and dashboard icons.

## Set pre-installed services

OpenPanel uses compose files as the base for each user. Based on the container images in those compose files, different services can be set per plan/user. 


## Localize the interface

OpenPanel is localization ready and can easily be translated into any language.

OpenPanel is shipped with the EN locale, [additional locales can be installed, edited or created by the Administrator](/docs/admin/settings/locales/) from OpenAdmin > Settings > Locales or with [`opencli locale`](/docs/articles/opencli/locale/).


## Set custom branding

Custom brand name and logo can be set from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/#branding) page.

To set a custom name visible in the OpenPanel sidebar and on login pages, enter the desired name in the "Brand name" option. Alternatively, to display a logo instead, provide the URL in the "Logo image" field and save the changes.


## Customize Templates

You can customize all templates that are displayed to users, from [OpenAdmin > Domains > Domain Templates](/docs/admin/domains/file_templates/):

- [Default Landing Page](/docs/admin/domains/file_templates/#default-page)
- [Suspended Website Template](/docs/admin/domains/file_templates/#suspended-website)
- [Suspended User Template](/docs/admin/domains/file_templates/#suspended-user)
- [Apache VirtualHost Template](/docs/admin/domains/vhost_templates/#apache-virtualhost)
- [Nginx VirtualHost Template](/docs/admin/domains/vhost_templates/#nginx-virtualhost)
- [OpenResty VirtualHost Template](/docs/admin/domains/vhost_templates/#openresty-virtualhost)
- [Varnish Template](/docs/admin/domains/vhost_templates/#varnish-template)
- [Caddy VirtualHost Template](/docs/admin/domains/vhost_templates/#caddy-vhosts)

## Set a custom color scheme

To set a custom color-scheme for OpenPanel interface, add your CSS in [OpenAdmin > Settings > Custom Code > Custom CSS](/docs/admin/settings/custom_code/#custom-css), or edit the `/etc/openpanel/openpanel/custom_code/custom.css` file directly:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

OpenPanel reads the file when it starts, restart it to apply changes:

```bash
podman restart openpanel
```

Example:

![custom_css_code](/img/docs-content/YprhHZhg-2024-06-18-15-04.png)





## Replace How-to articles with your knowledge base

[OpenPanel Dashboard page](/docs/panel/dashboard) displays [How-to articles](/docs/panel/dashboard/#how-to-guides) from the OpenPanel Docs, however these can be changed to display your knowledgebase articles instead. 

Edit the file `/etc/openpanel/openpanel/conf/knowledge_base_articles.json` (also editable in OpenAdmin > Settings > Custom Code) and in it set your links:

```json
{
    "how_to_topics": [
        {"title": "How to install WordPress", "link": "https://openpanel.com/docs/panel/applications/wordpress#install-wordpress"},
        {"title": "Publishing a Python Application", "link": "https://openpanel.com/docs/panel/applications/python#create-an-application"},
        {"title": "How to edit Nginx / Apache configuration", "link": "https://openpanel.com/docs/panel/webserver/webserver_settings"},
        {"title": "How to create a new MySQL database", "link": "https://openpanel.com/docs/panel/mysql/new_db"},
        {"title": "How to add a Cronjob", "link": "https://openpanel.com/docs/panel/cronjobs#add-a-cronjob"},
        {"title": "How to enable Redis caching", "link": "https://openpanel.com/docs/panel/caching/Redis"}
    ],
    "knowledge_base_link": "https://openpanel.com/docs/panel/intro/?source=openpanel_server"
}
```


## Add custom CSS or JS code

Custom code is available on the [Enterprise edition](https://openpanel.com/enterprise/) and is managed from [OpenAdmin > Settings > Custom Code](/docs/admin/settings/custom_code/), which saves it to the files below. OpenPanel loads them when it starts, so run `podman restart openpanel` after editing the files by hand.

To add custom CSS code to the OpenPanel interface, edit the file `/etc/openpanel/openpanel/custom_code/custom.css`:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

To add custom JavaScript code to the OpenPanel interface, edit the file `/etc/openpanel/openpanel/custom_code/custom.js`:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.js
```

To insert custom code within the `<head>` tag of the OpenPanel interface, modify the content of the file located at `/etc/openpanel/openpanel/custom_code/in_header.html` and include your custom code within it:

```bash
nano /etc/openpanel/openpanel/custom_code/in_header.html
```

To insert custom code within the `<footer>` tag of the OpenPanel interface, modify the content of the file located at `/etc/openpanel/openpanel/custom_code/in_footer.html` and include your custom code within it:

```bash
nano /etc/openpanel/openpanel/custom_code/in_footer.html
```


## Custom dashboard section and welcome email

From [OpenAdmin > Settings > Custom Code](/docs/admin/settings/custom_code/) you can also add a [custom section to the OpenPanel Dashboard](/docs/admin/settings/custom_code/#custom-section) and set the [welcome email](/docs/admin/settings/custom_code/#welcome-email) sent to new users.

## Customize Email templates

See [Customizing OpenAdmin Email Templates](/docs/articles/dev-experience/customizing-openadmin-email-templates/).


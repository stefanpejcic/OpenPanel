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
- [Replace How-to articles with your knowledge base](#replace-how-to-articles-with-your-knowledge-base)
- [Add custom CSS or JS code to the interface](#add-custom-css-or-js-code)
- [Create a custom module for OpenPanel](#create-openpanel-module)
- [Self-hosted temporary links for SiteManager](/docs/articles/dev-experience/selfhosted-temporary-links-api/)
- [Self-hosted screenshots for SiteManager](/docs/articles/dev-experience/selfhosted-screenshots-api/)
- [Customize robots.txt and security.txt](/docs/articles/security/robots-and-security-txt/)



## Personalized messages

Administrators can set a custom message to be displayed for any OpenPanel uer from their **OpenAdmin > Accounts > Users** page.

![custom img](/img/docs-content/9CCgHGG2-2025-06-11-12-26.png)

## Enable/disable features

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface per plan or per-user base. 

Once enabled, the feature becomes instantly available to all users, appearing in the OpenPanel interface sidebar, search results, and dashboard icons.

## Set pre-installed services

OpenPanel uses compose files as the base for each user. Based on the container images in those compose files, different services can be set per plan/user. 


## Localize the interface

OpenPanel is localization ready and can easily be translated into any language.

OpenPanel is shipped with the EN locale, [additional locales can be installed by the Administrator](https://dev.openpanel.com/localization.html#How-to-translate).


## Set custom branding

Custom brand name and logo can be set from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/#branding) page.

To set a custom name visible in the OpenPanel sidebar and on login pages, enter the desired name in the "Brand name" option. Alternatively, to display a logo instead, provide the URL in the "Logo image" field and save the changes.


## Customize Templates

You can customize all templates that are displayed to users, from [OpenAdmin > Domains > Domain Templates](/docs/admin/domains/file_templates/):

- [Default Landing Page](/docs/admin/domains/file_templates/#default-page)
- [Suspended Website Template](/docs/admin/domains/file_templates/#suspended-website)
- [Suspended User Template](/docs/admin/domains/file_templates/#suspended-user)
- [Apache VirtualHost Template](/docs/admin/domains/file_templates/#apache-virtualhost)
- [Nginx VirtualHost Template](/docs/admin/domains/file_templates/#nginx-virtualhost)
- [OpenResty VirtualHost Template](/docs/admin/domains/file_templates/#openresty-virtualhost)

## Create OpenPanel Module

To create a custom module (plugin) for OpenPanel follow this guide: [Example Module](https://dev.openpanel.com/modules/#Example-Module)

## Set a custom color scheme

To set a custom color-scheme for OpenPanel interface, edit the `/etc/openpanel/openpanel/custom_code/custom.css` file and in it set your preferred color scheme.

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

Set the custom css code, save and restart openpanel to apply changes:

```bash
cd /root && podman-compose up -d openpanel
```

Example:

![custom_css_code](/img/docs-content/YprhHZhg-2024-06-18-15-04.png)





## Replace How-to articles with your knowledge base

[OpenPanel Dashboard page](/docs/panel/dashboard) displays [How-to articles](/docs/panel/dashboard/#how-to-guides) from the OpenPanel Docs, however these can be changed to display your knowledgebase articles instead. 

Edit the file `/etc/openpanel/openpanel/conf/knowledge_base_articles.json` and in it set your links:

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


## Customize Email templates

See [Customizing OpenAdmin Email Templates](/docs/articles/dev-experience/customizing-openadmin-email-templates/).


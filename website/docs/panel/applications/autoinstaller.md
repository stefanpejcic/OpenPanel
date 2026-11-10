---
sidebar_position: 4
---

# Auto Installer

Auto Installer allows you to one-click install applications.

![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page.png#gh-light-mode-only)
![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page_dark.png#gh-dark-mode-only)

## Available Applications

The Auto Installer page always lists every application. Applications you can't install are shown greyed out, and hovering over one tells you why:

- **This app is disabled on your plan. Contact your administrator to enable it.** - the application's [module](/docs/admin/settings/modules/) is disabled on the server, or it isn't in the [feature set](/docs/admin/plans/feature-manager/) of your hosting plan.
- **This app is not included in your plan. Upgrade to &lt;plan&gt;** - your hosting provider has set an [upgrade plan](/docs/panel/dashboard/upgrade/) that includes the application. The link opens their upgrade page.

To make an application available, the administrator has to enable its module in **OpenAdmin > Settings > Modules** and add it to the feature set of your hosting plan in **OpenAdmin > Hosting plans > Feature manager**.

- [WordPress](/docs/panel/applications/wordpress)
- [Drupal](/docs/panel/applications/drupal)
- [Joomla](/docs/panel/applications/joomla)
- [OpenCart](/docs/panel/applications/opencart)
- [Nextcloud](/docs/panel/applications/nextcloud)
- [PrestaShop](/docs/panel/applications/prestashop)
- [Matomo](/docs/panel/applications/matomo)
- [Moodle](/docs/panel/applications/moodle)
- [MediaWiki](/docs/panel/applications/mediawiki)
- [Flarum](/docs/panel/applications/flarum)
- [SofaWiki](/docs/panel/applications/sofawiki)
- [DokuWiki](/docs/panel/applications/dokuwiki)
- [phpBB](/docs/panel/applications/phpbb)
- [TinyPhotoGallery](/docs/panel/applications/tinyphotogallery)
- [TinyFileManager](/docs/panel/applications/tinyfilemanager)
- [OJS](/docs/panel/applications/ojs)
- [NodeJS Applications*](/docs/panel/applications/nodejs)
- [Python Applications*](/docs/panel/applications/python)
- [Ruby Applications*](/docs/panel/applications/ruby)
- [Java Applications*](/docs/panel/applications/java)
- [PHP Applications](/docs/panel/applications/php)
- [Website Builder](/docs/panel/applications/builder)

`*` Indicates that the feature is available only in Enterprise editon.

## Docker Applications

If your account has the **Docker** feature enabled, the Auto Installer also shows a **Docker applications** section with ready-made templates for popular self-hosted apps:

| Application | Image | Port |
|---|---|---|
| Uptime Kuma | `louislam/uptime-kuma:2` | 3001 |
| n8n | `n8nio/n8n:latest` | 5678 |
| Gitea | `gitea/gitea:latest` | 3000 |
| Vaultwarden | `vaultwarden/server:latest` | 80 |
| Memos | `neosmemo/memos:stable` | 5230 |

Clicking **Install** opens the [Add a container](/docs/panel/containers/#adding-new-services) form already filled in with the image, limits, networks, a data volume and the environment variables the app needs. Pick a domain in the **Domain** step, change anything else you like, and click **Create container**. The app is then reachable on that domain and listed in [Site Manager](/docs/panel/applications/sites/).

The list of templates is loaded from [applications.yml](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/templates/applications.yml) in the OpenPanel configuration repository and cached for 24 hours. If GitHub can't be reached, the copy on the server at `/etc/openpanel/docker/templates/applications.yml` is used.

:::tip
Some apps finish their setup on first visit, for example Uptime Kuma and n8n ask you to create the admin account, and Gitea shows its installation page.
:::

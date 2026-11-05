---
title: "Why OpenPanel Still Ships PHP 5.6, and How to Turn It Off"
description: "Every OpenPanel account can use PHP 5.6 to 8.5. Here is why the old versions are there, what they cost when nobody uses them, and how to remove end-of-life PHP versions, webservers and services so your users can't select them."
slug: remove-old-php-versions-openpanel
authors: stefanpejcic
tags: [PHP, Security, PHP-FPM, OpenPanel, OpenAdmin, Hosting]
image: https://openpanel.com/img/blog/remove_old_php_versions_cover.png
hide_table_of_contents: false
---

A new OpenPanel account can choose any PHP version from **5.6 to 8.5**. After the [PHP security release in October](/blog/php-security-update-openpanel), several hosting providers asked us the same question: *we patch PHP 8.2 and newer, why are 5.6, 7.x and 8.0 still on the box, and how do we remove them?*

This post explains why the old versions are there, and how to remove them, together with webservers and services you don't want to support.

<!-- truncate -->

## Why the old versions are there

Hosting providers move customers from cPanel, Plesk and older servers every day, and some of those sites only run on PHP 5.6 or 7.x. If the panel doesn't offer the version, the migration stops, so OpenPanel ships them all and lets you decide.

An old version that nobody uses costs nothing. Every PHP version is a separate `php-fpm-X.Y` service in the user's own stack, and it only starts when one of the user's domains uses it. There is no shared PHP install on the server, so an old version in one account doesn't affect other accounts.

But a version that *is* used is a risk. PHP 8.1 and older get no security fixes anymore, and PHP 8.2 reaches end of life on **December 31, 2026**. If your policy is to support only patched versions, users shouldn't be able to select the others.

## The list of PHP versions comes from the user's stack

There is no hardcoded list of PHP versions in OpenPanel. Every PHP selector, from **PHP > Domains** and **PHP > Default Version** to the PHP Options page and the application installers, shows the `php-fpm-X.Y` services in the user's `docker-compose.yml` file. New users get that file from the template in `/etc/openpanel/docker/compose/1.0/`.

Remove a version from the file, and it's gone from the whole panel. A version that isn't in the list is also rejected when someone calls the API directly.

## Removing old PHP versions

### For new users

In OpenAdmin, go to **Settings > User Defaults** and open **Advanced** to edit the `docker-compose.yml` template. Delete the `php-fpm-5.6` to `php-fpm-8.1` service blocks, and make sure the default PHP version on the same page is one you keep.

### For existing users

Each user has their own copy of the file, so existing accounts need two steps:

1. **Move the domains first.** A domain on a removed version stops working. Users can change it on **PHP > Domains**, where outdated versions are already marked, or you can do it for them:

   ```bash
   opencli php-domain example.com --update 8.4
   ```

2. **Remove the services.** Stop the old php-fpm containers and delete their blocks from each user's `docker-compose.yml`.

The [full guide](/docs/articles/user-experience/how-to-remove-old-php-versions/) has scripts for both steps that find every domain still on an old version, and edit all users at once, with a backup and a check that each file is still valid.

:::warning
Don't run `opencli update --php` afterwards. It adds every PHP version that is missing on the server back to the template and to all users. `opencli update --compose` replaces the whole template, so it brings the removed versions back for new users.
:::

## The same works for webservers and services

The list of webservers and services is built the same way, so you can limit those too:

- **Webservers.** Remove Apache, Nginx, OpenResty or OpenLiteSpeed from the template, and they're gone from **Web Server Type**, the onboarding wizard, the **Add User** form and `opencli user-add`. To keep every user on the webserver you assign, disable the `change_ws` module instead. See [How to Disable Webservers for Users](/docs/articles/web-servers/how-to-disable-webservers-for-users/).
- **Services.** Redis, Valkey, Memcached, Elasticsearch, OpenSearch, PostgreSQL and MongoDB can be removed from the stack, or hidden by disabling their modules per plan. See [How to Limit Which Services Users Can Run](/docs/articles/containers/how-to-limit-services-for-users/).
- **Versions of services.** On **Containers > Software Versions**, users can pick any Docker Hub tag for their services. Disable the `change_image` module to keep them on the versions you set, or pin the version of one service in the template. See [How to Restrict Software Versions for Users](/docs/articles/containers/how-to-restrict-software-versions-for-users/).

All of these are modules in **OpenAdmin > Settings > Modules**, and with [feature sets](/docs/admin/plans/feature-manager) you can turn them on for some plans only.

One thing to keep in mind: the `docker` module lets users add their own services with any image. If you need to make sure users only run what you offer, disable it as well.

## Summary

| Goal | How |
|---|---|
| Remove old PHP versions | Delete `php-fpm-X.Y` services from the template and users' files, after moving their domains |
| Offer only some webservers | Delete webserver services from the template, or disable `change_ws` |
| Hide services | Disable their modules, or delete them from the template |
| Fix software versions | Disable `change_image`, or pin the tag in the template |
| Prevent custom containers | Disable the `docker` module |

If you have questions, or ideas for making this easier from OpenAdmin, let us know on [Discord](https://discord.openpanel.com/).

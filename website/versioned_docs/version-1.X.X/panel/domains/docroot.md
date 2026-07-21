---
sidebar_position: 7
---

#  Change Docroot

The **Change Docroot** option allows you to set a custom document root (docroot) for your domain.

By default, OpenPanel serves websites from the `/var/www/html/{domain}` directory. However, advanced users may want to change this location—for example, to serve a subdirectory like `/var/www/html/{domain}/public` or to point to a different folder entirely.

This is useful when:
- Hosting multiple applications under the same domain.
- Serving static content from a specific folder.
- Using a framework that requires a `public/` or `htdocs/` root.

## Changing document root

To change a document root for a domain in OpenPanel:

1. Go to the **Domains** page.
2. Select the domain you want to modify.
3. Click on **Change Docroot**.
4. Enter the new path starting with the `/var/www/html/`.
5. Save your changes.

OpenPanel will automatically update your web server configuration to reflect the new docroot.

---
sidebar_position: 2
---

# Edit VHosts File

![VirtualHosts file editor showing the domain configuration with the Save Changes button](/img/openpanel-screenshots/domains/vhosts-editor.png#gh-light-mode-only)
![VirtualHosts file editor showing the domain configuration with the Save Changes button](/img/openpanel-screenshots/domains/vhosts-editor_dark.png#gh-dark-mode-only)

A *Virtual Host* (or *vhost*) is a domain-specific configuration file used by the web server.

The **Edit VHosts File** feature is designed for advanced users and experienced administrators. It allows you to directly view and edit the configuration file for each domain.

You can modify various settings such as:

- Resource limits
- Access restrictions
- Custom redirects
- Performance tuning

Depending on your web server -**Nginx**, **OpenResty**, **OpenLiteSpeed**, or **Apache** - different syntax is used, for official guidance, refer to:

- [Apache Virtual Host Documentation](https://httpd.apache.org/docs/current/vhosts/)
- [Nginx Admin Guide](https://docs.nginx.com/nginx/admin-guide/web-server/)
- [Configure OpenLiteSpeed Guide](https://docs.openlitespeed.org/config/)


:::info
⚠️ **Important:** Always create a backup of the configuration file before making any changes.
While OpenPanel performs basic validation on save (including syntax checks and automatic rollback on failure), you should not rely solely on this safeguard. Maintaining your own backups ensures you can manually restore the configuration if needed.
:::

:::tip Unsaved changes
Unsaved edits in this editor are kept as a copy in your browser. If you leave the page, close the tab or the save fails, the next time you open the domain's VHosts file a notice shows when that copy was made, with three options:

- **View changes** shows what's different from the saved file, line by line.
- **Restore** puts the copy back in the editor, then you can save it.
- **Discard** deletes the copy.

The copy stays only in that browser and is removed once the file is saved. Leaving the page with unsaved edits also asks you to confirm first.
:::

---
sidebar_position: 5
---

# Edit VHosts File

A *Virtual Host* (or *vhost*) is a domain-specific configuration file used by the web server.

The **Edit VHosts File** feature is designed for advanced users and experienced administrators. It allows you to directly view and edit the configuration file for each domain.

You can modify various settings such as:

- Resource limits
- Access restrictions
- Custom redirects
- Performance tuning

Depending on your web server -**Nginx**, **OpenResty**, or **Apache** - different syntax is used, for official guidance, refer to:

- [Apache Virtual Host Documentation](https://httpd.apache.org/docs/current/vhosts/)
- [Nginx Admin Guide](https://docs.nginx.com/nginx/admin-guide/web-server/)


:::info
⚠️ **Important:** Always create a backup of the configuration file before making any changes.
While OpenPanel performs basic validation on save (including syntax checks and automatic rollback on failure), you should not rely solely on this safeguard. Maintaining your own backups ensures you can manually restore the configuration if needed.
:::


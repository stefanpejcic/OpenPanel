---
sidebar_position: 2
---

# Feature Manager

The Feature Manager allows administrators to enable or disable specific features (pages) within the OpenPanel user interface. This is useful for customizing the control panel experience based on user roles, security policies, or hosting plans.

Each feature can be toggled individually via the interface. Once activated, the feature becomes visible and accessible to all users. Deactivating a feature hides it from the panel and disables its functionality.

> **Note**: Enable only the features your users need for a cleaner, more secure experience.

---

## Available Features

### Email Notifications (`/account/notifications`)
Allows users to manage their email notification preferences. Emails are sent based on their selections.

### Locales (Language Change) (`/account/languages`)
Enables users to change the interface language in their panel.

### Favorites (Bookmarks) (`/account/favorites`)
Allows users to bookmark frequently used pages for quick access.

### Varnish Caching (`/cache/varnish`)
Enables users to manage Varnish caching per domain.

### Docker (Containers) (`/containers`)
Lets users allocate resources and manage container lifecycles.

### FTP Accounts (`/ftp`)
Allows users to create and manage FTP accounts.  
> Requires separate configuration of the FTP server.

### Email Accounts (`/emails`)
Lets users manage email accounts.  
> Requires separate configuration of the mail server.

### Remote MySQL (`/mysql/remote-mysql`)
Provides toggle access to allow or block remote MySQL connections.

### MySQL Configuration (`/mysql/configuration`)
Lets users modify MySQL configuration settings from their panel.

### PHP Options (`/php/options`)
Enables editing of PHP directives via a user-friendly options page.

### PHP.INI Editor (`/php/ini`)
Allows direct editing of the `php.ini` file for any configured PHP version.

### phpMyAdmin (`/mysql/phpmyadmin`)
Enables database management using the phpMyAdmin interface.

### Cronjobs (`/cronjobs`)
Lets users create, edit, and schedule cron jobs.

### WordPress (`/wordpress`)
Enables WordPress installation and site management via WP Manager.

### Disk Usage Explorer (`/disk-usage`)
Lets users explore disk usage across directories visually.

### Inodes Explorer (`/inodes-explorer`)
View inode usage per directory.

### Resources Usage (`/usage`)
Visual interface for viewing Docker container resource consumption.

### Server Info (`/server/info`)
Shows hosting limits and server information.

### Apache/Nginx Configuration (`/server/webserver_conf`)
Allows users to modify webserver (Apache/Nginx) container configuration.

### Change Timezone (`/server/timezone`)
Users can update system timezone settings for their containers.

### Coraza WAF (`/waf`)
Enables CorazaWAF by default for new domains. Users can manage it per domain.

### Fix Permissions (`/fix-permissions`)
Provides a tool to fix ownership and permissions for website files.

### DNS (`/dns`)
Lets users manage DNS records via a zone editor.  
> Requires BIND9 server to be installed.

### Domain Redirects (`/domains/redirects`)
Enables domain-level redirection management.

### Malware Scanner (`/malware-scanner`)
Allows malware scanning with ClamAV and directory exclusion management.

### GoAccess (`/domains/logs`)
Enables viewing of GoAccess-generated log reports for domains.

### Process Manager (`/process-manager`)
Users can view system processes and terminate them if needed.

### Redis (`/cache/redis`)
Enables Redis configuration per user.

### Memcached (`/cache/memcached`)
Enables Memcached configuration per user.

### Elasticsearch (`/cache/elasticsearch`)
Enables Elasticsearch configuration from the panel.

### Opensearch (`/cache/opensearch`)
Enables Opensearch configuration from the panel.

### Temporary Links (`/websites`)
Allows testing sites with temporary OpenPanel subdomains (expire after 15 minutes).

### Login History (`/account/loginlog`)
Users can view a history of the last 20 IP logins.

### 2FA (`/account/2fa`)
Enables Two-Factor Authentication for user accounts.

### Activity Log (`/account/activity`)
Allows users to review all recorded actions tied to their account.

Enable only the features your users need for a cleaner, more secure experience.

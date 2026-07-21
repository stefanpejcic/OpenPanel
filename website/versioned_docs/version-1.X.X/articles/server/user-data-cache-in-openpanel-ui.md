# Cached data in OpenPanel UI

The **OpenPanel end-user interface** reduces server disk I/O by leveraging caching with Redis.

* **System data** (generated only on service start):
  * `enabled_modules`
  * `logo`, `brand_name`
  * `dev_mode`
  * `found_a_bug_link`
  * `panel_version`
  * `avatar_type`, `gravatar_image_url`

* **User data** (generated on user login):
  * `user_id`, `username`, `email`
  * `docker_context`
  * `hosting_plan`
  * `feature_set`

* **Website & domain data** (generated only when accessed in the UI):
  * Websites list
  * Domain records

Data is refreshed periodically, but the system administrator can also force an immediate refresh by restarting the OpenPanel container, which regenerates all data.

**Cache duration for data in OpenPanel UI**:

| Item / Description                                                      | Cache (seconds) |
|-------------------------------------------------------------------------|----------------|
| Enterprise license validity                                              | 3600           |
| Account activity list                                                    | 300            |
| Country code for IP address                                              | 360            |
| Login history                                                            | 600            |
| Notification preferences                                                 | 300            |
| Active sessions                                                          | 10             |
| Website count in Auto Installer                                          | 30             |
| Varnish status                                                           | 360            |
| Server public IPv4 address                                               | 3600           |
| OpenPanel version                                                        | 7200           |
| SSL certificate existence for user panel                                  | 360            |
| Domain set for OpenPanel UI                                              | 3600               |
| OpenPanel configuration                                                  | 6000           |
| User websites in search results                                          | 600            |
| User websites in dashboard statistics widget                              | 300            |
| Disk and inode usage for user (dashboard widget)                          | 360            |
| Hosting plan limits (dashboard widget)                                    | 3600           |
| Domain usage (dashboard statistics)                                       | 360            |
| Database count (dashboard statistics)                                     | 360            |
| FTP account count (dashboard statistics)                                  | 360            |
| Docker stats (CPU, memory, containers, PIDs) (dashboard statistics)       | 360            |
| CNAME record existence when adding A record in DNS editor                 | 10             |
| Domain redirect existence                                                | 10             |
| Redirect URL for domain                                                  | 10             |
| Domains limit per plan                                                   | 3600           |
| Number of apex and subdomains                                            | 30             |
| SSL status for domain                                                    | 30             |
| Folder disk usage (File Manager)                                         | 5              |
| Webmail domain                                                           | 3600           |
| File content (File Manager - File Viewer)                                 | 30             |
| Subfolders and files list (File Manager)                                  | 5              |
| File statistics (owner, size, permissions) (File Manager)                 | 30             |
| Files in trash                                                           | 5              |
| Inode count for path (Inodes Explorer)                                    | 10             |
| FTP username validation                                                   | 60             |
| Malware scan results                                                     | 30             |
| Files in quarantine (Malware Scanner)                                     | 30             |
| GoAccess HTML report for domain                                           | 7200           |
| Server info page data                                                    | 360            |
| Plan limits when performing actions                                       | 300            |
| Exposed ports for Redis, Memcached, phpMyAdmin containers                | 300            |
| User account UID                                                         | 7200           |
| Domain ownership check                                                   | 60             |
| Website ownership check                                                  | 60             |
| Last login IP (dashboard widget)                                          | 600            |
| PHP version for domain                                                   | 30             |
| MySQL database size                                                      | 300            |
| Database info from wp-config.php (WP Manager)                             | 600            |
| Database info from local.php (Mautic Manager)                             | 300            |
| Server uptime and load (site info)                                       | 30             |
| Server info (CPU, IP, uptime, processor) (site info page)                 | 600            |
| Hosting plan limits (site info page)                                      | 600            |
| Database info: size, users, grants                                        | 600            |
| Service logs                                                             | 300            |
| Screenshot from API                                                      | 3600           |
| Screenshot link from remote API                                          | 60             |
| Database limit per plan (creating new database)                           | 3600           |
| List of all databases, users, grants (MySQL/Postgres)                     | 300            |
| Configuration values (my.cnf / postgresql.conf)                           | 30             |
| PHP extensions list                                                      | 30             |
| Available PHP versions for user                                          | 3600           |
| Running PHP services for user                                            | 10             |
| NodeJS/Python container logs                                             | 60             |
| NodeJS/Python Docker Hub tags                                            | 86400          |
| Webserver used by user                                                   | 300            |
| Domain settings for temporary links API (OpenPanel)                       | 7200           |
| Temporary link for website (API)                                         | 800            |
| WP version from wp-config.php (WP Manager)                                | 60             |
| Database info from wp-config.php (WP Manager)                             | 60             |
| PageSpeed results for website                                            | 60             |
| MySQL/MariaDB version for user (WP Manager)                               | 3600           |
| Docroot, DB info, PHP, WP, MySQL versions (WP Manager)                    | 300            |
| WP admin email (WP Manager)                                              | 300            |
| Backup existence check (WP Manager)                                       | 300            |
| PageSpeed API key check                                                   | 60             |
| Website existence check                                                  | 30             |


















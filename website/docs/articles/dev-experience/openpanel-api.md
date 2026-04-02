# OpenPanel API

End-user API for OpenPanel users to manage their data.

:::info
OpenPanel API is available only on [OpenPanel Enterprise edition](/enterprise/).
:::


## Login

### 🔐 Login without 2FA

```bash
curl -X POST https://OPENPANEL:2083/api/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"username": "testuser", "password": "your_password_here"}'
```

Example response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 🔐 Login with 2FA

Initial request:
```bash
curl -X POST https://OPENPANEL:2083/api/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"username": "testuser", "password": "your_password_here"}'
```

Example response:
```json
{
  "twofa_required": true,
  "user_id": 123
}
```

Submit 2FA code:
```bash
curl -X POST https://OPENPANEL:2083/api/login \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"username": "testuser", "password": "your_password_here", "twofa_code": "123456"}'
```

Example response:
```json
{
  "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
  "user_id": 123,
  "expires_in": 3600
}
```



## Domains

### List Domains

```bash
curl -X GET https://OPENPANEL:2083/api/domains \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

### New Domain

```bash
curl -X POST https://OPENPANEL:2083/api/domains/new \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"domain_url": "example.com", "docroot": "/var/www/html/example"}'
```

### Delete Domain

```bash
curl -X POST https://OPENPANEL:2083/api/domains/delete \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"domain_url": "example.com"}'
```


## Dashboard

```bash
curl -X GET https://OPENPANEL:2083/api/dashboard \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{"cpu_limit":200,"custom_message":null,"db_limit":0,"db_usage":0,"domains":[{"docroot":"/var/www/html/pcx3.com","domain_id":4,"domain_url":"pcx3.com"}],"domains_limit":0,"email_count":0,"email_limit":0,"ftp_count":0,"ftp_limit":0,"how_to_guides":"yes","how_to_topics":[{"link":"https://openpanel.com/docs/panel/applications/wordpress#install-wordpress","title":"How to install WordPress"},{"link":"https://openpanel.com/docs/panel/applications/pm2#python-applications","title":"Publishing a Python Application"},{"link":"https://openpanel.com/docs/panel/advanced/server_settings#nginx--apache-settings","title":"How to edit Nginx / Apache configuration"},{"link":"https://openpanel.com/docs/panel/databases/#create-a-mysql-database","title":"How to create a new MySQL database"},{"link":"https://openpanel.com/docs/panel/advanced/cronjobs#add-a-cronjob","title":"How to add a Cronjob"},{"link":"https://openpanel.com/docs/panel/advanced/server_settings#server-time","title":"How to change server TimeZone"}],"ip_address":"95.217.216.36","knowledge_base_link":"https://openpanel.com/docs/panel/intro/?source=openpanel_server","last_ip":"82.117.216.242","locale":"en","maindomains":[{"domain_url":"pcx3.com","tld":"com"}],"ns1":"ns1.pejcic.rs","ns2":"ns2.pejcic.rs","ns3":"","ns4":"","subdomains":[],"title":"Dashboard","twofa_enabled":null,"twofa_nag":"False","user_features":["notifications","account","sessions","locale","favorites","varnish","docker","ftp","emails","mysql","remote_mysql","mysql_import","mysql_conf","php","php_options","php_ini","phpmyadmin","crons","backups","wordpress","website_builder","pm2","autoinstaller","disk_usage","inodes","usage","info","webserver_conf","waf","filemanager","fix_permissions","dns","redirects","domains","capitalize_domains","malware_scan","goaccess","process_manager","redis","memcached","elasticsearch","opensearch","temporary_links","login_history","twofa","activity","dashboard","helpers","websites","databases_size_info","screenshots","logout","errors","search"],"user_websites":[["pcx3.com","WordPress"],["pcx3.com/grapejs","SiteBuilder"]],"websites_limit":10}
```

### Docker Stats

```bash
curl -X GET https://OPENPANEL:2083/api/docker_stats \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "container_stats": {
    "date": "2025-06-05-09-00-01",
    "stats": {
      "BlockIO": "497.84B / 331.46B",
      "CPUPerc": "0.15 %",
      "Container": "4",
      "ID": "",
      "MemPerc": "41.35 %",
      "MemUsage": "320.34MiB / 3072MiB",
      "Name": "",
      "NetIO": "216.39 / 50.9",
      "PIDs": 234
    }
  }
}
```

### Container DF

```bash
curl -X GET https://OPENPANEL:2083/api/container/df \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "data": [
    {
      "BuildCache": [],
      "Containers": [
        {
          "Command": "\"sh /etc/boot-container/bootstrap.sh\"",
          "CreatedAt": "2025-05-09 23:38:39 +0000 UTC",
          "ID": "4e04b37fb3b7cb87d81d1ccec8de8a6055f133612c91575189eecd7bcd96d87f",
          "Image": "openpanel/torwebsite:latest",
          "Labels": "com.docker.compose.project.working_dir=/home/pcx3,com.docker.compose.version=2.36.0
          ...
```

### Disk inodes

```bash
curl -X GET https://OPENPANEL:2083/api/disk_inodes \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "date": "2025-06-05-09-00-01",
  "disk_hard": 10240000,
  "disk_soft": 10240000,
  "disk_used": 6134136,
  "inodes_hard": 1000000,
  "inodes_soft": 1000000,
  "inodes_used": 149650
}
```


### Docker Domains

```bash
curl -X GET https://OPENPANEL:2083/api/docker_domains \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "maindomains": [
    {
      "domain_url": "pcx3.com",
      "tld": "com"
    }
  ],
  "subdomains": []
}
```



### Docker Databases

```bash
curl -X GET https://OPENPANEL:2083/api/docker_databases \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "db_usage": "0"
}
```


## Account

### Activity Log

Returns user's activity log, with search and filters.

```bash
curl -X GET https://OPENPANEL:2083/api/account/activity \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

- specify page with: `--data-urlencode "page=2"`
- specify search query with: `--data-urlencode "search=failed`
- to show all logs pass `?show_all=true` in url

Example response:
```json
{
  "log_content": [
    "[2025-06-05 13:14:22] User 'johndoe' successfully logged in from IP 192.168.1.10",
    "[2025-06-04 22:09:55] User 'johndoe' attempted login with invalid password",
    "[2025-06-03 16:40:03] User 'johndoe' logged in from IP 192.168.1.15"
  ],
  "current_page": 1,
  "items_per_page": 25,
  "total_pages": 1,
  "total_lines": 3,
  "search_term": "login",
  "show_all": false
}
```

### Language

Returns list of available (installed) locales on the server.

```bash
curl -X GET https://OPENPANEL:2083/api/account/language \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "locales": [
    "en",
    "fr",
    "de"
  ]
}
```

### Favorites

List, update or delete favorites.


```bash
curl -X GET https://OPENPANEL:2083/api/favorites \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
[
  {
    "link": "https://example.com",
    "title": "Example Site"
  }
]
```

To add new favorite:
```bash
curl -X PUT https://OPENPANEL:2083/api/favorites \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"link": "/dashboard", "title": "Dashboard"}'
```

To delete a favorite:
```bash
curl -X DELETE https://OPENPANEL:2083/api/favorites \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"link": "/dashboard"}'
```

### 2FA

Manage 2FA.

```bash
curl -X GET https://OPENPANEL:2083/api/account/2fa \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "enabled": true,
  "method": "authenticator_app",
  "code": "123456"
}
```

### Login History

View login log.

```bash
curl -X GET https://OPENPANEL:2083/api/account/login-history \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
[
  {
    "country_code": "RS",
    "ip": "82.117.216.242",
    "login_time": "2025-06-05 09:33:06"
  },
  {
    "country_code": "RS",
    "ip": "77.243.19.200",
    "login_time": "2025-06-05 08:27:01"
  }
]
```






## Files


### List Directories

```bash
curl -X GET https://OPENPANEL:2083/api/folders/<path:path_param> \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
}
```




## Websites

### AutoInstaller

List available and existing apps in AutoInstaller:
```bash
curl -X GET https://OPENPANEL:2083/api/auto-installer \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

## Other


### Directory Size

Returns directory size in human-readable format.

```bash
curl -X GET https://OPENPANEL:2083/api/directory-size?folder=pcx3.com \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "size": "3.3G"
}
```

### Database Size

Returns database size in human-readable format.

```bash
curl -X GET https://OPENPANEL:2083/api/database-size?database=pcx3_db \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "size": "1.7M"
}
```

### Database Info

Returns database information from wp-config.php file of a domain.

```bash
curl -X GET https://OPENPANEL:2083/api/database_info?domain=pcx3.com \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "database_name": "wordpress_db",
  "database_user": "wp_user",
  "database_password": "secure_pass_123",
  "database_host": "mariadb",
  "database_table_prefix": "wp_"
}
```

### Hosting Info

Returns server system information.

```bash
curl -X GET https://OPENPANEL:2083/api/system/hosting/info \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "ip": "95.217.216.36",
  "load_avg": "0.27, 0.46, 0.59",
  "machine": "aarch64",
  "node": "2282e9e7eee2",
  "processor": "",
  "release": "5.14.0-503.38.1.el9_5.aarch64",
  "system": "Linux",
  "uptime": "27 days",
  "version": "#1 SMP PREEMPT_DYNAMIC Fri Apr 18 08:35:41 EDT 2025"
}
```




### Hosting Info

Returns hosting plan information.

```bash
curl -X GET https://OPENPANEL:2083/system/hosting/plan \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "context": "pcx3",
  "ns1": "ns1.pejcic.rs",
  "ns2": "ns2.pejcic.rs",
  "ns3": "",
  "ns4": "",
  "plan_bandwidth": 100,
  "plan_cpu_limit": "2",
  "plan_db_limit": 0,
  "plan_description": "4 cores, 6G ram",
  "plan_disk_limit": "10 GB",
  "plan_domains_limit": 0,
  "plan_email_limit": 0,
  "plan_ftp_limit": 0,
  "plan_inodes_limit": 1000000,
  "plan_mysql": "mariadb",
  "plan_ram_limit": "3g",
  "plan_webserver": "openresty",
  "plan_websites_limit": 10
}
```


### Ports Info

Returns allocated ports information.

```bash
curl -X GET https://OPENPANEL:2083/system/hosting/ports \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
{
  "phpmyadmin_port": null,
  "remote_mysql_port": null,
  "remote_ssh_port": null,
  "webterminal_port": null
}
```




### MySQL Size

Returns size of mysql databases.

Default unit is `bytes`, unit can be specified with `--data-urlencode "unit="` - available units are: 'bytes', 'kb', 'mb', 'gb'.

By default, system databases are excluded. To include them: `--data-urlencode "show_all=true"`.

```bash
curl -X GET https://OPENPANEL:2083/api/mysql-size \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

Example response:
```json
[
  {
    "Database": "user_db",
    "Size (MB)": 12.54
  },
  {
    "Database": "another_db",
    "Size (MB)": 48.12
  }
]
```

### Screenshot

If using remote screenshots api, returns url, else if local screenshots are configured, returns image as mimetype 'image/png;.

```bash
curl -X GET https://OPENPANEL:2083/api/screenshot/<path:domain> \
  -H "Authorization: Bearer JWT_TOKEN_HERE"
```

### Search Files

Search website files.

```bash
curl -X GET https://OPENPANEL:2083/api/search/files \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=backup"
```

- set query use: `--data-urlencode "q=" \`
- specify an extension: `--data-urlencode "ext=tar.gz"`
- search in specific folder only: `--data-urlencode "folder=backups" \`

Full example (search query + folder + extension):
```bash
curl -X GET https://OPENPANEL:2083/api/search/files \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=db" \
  --data-urlencode "folder=backups" \
  --data-urlencode "ext=sql"
```

Example response:
```json
[
  {
    "path": "/var/www/html/backups/db_backup_2024.sql",
    "size": 204800,
    "modified": "2024-06-01T10:23:00"
  },
  {
    "path": "/var/www/html/backups/db_notes.sql",
    "size": 102400,
    "modified": "2024-06-02T14:55:00"
  }
]
```


### Search Folders

Search website folders.

```bash
curl -X GET https://OPENPANEL:2083/api/search/folders \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=backup"
```

- set query use: `--data-urlencode "q=" \`
- search in specific folder only: `--data-urlencode "folder=backups" \`

Full example (search query + folder + extension):
```bash
curl -X GET https://OPENPANEL:2083/api/search/folders \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=db" \
  --data-urlencode "folder=backups"
```

Example response:
```json
[
  {
    "path": "db",
    "path": "/var/www/html/backups/"
  },
  {
    "path": "db-dir",
    "path": "/var/www/html/backups/OLD/"
  }
]
```

### Search Features

Search features in OpenPanel .

```bash
curl -X GET https://OPENPANEL:2083/api/search/features \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=backup"
```

- set query use: `--data-urlencode "q=" \`

Example response:
```json
[
  {
    "description": "Manage backups.",
    "link": "/backups",
    "module": "backups",
    "name": "Backups"
  },
  {
    "description": "Change backup settings: schedule, exclude, etc.",
    "link": "/backups/settings",
    "module": "backups",
    "name": "Backup Settings"
  },
  ...
```


### Search Websites

Search websites.

```bash
curl -X GET https://OPENPANEL:2083/api/search/websites \
  -H "Authorization: Bearer JWT_TOKEN_HERE" \
  --data-urlencode "q=pcx3"
```

- set query use: `--data-urlencode "q=" \`

Example response:
```json
[
  [
    "pcx3.com"
  ],
  [
    "pcx3.com/grapejs"
  ]
]
```


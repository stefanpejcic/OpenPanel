# Auto-start Services

Services in OpenPanel start only when they are actually needed, to avoid wasting resources.

## Auto-start Services in OpenAdmin

Upon installing OpenPanel, only the following services are started:

- **OpenAdmin** – For managing the entire server and users.
- **Docker** – Needed for all other containerized services and user accounts.
- **Database** – MySQL database is created and initialized. This database holds Plans, Websites, Domains, and Users.
- **Firewall** – [Sentinel Firewall](https://sentinelfirewall.org/) is installed and started.

Other services are installed and started only when required.

| Service                | Installed | Auto-start                |
|------------------------|-----------|---------------------------|
| OpenAdmin              | ✔       | On installation            |
| Docker                 | ✔       | On installation            |
| Database               | ✔       | On installation            |
| Sentinel Firewall   | ✔       | On installation            |
| OpenPanel              | ✘        | After adding first user account |
| BIND9                  | ✘        | After adding first domain name  |
| Certbot                  | ✘        | After adding first domain name  |
| ClamAV       | ✘        | When enabled by Administrator  |
| Dovecot & Postfix       | ✘        | When enabled by Administrator  |
| FTP                    | ✘        | When enabled by Administrator, after first FTP account is created |

## Auto-start Services in OpenPanel

Similar to OpenAdmin, services in OpenPanel also start only when they are needed. This allows for better resource management.

Services that auto-start for each user:

| Service            | Installed | Auto-start                                         |
|--------------------|-----------|---------------------------------------------------|
| Apache / Nginx / OpenLitespeed     | ✔       | After the user adds the first domain               |       |
| REDIS              | ✘        | After the user activates it           |
| Memcached          | ✘        | After the user activates it           |
| Elasticsearch      | ✘        | After the user activates it           |
| MySQL / MariaDB    | ✔       | After the user adds at least 1 database            |
| Cron               | ✔       | After the user adds at least 1 cron job            |
| PHP versions            | ✘        | After the user sets it for at least 1 domain |

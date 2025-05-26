# Auto-start Services

Services in OpenPanel start only when they are actually needed, to avoid wasting resources.

## Auto-start Services in OpenAdmin

Upon installing OpenPanel, only the following services are started:

- **OpenAdmin** – For managing the entire server and users.
- **Docker** – Needed for all other containerized services and user accounts.
- **Database** – MySQL database is created and initialized. This database holds Plans, Websites, Domains, and Users.
- **Firewall** – CSF is installed and started.

Other services are installed and started only when required.

| Service                | Installed | Auto-start                |
|------------------------|-----------|---------------------------|
| OpenAdmin              | ✔       | On installation            |
| Docker                 | ✔       | On installation            |
| Database               | ✔       | On installation            |
| ConfigServer Firewall   | ✔       | On installation            |
| Uncomplicated Firewall  | ✘        | On installation            |
| OpenPanel              | ✘        | After adding first user account |
| BIND9                  | ✘        | After adding first domain name  |
| Certbot                  | ✘        | After adding first domain name  |
| ClamAV       | ✘        | When enabled by Administrator  |
| Dovecot & Postfix       | ✘        | When enabled by Administrator  |
| FTP                    | ✘        | When enabled by Administrator, after first FTP account is created |

## Auto-start Services in OpenPanel

Using the `/etc/entrypoint.sh` for each user, Administrators define which services to auto-start when their container is started. The container is started after a user is created and on every Docker service restart.

Similar to OpenAdmin, services in OpenPanel also start only when they are needed. This allows for better resource management. An idle account will use around 15MB of RAM, while an account with domains, websites, and MySQL databases may use up to 1GB of RAM.

Services that auto-start for each user:

| Service            | Installed | Auto-start                                         |
|--------------------|-----------|---------------------------------------------------|
| Apache / Nginx     | ✔       | After the user adds the first domain               |       |
| REDIS              | ✘        | After the user installs and activates it           |
| Memcached          | ✘        | After the user installs and activates it           |
| Elasticsearch      | ✘        | After the user installs and activates it           |
| MongoDB            | ✘        | After the user installs and activates it           |
| MySQL / MariaDB    | ✔       | After the user adds at least 1 database            |
| Cron               | ✔       | After the user adds at least 1 cron job            |
| PHP 8.3            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 8.2            | ✔       | After the user adds at least 1 domain              |
| PHP 8.1            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 8.0            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 7.4            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 7.3            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 7.2            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 7.1            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 7.0            | ✘        | After the user installs and sets it for at least 1 domain |
| PHP 5.6            | ✘        | After the user installs and sets it for at least 1 domain |

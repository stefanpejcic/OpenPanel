# Volume Management in OpenPanel

OpenPanel uses **volumes** to persist important user data across container restarts and updates.
Unlike container storage, which is ephemeral, volumes ensure that your databases, website files, and configurations remain intact even if containers are removed or recreated.

This setup provides both **data persistence** and **logical separation of responsibilities**, so each service stores its data in a dedicated, isolated volume.

---

## Available Volumes

OpenPanel provides several pre-configured volumes, each serving a specific purpose:

| Volume Name      | Description                                                 | Physical Location on OS                                             |
| ---------------- | ----------------------------------------------------------- | ------------------------------------------------------------------- |
| `mysql_data`     | Persistent storage for MySQL/MariaDB databases              | `/home/USERNAME/docker-data/volumes/USERNAME_mysql_data/_data/`     |
| `mysql_dumps`    | Storage for database backups in `.sql` format               | `/home/USERNAME/docker-data/volumes/USERNAME_mysql_dumps/_data/`    |
| `html_data`      | Persistent storage for website files under `/var/www/html/` | `/home/USERNAME/docker-data/volumes/USERNAME_html_data/_data/`      |
| `mail_data`      | Persistent storage for mailboxes under `/var/mail/`         | `/home/USERNAME/docker-data/volumes/USERNAME_mail_data/_data/`      |
| `webserver_data` | Storage for webserver vhost configurations                  | `/home/USERNAME/docker-data/volumes/USERNAME_webserver_data/_data/` |
| `pg_data`        | Persistent storage for PostgreSQL databases                 | `/home/USERNAME/docker-data/volumes/USERNAME_pg_data/_data/`        |

---

## `mysql_data` Volume

The **mysql_data** volume is where MySQL and MariaDB databases are stored.
It ensures that database data persists even if the database container is restarted or rebuilt.

* Path inside container: `/var/lib/mysql`
* Purpose: **Database persistence**

---

## `mysql_dumps` Volume

The **mysql_dumps** volume holds backup `.sql` dumps generated during backup operations.
It is separate from the main database volume to keep live data and backups isolated.

* Typical contents: `.sql` dump files
* Purpose: **Backup storage**

---

## `html_data` Volume

The **html_data** volume maps to the `/var/www/html/` directory used by web servers.
This is where website source files (PHP, HTML, JS, assets) are stored.

* Path inside container: `/var/www/html/`
* Purpose: **Website storage**

---

## `mail_data` Volume

The **mail_data** volume stores all email data, including user mailboxes.
This ensures mail persists independently of the mail service container lifecycle.

* Path inside container: `/var/mail/`
* Purpose: **Mail persistence**

---

## `webserver_data` Volume

The **webserver_data** volume is used to store configuration files for webservers such as Nginx or Apache.
This allows vhost configurations to persist across restarts and be easily modified.

* Typical contents: Virtual host configs, SSL configs
* Purpose: **Webserver configuration storage**

---

## `pg_data` Volume

The **pg_data** volume holds PostgreSQL database files.
Like `mysql_data`, this ensures database persistence and durability.

* Path inside container: `/var/lib/postgresql/data`
* Purpose: **Database persistence**

---

✅ With these volumes, OpenPanel cleanly separates **data**, **backups**, **configurations**, and **service-specific files**, ensuring both persistence and maintainability.

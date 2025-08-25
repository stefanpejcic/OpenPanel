# MySQL, MariaDB, or Percona

When creating a user, you can select which MySQL-compatible database type to use:

* **MySQL**
* **MariaDB**

Users also have the *Change MySQL Type* option, which allows switching between these types after creation.

Administrators can configure the default database type for new users from [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/).

---

## MySQL

MySQL can be assigned either to all new users by default or to individual users during creation.

### Set for a Single User

To set MySQL for a single user during creation:

1. Open the **New User** form.
2. Under the **Advanced** section, select **MySQL** as the database type.

![mysql for user](/img/guides/mysql.png)

### Set as Default

To make MySQL the default for all newly created users:

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **MySQL** as the default database type.

![mysql as default](/img/guides/mysql_default.png)

---

## MariaDB

MariaDB can be assigned either to all new users by default or to individual users during creation.

### Set for a Single User

To set MariaDB for a single user during creation:

1. Open the **New User** form.
2. Under the **Advanced** section, select **MariaDB** as the database type.

![mariadb for user](/img/guides/mariadb.png)


### Set as Default

To make MariaDB the default for all newly created users:

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **MariaDB** as the default database type.

![mariadb as default](/img/guides/mariadb_default.png)

---

## Percona

Percona is a drop-in replacement for MySQL. It can be set by changing the Docker image from `mysql` to `percona`. The interface and all actions remain fully compatible.

### Set for a Single User

To set Percona for a single user during creation:

1. Open the **New User** form.
2. Under the **Advanced** section, select **MySQL** as the database type.
3. Create the user.
4. Edit the user's Compose file and change the MySQL image to `percona`.

![percona for user](/img/guides/percona.png)

### Set as Default

To make Percona the default for all newly created users:

* Go to [**OpenAdmin > Settings > User Defaults > Advanced**](/docs/admin/settings/defaults/)
* Change the MySQL service image to `percona`.

![percona as default](/img/guides/percona_default.png)


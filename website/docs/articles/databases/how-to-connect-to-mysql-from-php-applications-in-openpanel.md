---
sidebar_label: "Connecting to MySQL Server from Applications in OpenPanel"
---

# How to Connect to MySQL from PHP Applications (Hostname, Port, Credentials)

OpenPanel runs each user service inside its own container. The MySQL/MariaDB container shares its Unix socket with the PHP containers (PHP-FPM and OpenLiteSpeed) of the same user, so PHP applications can connect to the database on `localhost` just like on a regular server. Other applications (Node.js, Python, Ruby, Java...) run in their own containers and connect through the container hostname instead.

For more information about networks: [Network Isolation in OpenPanel](/docs/articles/containers/network-isolation-openpanel/)

| Application | Hostname | Port |
|---|---|---|
| **PHP** (WordPress, Joomla, Laravel, custom PHP...) | `localhost` | not needed, it uses the Unix socket |
| **Node.js, Python, Ruby, Java** and other non-PHP apps | `mysql` (for MySQL) or `mariadb` (for MariaDB) | `3306` |

**Username/Password:** created for the database from OpenPanel UI.

:::info
`mysql` / `mariadb` as the hostname keeps working for PHP apps too, so existing sites don't need to be changed.

On accounts created before the Unix socket was added, `localhost` is not available in PHP yet, use `mysql` or `mariadb` there. Websites installed from OpenPanel pick the right host for the account automatically.
:::

⚠️ **Important:** Never use `127.0.0.1` as the database host. It connects over TCP inside the PHP container itself, where no database is listening.

---

## Example: WordPress

When setting up WordPress inside OpenPanel, use the following database configuration in `wp-config.php`:

```php
define( 'DB_HOST', 'localhost' );
```

WordPress connects to the database through the Unix socket shared with the PHP container. On older accounts use `mysql` or `mariadb` instead.

---

## Example: Custom PHP Application

In a custom PHP app, you can connect like this:

```php
<?php
$host = "localhost";   // or "mysql" / "mariadb"
$db   = "my_database";
$user = "my_user";
$pass = "my_password";

$dsn = "mysql:host=$host;dbname=$db;charset=utf8mb4";

try {
    $pdo = new PDO($dsn, $user, $pass);
    echo "Connected successfully!";
} catch (PDOException $e) {
    echo "Connection failed: " . $e->getMessage();
}
```

With `localhost`, PHP connects through the `/var/run/mysqld/mysqld.sock` socket, which is set as the default in `php.ini` (`mysqli.default_socket` and `pdo_mysql.default_socket`).

---

## Example: Node.js Application

For Node.js using `mysql2` or `sequelize`:

```js
const mysql = require('mysql2');

const connection = mysql.createConnection({
  host: 'mysql',       // or "mariadb"
  user: 'my_user',
  password: 'my_password',
  database: 'my_database'
});

connection.connect(err => {
  if (err) {
    console.error('Connection error:', err);
    return;
  }
  console.log('Connected successfully!');
});
```

---

## Example: Python Application

Using `mysql-connector-python`:

```python
import mysql.connector

conn = mysql.connector.connect(
    host="mysql",        # or "mariadb"
    user="my_user",
    password="my_password",
    database="my_database"
)

cursor = conn.cursor()
cursor.execute("SELECT DATABASE();")
print("Connected to:", cursor.fetchone())
```

---

✅ **Summary:**

* PHP apps: use `localhost` (or `mysql` / `mariadb` on older accounts).
* Node.js, Python and other non-PHP apps: use `mysql` or `mariadb` on port `3306`, they reach the database through [the `db` network](/docs/articles/containers/network-isolation-openpanel/).
* Never use `127.0.0.1`.

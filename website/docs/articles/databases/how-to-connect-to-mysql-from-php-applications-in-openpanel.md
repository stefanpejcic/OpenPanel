# Connecting to MySQL Server from Applications in OpenPanel

OpenPanel runs each user service inside its own container and uses local networks to isolate them. This means that applications do not connect to the database via `localhost` or `127.0.0.1`, but instead through container hostnames.

For more information about networks: [Network Isolation in OpenPanel](/docs/articles/docker/network-isolation-openpanel/)

To connect from your **PHP**, **Node.js**, or **Python** application to a MySQL/MariaDB database, you must use:

* **Hostname:** `mysql` (for MySQL) or `mariadb` (for MariaDB)
* **Port:** `3306` (default)
* **Username/Password:** created for the database from OpenPanel UI

⚠️ **Important:** Never use `localhost` or `127.0.0.1` as the database host. These will not work because containers are isolated.

---

## Example: WordPress

When setting up WordPress inside OpenPanel, use the following database configuration:

```ini
DB_HOST=mysql
```

Here, `DB_HOST` is set to `mysql`, so WordPress can reach the MySQL container. If you are using MariaDB, then repalce `mysql` with `mariadb`.

---

## Example: Custom PHP Application

In a custom PHP app, you can connect like this:

```php
<?php
$host = "mysql";       // or "mariadb"
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

* Use `mysql` or `mariadb` as the hostname.
* Never use `localhost` or `127.0.0.1`.
* Works across PHP, WordPress, Node.js, Python, and any other app in [the `db` network](/docs/articles/docker/network-isolation-openpanel/).

# Csatlakozás a MySQL Serverhez az OpenPanel alkalmazásaiból

Az OpenPanel minden felhasználói szolgáltatást saját tárolójában futtat, és helyi hálózatokat használ az elkülönítésre. Ez azt jelenti, hogy az alkalmazások nem a "localhost" vagy a "127.0.0.1" segítségével csatlakoznak az adatbázishoz, hanem tároló gazdagépneveken keresztül.

További információ a hálózatokról: [Network Isolation in OpenPanel](/docs/articles/docker/network-isolation-openpanel/)

A **PHP**, **Node.js** vagy **Python** alkalmazásból MySQL/MariaDB adatbázishoz való csatlakozáshoz a következőket kell használnia:

* **Gazdagépnév:** `mysql` (MySQL-hez) vagy `mariadb` (MariaDB-hez)
* **Port:** `3306` (alapértelmezett)
* **Felhasználónév/Jelszó:** az adatbázishoz az OpenPanel UI-ból készült

⚠️ **Fontos:** Soha ne használja a "localhost" vagy a "127.0.0.1" értéket adatbázis-gazdaként. Ezek nem működnek, mert a konténerek el vannak szigetelve.

---

## Példa: WordPress

A WordPress OpenPanelen belüli beállításakor használja a következő adatbázis-konfigurációt:

```ini
DB_HOST=mysql
```

Itt a `DB_HOST` `mysql-re van állítva, így a WordPress elérheti a MySQL-tárolót. Ha MariaDB-t használ, cserélje ki a "mysql"-t a "mariadb"-re.

---

## Példa: Egyéni PHP alkalmazás

Egy egyéni PHP alkalmazásban a következőképpen csatlakozhat:

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

## Példa: Node.js alkalmazás

A „mysql2” vagy „sequelize” parancsot használó Node.js esetén:

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

## Példa: Python alkalmazás

A "mysql-connector-python" használata:

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

✅ **Összefoglaló:**

* Használja a "mysql" vagy a "mariadb" nevet gazdagépnévként.
* Soha ne használja a "localhost" vagy a "127.0.0.1" kódot.
* Működik PHP, WordPress, Node.js, Python és bármely más alkalmazásban [a `db` hálózaton] (/docs/articles/docker/network-isolation-openpanel/).

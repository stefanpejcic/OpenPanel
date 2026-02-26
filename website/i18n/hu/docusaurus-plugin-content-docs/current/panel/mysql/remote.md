---
sidebar_position: 6
---

# Távoli MySQL

![databases_remote_mysql.png](/img/panel/v2/databases_remotedis.png)

A távoli MySQL-hozzáférés lehetővé teszi, hogy egy másik (távoli) eszközről vagy helyről az interneten keresztül csatlakozzon egy MySQL-adatbázishoz ezen a kiszolgálón.

A távoli MySQL hozzáférés engedélyezése megnyitja az adatbázist a teljes internet kapcsolatai számára, ami biztonsági kockázatot jelenthet. Kérjük, vegye figyelembe a következőket:

- Biztonsági sebezhetőségek: Az internetről való hozzáférés engedélyezése potenciális biztonsági réseknek teheti ki az adatbázist, ami növeli a jogosulatlan hozzáférés, az adatszivárgás és az adatvesztés kockázatát.
- Adatvédelem: Az Ön érzékeny adatai veszélybe kerülhetnek, ha nincs megfelelően biztosítva. Ügyeljen arra, hogy erős jelszavakat és titkosítást használjon adatai védelme érdekében.
- Rendszeres biztonsági mentések: Győződjön meg arról, hogy rendszeres adatbázis-mentésekkel rendelkezik az adatok helyreállításához biztonsági incidensek esetén.
- Mielőtt engedélyezné a távoli MySQL-hozzáférést, tekintse át biztonsági beállításait, és alaposan fontolja meg a lehetséges kockázatokat.

Ha nem biztos a biztonsági vonatkozásban, vagy segítségre van szüksége, forduljon a tárhely rendszergazdájához.

---

## Távoli MySQL hozzáférés engedélyezése

A távoli MySQL hozzáférés alapértelmezés szerint le van tiltva.

Az adatbázisokhoz való távoli hozzáférés engedélyezéséhez kattintson a "Távoli MySQL-elérés engedélyezése" gombra az "Adatbázisok > Távoli MySQL" oldalon.

Miután engedélyezte, kap egy IP-címet és egy portot, amelyek segítségével biztonságos kapcsolatot létesíthet a MySQL szolgáltatással egy távoli szerverről.

![databases_remote_mysql_enabled.png](/img/panel/v2/databases_remoten.png)

:::info
A megjelenített port véletlenszerűen jön létre, és egyedi a MySQL-példányhoz. Kerülje a 3306-os szabványos port használatát távoli hozzáféréshez, mert az nem fog működni.
:::

---

## Távoli MySQL hozzáférés tesztelése

A távoli MySQL-elérés teszteléséhez egy másik szerverről használhatja a "mysql" parancssori klienst vagy egy adatbázis-kezelő eszközt.

### mysql parancssori kliens

Győződjön meg arról, hogy a MySQL-ügyfél telepítve van azon a kiszolgálón, amelyről hozzá kíván férni a MySQL-adatbázishoz.

A következő paranccsal csatlakozhat a MySQL szerverhez a távoli kiszolgálóról:
```bash
mysql -h <IP_ADDRESS> -P <PORT> -u <USERNAME> -p <DATABASE_NAME>
```

### PHP

Ha PHP-alapú kapcsolatot szeretne létrehozni távoli szerverről egy OpenPanel-en tárolt MySQL-példányhoz, íme egy példa kódrészlet:

```php
<?php
// MySQL server details
$hostname = 'IP_ADDRESS_HERE';
$port = 'CUSTOM_PORT'; // Custom port from the OpenPanel interface
$username = 'YOUR_MYSQL_USERNAME';
$password = 'YOUR_MYSQL_PASSWORD';
$database = 'YOUR_MYSQL_DATABASE';

// Create a connection
$conn = new mysqli($hostname, $username, $password, $database, $port);

// Check the connection
if ($conn->connect_error) {
   die("Connection failed: " . $conn->connect_error);
}

echo "Connected to the MySQL server on $hostname:$port successfully.";

// Close the connection
$conn->close();
?>
```

### WordPress

Ha távoli WordPress-telepítésből szeretne csatlakozni egy MySQL-példányhoz az OpenPanel-en, szerkesztenie kell a wp-config.php fájlt, és módosítania kell az adatbázis-kapcsolati információkat.

```php
// ** MySQL settings - You can get this info from your web host **
/** The name of the database for WordPress */
define('DB_NAME', 'YOUR_MYSQL_DATABASE');

/** MySQL database username */
define('DB_USER', 'YOUR_MYSQL_USERNAME');

/** MySQL database password */
define('DB_PASSWORD', 'YOUR_MYSQL_PASSWORD');

/** MySQL hostname with custom port */
define('DB_HOST', 'IP_ADDRESS:PORT');
```

---

## A távoli MySQL hozzáférés letiltása

Ha le szeretné tiltani a hozzáférést, egyszerűen kattintson a "Távoli MySQL-elérés letiltása" gombra, és azonnal deaktiválja a távoli hozzáférést a MySQL konfigurációjában. Kérjük, vegye figyelembe, hogy ehhez a művelethez a MySQL szolgáltatás újraindítása is szükségessé válik az új beállítás alkalmazásához.

![databases_remote_mysql_disabled.png](/img/panel/v2/databases_remotedis.png)

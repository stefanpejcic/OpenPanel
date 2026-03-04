# MySQL, MariaDB vagy Percona

Felhasználó létrehozásakor kiválaszthatja, hogy melyik MySQL-kompatibilis adatbázistípust használja:

* **MySQL**
**MariaDB**

A felhasználók a *Change MySQL Type* opcióval is rendelkeznek, amely lehetővé teszi a váltást ezek között a típusok között a létrehozás után.

A rendszergazdák az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések**](/docs/admin/settings/defaults/) oldalon konfigurálhatják az alapértelmezett adatbázistípust az új felhasználók számára.

---

## MySQL

A [MySQL](https://www.mysql.com/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a létrehozás során.

### Beállítás egy felhasználó számára

A MySQL beállítása egyetlen felhasználó számára a létrehozás során:

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza ki a **MySQL** lehetőséget adatbázistípusként.

![mysql for user](/img/guides/mysql.png)

### Beállítás alapértelmezettként

A MySQL alapértelmezetté tétele minden újonnan létrehozott felhasználó számára:

* Nyissa meg az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések**] (/docs/admin/settings/defaults/) menüpontot, és válassza ki a **MySQL**-t alapértelmezett adatbázistípusként.

![mysql alapértelmezettként](/img/guides/mysql_default.png)

---

## MariaDB

A [MariaDB](https://mariadb.org/) alapértelmezés szerint hozzárendelhető az összes új felhasználóhoz, vagy az egyes felhasználókhoz a létrehozás során.

### Beállítás egy felhasználó számára

A MariaDB beállítása egyetlen felhasználóhoz a létrehozás során:

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza ki a **MariaDB** adatbázistípust.

![mariadb for user](/img/guides/mariadb.png)


### Beállítás alapértelmezettként

Ha a MariaDB-t szeretné alapértelmezettként használni az összes újonnan létrehozott felhasználó számára:

* Nyissa meg az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések**] (/docs/admin/settings/defaults/) menüpontot, és válassza ki a **MariaDB**-t alapértelmezett adatbázistípusként.

![mariadb alapértelmezettként](/img/guides/mariadb_default.png)

---

## Percona

A [Percona](https://www.percona.com/mysql/software) a MySQL helyettesítője. A Docker-kép `mysql`-ről `percona`-ra történő módosításával állítható be. A felület és az összes művelet teljes mértékben kompatibilis marad.

### Beállítás egy felhasználó számára

A Percona beállítása egyetlen felhasználóhoz a létrehozás során:

1. Nyissa meg az **Új felhasználó** űrlapot.
2. A **Speciális** részben válassza ki a **MySQL** lehetőséget adatbázistípusként.
3. Hozza létre a felhasználót.
4. Szerkessze a felhasználó Compose fájlját, és módosítsa a MySQL-képet "percona"-ra.

![percona for user](/img/guides/percona.png)

### Beállítás alapértelmezettként

A Percona alapértelmezett beállítása az összes újonnan létrehozott felhasználó számára:

* Lépjen az [**OpenAdmin > Beállítások > Felhasználói alapértelmezések > Speciális**] oldalra (/docs/admin/settings/defaults/)
* Módosítsa a MySQL szolgáltatás képét "percona"-ra.

![percona alapértelmezettként](/img/guides/percona_default.png)


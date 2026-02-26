# Adatbázis importálása

A táblákat a következő módszerek egyikével importálhatja egy MySQL/MariaDB adatbázisba:

* **Adatbázis importálása** funkció (OpenPanelen keresztül)
* **phpMyAdmin** felület
* **Terminal** funkció (OpenPanelen keresztül)

---

## 1. Adatbázis importálása (OpenPanelen keresztül)

Az **Adatbázis importálása** funkciót engedélyezni kell a tervben, hogy a lehetőség megjelenjen az OpenPanel irányítópultján.
Ha elérhető, kövesse az alábbi útmutatót: [Adatbázis importálási dokumentáció](/docs/panel/mysql/import/)

---

## 2. A phpMyAdmin használata

Ha a **phpMyAdmin** funkció engedélyezve van:

* Nyissa meg a phpMyAdmin programot az irányítópulton.
* Válassza ki a céladatbázist.
* Használja az **Importálás** lapot az `.sql` fájl feltöltéséhez és végrehajtásához.

Alternatív megoldásként feltöltheti az SQL fájlt a **Fájlkezelő** segítségével a `/var/www/html` könyvtárba. A feltöltött fájlok ezután láthatók lesznek a phpMyAdmin importálási oldalán.

Dokumentáció: [phpMyAdmin importálási útmutató](https://openpanel.com/docs/panel/mysql/phpmyadmin/)

---

## 3. A terminál használata (Docker)

A terminálhozzáférés használatához:

* A **Docker** funkciót engedélyezni kell a tervben, hogy a tárolóhoz való hozzáférés látható legyen az OpenPanelben.
* Ha engedélyezve van, válassza ki a MySQL/MariaDB tárolót az irányítópulton.
* Nyissa meg a terminált, és futtassa a következő parancsokat az SQL-fájl importálásához:

```bash
mysql

USE your_database_name;

wget https://example.com/some_file.sql

SOURCE ./your_file.sql;
```

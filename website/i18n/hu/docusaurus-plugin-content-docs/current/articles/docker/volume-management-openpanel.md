# Kötetkezelés az OpenPanelben

Az OpenPanel **köteteket** használ a fontos felhasználói adatok megőrzésére a tároló újraindítása és frissítése során.
A konténertárolástól eltérően, amely átmeneti jellegű, a kötetek biztosítják, hogy az adatbázisok, a webhelyfájlok és a konfigurációk érintetlenek maradjanak még a tárolók eltávolítása vagy újbóli létrehozása esetén is.

Ez a beállítás biztosítja az **adatmegmaradást** és a **a felelősségek logikai szétválasztását** is, így minden szolgáltatás egy dedikált, elkülönített kötetben tárolja adatait.

---

## Elérhető kötetek

Az OpenPanel számos előre konfigurált kötetet kínál, amelyek mindegyike meghatározott célt szolgál:

| Kötet neve | Leírás | Fizikai hely az operációs rendszeren |
| ----------------- | ------------------------------------------------------------ | -------------------------------------------------------------------- |
| `mysql_data` | Állandó tárolás a MySQL/MariaDB adatbázisokhoz | `/home/FELHASZNÁLÓNÉV/docker-data/volumes/USERNAME_mysql_data/_data/` |
| `mysql_dumps` | Tárhely adatbázis-mentésekhez `.sql` formátumban | `/home/USERNAME/docker-data/volumes/USERNAME_mysql_dumps/_data/` |
| `html_data` | A webhelyfájlok állandó tárolása a `/var/www/html/` | `/home/USERNAME/docker-data/volumes/USERNAME_html_data/_data/` |
| `mail_data` | Állandó tárhely a postafiókokhoz a `/var/mail/` | alatt `/home/FELHASZNÁLÓNÉV/docker-data/volumes/USERNAME_mail_data/_data/` |
| `webszerver_adatai` | Tárhely a webszerver vhost konfigurációihoz | `/home/FELHASZNÁLÓNÉV/docker-data/volumes/USERNAME_webserver_data/_data/` |
| `pg_data` | Állandó tárolás a PostgreSQL adatbázisokhoz | `/home/USERNAME/docker-data/volumes/USERNAME_pg_data/_data/` |

---

## `mysql_data` kötet

A **mysql_data** kötet az, ahol a MySQL és a MariaDB adatbázisok tárolódnak.
Gondoskodik arról, hogy az adatbázisadatok akkor is megmaradjanak, ha az adatbázis-tárolót újraindítják vagy újraépítik.

* Útvonal a tárolón belül: `/var/lib/mysql`
* Cél: **Adatbázis fennmaradása**

---

## `mysql_dumps` kötet

A **mysql_dumps** kötet a biztonsági mentési műveletek során létrehozott `.sql` biztonsági másolatokat tartalmazza.
Elkülönül a fő adatbázis-kötettől, hogy az élő adatokat és a biztonsági másolatokat elkülönítve tartsa.

* Tipikus tartalom: `.sql` dump fájlok
* Cél: **Biztonsági mentés**

---

## `html_data` Kötet

A **html_data** kötet a webszerverek által használt `/var/www/html/` könyvtárra van leképezve.
Itt tárolódnak a webhely forrásfájljai (PHP, HTML, JS, eszközök).

* Útvonal a tárolón belül: `/var/www/html/`
* Cél: **Webhelytárhely**

---

## `mail_data` Kötet

A **mail_data** kötet az összes e-mail adatot tárolja, beleértve a felhasználói postafiókokat is.
Ez biztosítja, hogy a levelek a levelezési szolgáltatás tárolójának életciklusától függetlenül fennmaradjanak.

* Útvonal a tárolón belül: `/var/mail/`
* Cél: **levélmegmaradás**

---

## `webserver_data` Kötet

A **webserver_data** kötet a webszerverek, például az Nginx vagy az Apache konfigurációs fájlok tárolására szolgál.
Ez lehetővé teszi, hogy a vhost konfigurációk az újraindítások során is fennmaradjanak, és könnyen módosíthatók.

* Tipikus tartalom: Virtuális gazdagép konfigurációk, SSL konfigurációk
* Cél: **Webszerver konfigurációs tárhely**

---

## `pg_data` Kötet

A **pg_data** kötet PostgreSQL adatbázisfájlokat tartalmaz.
A "mysql_data"-hoz hasonlóan ez is biztosítja az adatbázis állandóságát és tartósságát.

* Útvonal a tárolóban: `/var/lib/postgresql/data`
* Cél: **Adatbázis fennmaradása**

---

✅ Ezekkel a kötetekkel az OpenPanel tisztán szétválasztja az **adatokat**, **biztonsági mentéseket**, **konfigurációkat** és **szolgáltatásspecifikus fájlokat**, biztosítva a tartósságot és a karbantarthatóságot.

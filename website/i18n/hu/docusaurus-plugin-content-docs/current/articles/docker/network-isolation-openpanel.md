# Hálózati elkülönítés az OpenPanelben

Az OpenPanel konténereket használ az egyes felhasználók szolgáltatásainak elkülönítésére. Ezeket a konténereket ezután helyi hálózatokká különítik el, további szigetelési réteget biztosítva.

Ez a beállítás biztosítja, hogy az azonos hálózaton belüli konténerek biztonságosan kommunikáljanak anélkül, hogy a portokat kitennék a nyilvános internetnek.

## Elérhető hálózatok

A tárolók három hálózat egyikéhez rendelhetők:

* **`db`** – konténerek, amelyeknek adatbázis-hozzáférésre van szükségük
* **`www`** – olyan tárolók, amelyekhez webszerver- vagy gyorsítótár-hozzáférésre van szükség
* **`nincs hálózat`** – olyan konténerek, amelyek nem férnek hozzá más tárolókhoz

Az ugyanabban a hálózatban lévő tárolók közötti kommunikáció a konténernevek, mint gazdagépnevek használatával történhet. Például a PHP-konténerek egyszerűen kapcsolódhatnak a MySQL-hez a "mysql" vagy a "mariadb" gazdagépnév használatával.

---

## `db` Hálózat

A **db** hálózatot az adatbázisokhoz való hozzáférést igénylő szolgáltatásokhoz tervezték.

Alapértelmezés szerint ez a hálózat a következőket tartalmazza:

* MySQL
* MariaDB
* phpMyAdmin
* PostgreSQL
* PgAdmin
* Tor
* PHP-FPM szolgáltatások

Ha olyan egyéni szolgáltatást (tárolót) ad hozzá, amelynek kapcsolódnia kell egy adatbázishoz, csatolja azt a „db” hálózathoz.

---

## `www` Hálózat

A **www** hálózatot olyan szolgáltatásokhoz használják, amelyek hozzáférést igényelnek a webszerverekhez, gyorsítótárazáshoz vagy keresőmotorokhoz.

Alapértelmezés szerint ez a hálózat a következőket tartalmazza:

* OpenLiteSpeed
* Apache
* Nginx
* Lakk
* Tor
* Redis
* Memcached
* Elaszticsearch
* OpenSearch
* PHP-FPM szolgáltatások

Ha olyan egyéni szolgáltatást (tárolót) ad hozzá, amelynek kommunikálnia kell egy webszerverrel, Redis-szel vagy Memcached-el, csatolnia kell a „www” hálózathoz.

---

## `nincs hálózat`

A hálózat nélküli tárolók teljesen elszigeteltek, és nem tudnak kommunikálni más tárolókkal.

Ezek a tárolók általában csak a terminálon vagy az OpenPanel UI-n keresztül érhetők el. Ideálisak olyan szolgáltatásokhoz, amelyek nem igényelnek interakciót adatbázisokkal, webszerverekkel vagy más tárolókkal.

# Felhasználói tárolók kezelése a terminálról

Az OpenPanelben minden felhasználónak megvan a saját [Docker-környezete](https://docs.docker.com/engine/manage-resources/contexts/), amely [rootless módban] (https://docs.docker.com/engine/security/rootless/) fut. Ha ismeri a Dockert, közvetlenül a terminálról kezelheti a felhasználó fájljait és szolgáltatásait.

A Docker-környezetek közötti váltáshoz SSH-hozzáférés szükséges a kiszolgálón **root felhasználóként**.

---

## Beállítás létrehozása

Minden felhasználó környezete a következőkön alapul:

* **`docker-compose.yml`** – szolgáltatásokat, hálózatokat és köteteket határoz meg.
* **`.env`** – CPU/memóriakorlátokat, Docker-képcímkéket és egyéni portokat határoz meg.

> ⚠️ Ezeket a fájlokat mindig az OpenAdmin felületen keresztül szerkessze. A felület érvényesíti a változtatásokat. A közvetlenül a terminálról történő szerkesztés megkerüli az érvényesítést, és megszakíthatja a szolgáltatásokat. Bármilyen módosítás előtt készítsen biztonsági másolatot.

* [Kötetkezelés az OpenPanelben](/docs/articles/docker/volume-management-openpanel/)
* [Network Isolation in OpenPanel](/docs/articles/docker/network-isolation-openpanel/)

---

## Fájlszerkezet

Minden felhasználó saját könyvtára a következő címen található:

```
/home/USERNAME/
```

Konfigurációs fájlokat és felhasználói adatokat tartalmaz. Példa szerkezetre:

```
├── .env                 # Environment variables: versions, ports, CPU/memory limits
├── backup.env           # Backup configuration
├── crons.ini            # Cron jobs
├── custom.cnf           # MySQL configuration
├── default.vcl          # Varnish cache configuration
├── docker-compose.yml   # Service definitions
├── docker-data          # Docker storage
│   ├── containerd       # Container runtime data
│   ├── containers       # Individual container data
│   ├── image            # Docker images
│   ├── network          # Docker networks
│   ├── overlay2         # Overlay filesystem storage
│   ├── plugins          # Docker plugins
│   └── volumes          # Persistent data volumes
│       ├── <USERNAME>_html_data   # Website files (/var/www/html/)
│       └── <USERNAME>_mysql_data  # MySQL database files (/var/lib/mysql/)
├── httpd.conf           # Apache configuration
├── my.cnf               # MySQL root credentials for service commands
├── nginx.conf           # NGINX configuration
├── openlitespeed.conf   # OpenLiteSpeed configuration
├── openresty.conf       # OpenResty configuration
├── php.ini/             # PHP configuration files
├── pma.php              # phpMyAdmin entry point
└── sockets              # Sockets (MySQL, Redis, PostgreSQL, etc.)
```

* A webhelyfájlok a Docker `USERNAME_html_data' kötetében tárolódnak, és fizikailag a következő helyen találhatók:

```
/home/USERNAME/docker_data/volumes/USERNAME_html_data/_data/
```

* A fájl tulajdonjogának meg kell egyeznie az OpenPanel felhasználói azonosítójával. Ellenőrizze a következővel:

```bash
id -u USERNAME
```

* A Docker **rootless módban** fut, és a felhasználói UID-t a tárolókon belüli "root"-hoz rendeli hozzá. Ezzel elkerülhető az engedélyekkel kapcsolatos problémák, miközben a felhasználó gyökér szintű jogosultságokat ad a szolgáltatásaikban.

* A lemezkvótákat a [quota](https://linux.die.net/man/2/quotactl) eszközzel kényszerítik ki. A felhasználó lemez- és inodes-használatának ellenőrzése:

```bash
quota -u USERNAME
```

---

## Szolgáltatások kezelése

A szolgáltatások a `/home/FELHASZNÁLÓNÉV/docker-compose.yml` fájlban vannak definiálva. Ezt a fájlt egyéni szolgáltatásokkal bővítheti, de az OpenPanel felhasználói felülettel való kompatibilitás megőrzése érdekében [kövesse az irányelveket](#).

Használja a felhasználó Docker-környezetét a szolgáltatások kezeléséhez. Példák:

* **Futó szolgáltatások listája:**

```bash
docker --context=USERNAME ps -a
```

**Szolgáltatás leállítása:**

```bash
docker --context=USERNAME stop mysql
```

**Szolgáltatás újraindítása:**

```bash
cd /home/USERNAME && \
docker --context=USERNAME compose down mysql && \
docker --context=USERNAME compose up -d mysql
```

* **Végrehajtás szolgáltatásba:**

```bash
docker --context=USERNAME exec -it mysql bash
```

* **Erőforráshasználat megtekintése:**

```bash
docker --context=USERNAME stats --no-stream
```

---

## Domainek

A tartománykonfigurációs fájlok a következő helyen tárolódnak:

* **Caddy konfiguráció:**
```
/etc/openpanel/caddy/domains/<DOMAIN>.conf
```

Minden domainnek saját Caddyfile-ja van. A Caddy kezeli az SSL-tanúsítványokat, és fordított proxyként működik a felhasználó webszerveréhez.
További információ: [Hogyan áramlik a webes forgalom felhasználói tárolókkal](https://openpanel.com/docs/articles/docker/how-traffic-flows-in-openpanel/)


* **BIND9 zónafájl:**
```
/etc/bind/zones/<DOMAIN>.db
```
Tartalmazza a tartomány DNS-rekordjait.

---

## Biztonsági mentések

Az OpenPanel az [offen/docker-volume-backup](https://offen.github.io/docker-volume-backup/) szolgáltatást használja a biztonsági mentésekhez. A konfiguráció a `/home/{CONTEXT}/backup.env` helyen található.

* A rendszergazdák ütemezhetik az automatikus biztonsági mentést a felhasználók számára.
* A felhasználók saját biztonsági mentéseiket is engedélyezhetik és kezelhetik, ha ez engedélyezett.
* Csak a Docker-kötetek biztonsági másolata készül, amelyek a tényleges felhasználói adatokat (webhelyfájlokat és adatbázisokat) tartalmazzák.

A részletes utasításokért lásd: [Configuring OpenPanel Backups](/docs/articles/backups/configuring-backups)

---

## Crons

Az OpenPanel a [mcuadros/ofelia](https://github.com/mcuadros/ofelia) függvényt használja a cron obs-hoz. A konfiguráció a `/home/{CONTEXT}/crons.ini` helyen található.

---

## Portok

A portok a felhasználó `.env` fájljában vannak meghatározva:

```bash
root@stefan:/# grep _PORT /home/stefan/.env 
HTTP_PORT="127.0.0.1:32780:80"
HTTPS_PORT="127.0.0.1:32781:443"
PROXY_HTTP_PORT="127.0.0.1:32782:80" 
MYSQL_PORT="127.0.0.1:32777:3306"
PMA_PORT="32779:80"
POSTGRES_PORT="0:5432"
PGADMIN_PORT="0:80"
```

* A portok a felhasználó létrehozása során jönnek létre, de az adminisztrátor módosíthatja azokat. A változtatások után futtassa a `compose down && Compose up` parancsot, hogy a változtatások érvénybe lépjenek.
* ⚠️ A Docker gyökér nélküli üzemmódja megakadályozza a **1024** alatti portok használatát.
* A szolgáltatások nem tehetik közzé a portokat kívülről, kivéve, ha távoli hozzáférésre van szükség (pl. phpMyAdmin, távoli MySQL vagy pgAdmin).

---

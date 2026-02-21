# Hiba az adatbázis-kapcsolat létrehozása során

Ha az OpenAdmin felületen a **„Hiba az adatbázis-kapcsolat létrehozásakor”** hibát látja, az azt jelzi, hogy az adatbázis-kapcsolat meghiúsult.

![admin error mysql](https://i.postimg.cc/tJMCcCx1/mysql-error-on-admin.png)

---

## MySQL szolgáltatás

Az OpenPanel és az OpenAdmin a `/root/docker-compose.yml` fájlban definiált MySQL Docker szolgáltatást használ.

Először győződjön meg arról, hogy a szolgáltatás fut:

**Az OpenAdmintól:**

1. Nyissa meg a **Szolgáltatások > Szolgáltatások állapota** menüpontot.
2. Ellenőrizze a MySQL szolgáltatás állapotát.
3. Ha szükséges, indítsa újra erről az oldalról.

**A terminálról:**

```bash
docker ps -a
```

Keresse meg az `openpanel_mysql` szolgáltatást a kimenetben.

---

## A MySQL nem indul el

Ha a MySQL szolgáltatás nem indul el, a Docker folyamatosan újraindítja. Ezt a `docker ps -a` kimenetéből figyelhetjük meg.

Példa:
```bash
root@openpanel:~# docker ps -a
CONTAINER ID   IMAGE                COMMAND                  CREATED          STATUS                          PORTS     NAMES
6d9885164cba   mysql/mysql-server   "/entrypoint.sh mysql…"   30 minutes ago   Restarting (1) 22 seconds ago             openpanel_mysql
```


* Ha az üzemidő csak néhány másodperc, és az állapot **nem fut** vagy **nem működik**, ellenőrizze a szolgáltatásnaplókat:

```bash
docker logs -f openpanel_mysql
```

Példa:
```bash
root@openpanel:~# docker logs -f openpanel_mysql
[Entrypoint] MySQL Docker Image 8.0.32-1.2.11-server
[Entrypoint] Starting MySQL 8.0.32-1.2.11-server
2025-10-17T15:36:55.291441Z 0 [Warning] [MY-011068] [Server] The syntax '--skip-host-cache' is deprecated and will be removed in a future release. Please use SET GLOBAL host_cache_size=0 instead.
2025-10-17T15:36:55.293565Z 0 [System] [MY-010116] [Server] /usr/sbin/mysqld (mysqld 8.0.32) starting as process 1
2025-10-17T15:36:55.304336Z 1 [System] [MY-013576] [InnoDB] InnoDB initialization has started.
2025-10-17T15:36:55.546466Z 1 [System] [MY-013577] [InnoDB] InnoDB initialization has ended.
2025-10-17T15:36:55.852133Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2025-10-17T15:36:55.852166Z 0 [System] [MY-013602] [Server] Channel mysql_main configured to support TLS. Encrypted connections are now supported for this channel.
2025-10-17T15:36:55.852534Z 0 [ERROR] [MY-010259] [Server] Another process with pid 60 is using unix socket file.
2025-10-17T15:36:55.852547Z 0 [ERROR] [MY-010268] [Server] Unable to setup unix socket lock file.
2025-10-17T15:36:55.852555Z 0 [ERROR] [MY-010119] [Server] Aborting
2025-10-17T15:36:57.431776Z 0 [System] [MY-010910] [Server] /usr/sbin/mysqld: Shutdown complete (mysqld 8.0.32)  MySQL Community Server - GPL.
```

Ebben a példában a hiba a következő: *A pid 60-as másik folyamat unix socket fájlt használ.*

[Googlázva a hibát](https://stackoverflow.com/questions/36103721/docker-db-container-running-another-process-with-pid-id-is-using-unix-socket) megkapjuk a megoldást: `cd /root &&& docker compose down --volumes, ez a terv eltávolítja az összes létező felhasználót, beleértve a tartományt. csak friss telepítésen futtassa, pl. ha ez a mysql hiba az openpanel** telepítése után következik be.

* Másolja ki a hibaüzeneteket, és keressen online. A gyakori problémák a következők:

* Probléma a legújabb mysql docker képcímkével
* A MySQL nem fut megfelelően az ARM CPU-kon

---

## A MySQL parancs nem működik

A kapcsolatot közvetlenül a terminálról tesztelheti a "mysql" paranccsal.

Példa hiba:

```bash
root@openpanel:~# mysql
ERROR 2026 (HY000): TLS/SSL error: SSL is required, but the server does not support it
```

A hiba megoldásához *TLS/SSL hiba: SSL szükséges, de a szerver nem támogatja* futtassa: `echo "skip-ssl = true" >> /etc/my.cnf` [FORRÁS](https://stackoverflow.com/questions/78677369/mariadb-11-also-mysql-cli-error-2026-hy000-tls-ssl-error-ssl-is-required)

---


Példa hiba:
```bash
root@openpanel:~# mysql
ERROR 2003 (HY000): Can't connect to MySQL server on '127.0.0.1:3306' (111)
```

Ha ez egy friss telepítés, és még nincsenek felhasználók vagy tervek, előfordulhat, hogy a MySQL inicializálása nem sikerül a telepítés során. Ennek javításához:

```
cd /root
docker compose down
docker volume rm root_mysql # DANGER: this will delete all users and plans, only run it on a fresh install!
docker compose up -d openpanel_mysql openpanel
```

Ellenkező esetben ellenőrizze, hogy az "/etc/my.cnf" fájlban tárolt hitelesítő adatok helyesek-e:

```bash
root@demo:~# cat /etc/my.cnf 
[client]
user = panel
database = panel
password = e391ac94321d110c
host = 127.0.0.1
protocol = tcp
```

---

## Tűzfal

Engedélyezni kell a kimenő kapcsolatokat a **3306-os porton**. Győződjön meg arról, hogy ez a port nyitva van a Sentinel Firewallban (CSF).

Ellenőrizze az `/etc/csf/csf.conf` fájlban a `TCP_OUT=` beállítást, és győződjön meg arról, hogy a `3306-os port szerepel benne. Ha nem, adja hozzá, és indítsa újra a CSF-et:

```bash
csf -r
```

---

Ha a fenti lépések egyike sem oldja meg a problémát, lépjen kapcsolatba velünk a fórumon keresztül, vagy nyissa meg a támogatási jegyet, és segítünk a probléma elhárításában.

---
sidebar_position: 3
---

# Alapértelmezések

Az **OpenAdmin > Beállítások > Alapértelmezések** részben a rendszergazdák szerkeszthetik az új felhasználókhoz használt `docker-compose.yml` és `.env` fájlok értékeit.

![defaults basic](https://i.postimg.cc/KFRzLrGY/admin-defaults.png)

Ezek a fájlok határozzák meg az új felhasználók szolgáltatásait és korlátait.

> [Az Apache, az Nginx, az OpenResty, az OpenLiteSpeed ​​és a Varnish beállítása alapértelmezett webszerverként](/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
> [A MySQL, a MariaDB vagy a Percona beállítása alapértelmezett adatbázistípushoz] (/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/)

---

A 'Speciális' opcióval közvetlenül szerkesztheti a fájlokat.

![defaults advanced](https://i.postimg.cc/74BhfQyc/admin-defaults-advanced.png)

Ezekben a fájlokban további szolgáltatásokat (docker konténereket) konfigurálhat, és módosíthatja a meglévő szolgáltatások alapértelmezett beállításait.

Ne feledje, hogy ez haladó felhasználóknak készült, és a hibás konfigurálás szabaddá teheti a rendszerportokat, a felhasználó elkapja az erőforrásokat vagy túllépi a lemezkorlátokat.

Új szolgáltatások hozzáadásakor vegye figyelembe a következőket:

- A tároló nevének meg kell egyeznie a szolgáltatás nevével
- A szolgáltatás processzor- és memóriakorlátait a következő formátumban kell megnevezni: `SERVICE_`CPU és `SERVICE_`RAM.
- a szolgáltatás egyéb változóinál is elő kell írni a 'SERVICE_' ​​előtagot
- A tárolókban lévő folyamatokat root (0`) felhasználóként kell futtatni annak érdekében, hogy a tárolófájlokat beleszámítsák a felhasználói kvótába, és elkerüljék az engedélyekkel kapcsolatos problémákat.
- A konfigurációs fájlokat csak olvasható módban kell felcsatolni.


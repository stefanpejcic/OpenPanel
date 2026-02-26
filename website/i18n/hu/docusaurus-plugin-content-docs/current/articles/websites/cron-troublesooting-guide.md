# OpenPanel --- Cron hibaelhárítási útmutató

Ez egy **általános hibaelhárítási útmutató** az OpenPanel cron feladatokhoz.

-------------------------------------------------------------------------

## **1) Hogyan működik a cron az OpenPanelben**

OpenPanelben:

- A Cron **NEM fut közvetlenül a gazdagépen**, mint a hagyományos Linux
crontab.
- Minden **felhasználónak saját cron tárolója van**.
- A feladatok **meghatározott tárolóban** futnak (pl. „openlitespeed”,
"nginx", "php" stb.).
- A Cron jobok a **`/home/USERNAME/crons.ini`** fájlban vannak definiálva, nem pedig
`/etc/crontab`.

-------------------------------------------------------------------------

## **2) Első diagnosztikai lépés --- futtassa manuálisan a parancsot (még a cron-hoz kapcsolódik?)**

A cron hibaelhárítása előtt mindig ellenőrizze, hogy maga a parancs
művek.

Hozzáférés a felhasználói konténerterminálhoz:
https://openpanel.com/docs/panel/containers/terminal/#accessing-the-terminal

A tárolón belül futtassa manuálisan a parancsot, például:

``` bash
php -q /var/www/html/example.com/apps/console/console.php example-task
```

### Ha ez nem sikerül:

- A probléma **NEM cron**
- Először javítsa ki a hibát (hiányzó PHP kiterjesztés, rossz elérési út,
engedélyek, alkalmazáskonfiguráció stb.)

Csak akkor folytassa a cron hibaelhárítását, **miután ez manuálisan működik.**

-------------------------------------------------------------------------

## **3) Ellenőrizze, hogy fut-e a cron szolgáltatás**

Minden felhasználónak aktívnak kell lennie a saját cron szolgáltatásával.

Érkezés az OpenPanelben: https://openpanel.com/docs/panel/advanced/services/

Látnia kell egy **cron szolgáltatást, amely fut a felhasználó számára**.

### Ha NEM fut, indítsa el manuálisan (terminál):

``` bash
cd /home/USERNAME && docker --context=USERNAME compose up -d cron
```


-------------------------------------------------------------------------

## **4) Ellenőrizze a cron fájlformátumot (`crons.ini`)**

Ellenőrizze a felhasználó cron fájlját:

``` bash
cat /home/USERNAME/crons.ini
```

A helyes munka így néz ki:

``` ini
[job-exec "example-job"]
schedule = @every 1m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php example-task
```

### Gyakori hibák

❌ Gazda PHP elérési út használata:

``` bash
/usr/local/lsws/lsphp84/bin/php
```

✅ Használjon helyette PHP tárolót:

``` bash
php
```

❌ Rossz konténer:

``` ini
container = web
```

✅ Meg kell egyeznie a tényleges tárolóval:

``` ini
container = openlitespeed
```

-------------------------------------------------------------------------

## **5) Ellenőrizze a cron végrehajtási naplóit (nagyon fontos)**

Az OpenPanel munkánkénti végrehajtási naplókat biztosít.

Cron munkanaplók: https://openpanel.com/docs/panel/advanced/cronjobs/#logs

Keresse a következőket: - A munka elkezdődött - A munka sikertelen - PHP hibák - Engedély megtagadva -
"A tároló nem található"

-------------------------------------------------------------------------

## **6) Ellenőrizze a tárolón belüli útvonalakat**

A Dockeren belüli útvonalak eltérhetnek a gazdagép elérési útjaitól.

Ellenőrizze a tartály belsejét:

``` bash
docker --context USERNAME exec -it CONTAINER bash
ls /var/www/html
```

Erősítse meg, hogy:

/var/www/html/example.com/

valóban létezik a tárolóban.

-------------------------------------------------------------------------

## **7) Vizsgálja meg a cron tárolónaplóit**

Ha a feladatok továbbra sem futnak, ellenőrizze magát a cron tárolót:

``` bash
docker --context USERNAME logs cron
```

Keres:
- Ütemező hibák
- Konténerindítási problémák
- Engedélyezési problémák

-------------------------------------------------------------------------

## **8) Érvényesítse a megfelelő tárolóválasztást**

Győződjön meg arról, hogy a feladat a megfelelő tárolót használja:

Alkalmazástípus Valószínű tároló
  --------------------- ------------------
OpenLiteSpeed ​​+ PHP `openlitespeed` vagy `php`
Nginx + PHP `nginx` vagy `php`
Önálló tároló: „CONTAINER_NAME”.

Ha nem a megfelelőt választja → a munka meghiúsul.

-------------------------------------------------------------------------

## **9) Általános működő cron sablon**

Az „example.com” címen tárolt alkalmazás esetén:

``` ini
[job-exec "task-1"]
schedule = @every 1m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php task-1

[job-exec "task-2"]
schedule = @every 5m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php task-2

[job-exec "task-hourly"]
schedule = @hourly
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php hourly

[job-exec "task-daily"]
schedule = @daily
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php daily
```

-------------------------------------------------------------------------

## **10) Gyors hibaelhárítási döntési fa**

### ❓ "A Cron egyáltalán nem fut"

Ellenőrizze: - Fut a cron szolgáltatás? → Szolgáltatások oldal\
- Van cron konténer? → `docker --context FELHASZNÁLÓNÉV ps | grep cron`\
- Létezik a `/home/USERNAME/crons.ini`?

### ❓ "A Cron fut, de a parancs sikertelen"

Ellenőrizze: - Először futtassa manuálisan a tárolóban\
- Ellenőrizze a cron naplóit az OpenPanelben\
- Ellenőrizze a konténernaplókat

### ❓ "A Cron fut, de az alkalmazás továbbra is panaszkodik"

Ellenőrizze: - Alkalmazásnaplók a `/var/www/html/example.com/.../logs/`\ mappában
- Fájlengedélyek a tárolón belül

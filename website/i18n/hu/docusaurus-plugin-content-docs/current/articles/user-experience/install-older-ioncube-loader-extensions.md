# Egyéni vagy régebbi IonCube Loader verziók telepítése az OpenPanelben

Az **OpenPanel 1.7.2-es verziójától** kezdve az IonCube Loader automatikusan elérhető minden olyan PHP-verzióhoz, amely támogatja az új telepítéseknél.

## Régebbi verziók

Ha egy régebbi ionCube Loader csomagra szeretne visszaváltani (vagy egyéni csomagot szeretne használni), ezt egyszerűen megteheti.

Miután elhelyezte a fájlokat a szerveren:

* **Egyetlen felhasználó esetén:** Szerkessze a `docker-compose.yml` fájljukat, és frissítse a **php-fpm-*** szolgáltatások kötetbeillesztési pontjait, hogy azok hivatkozzanak az Ön egyéni fájljaira.
* **Minden új felhasználó számára:** Szerkessze az `/etc/openpanel/docker/compose/1.0/docker-compose.yml` címen található sablonfájlt.

## Ellenőrizze, hogy engedélyezve van-e

Az alábbi módszerek bármelyikével ellenőrizheti, hogy az IonCube Loader aktív-e az Ön PHP-verziójához:

A terminál futtatásából:

```bash
php -i | grep ioncube
```

Vagy hozzon létre egy **info.php** nevű fájlt a domain nyilvános könyvtárában a következő tartalommal:

```php
<?php
phpinfo();
```

Ezután nyissa meg a böngészőjében:
`https://yourdomain.tld/info.php`

Keresse meg az **IonCube PHP Loader** elemet a kimenetben, hogy ellenőrizze, hogy engedélyezve van-e.

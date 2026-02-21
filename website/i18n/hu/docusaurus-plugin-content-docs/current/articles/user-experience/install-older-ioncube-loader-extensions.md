# How to Install Custom or Older IonCube Loader Versions in OpenPanel

Starting with **OpenPanel version 1.7.2**, IonCube Loader is automatically available for all PHP versions that support it on new installations.

## Older Versions

If you want to downgrade to an older ionCube Loader bundle (or use a custom bundle) you can do so easily.

After placing the files on your server:

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

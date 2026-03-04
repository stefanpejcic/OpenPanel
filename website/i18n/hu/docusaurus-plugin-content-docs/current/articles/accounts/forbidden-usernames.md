# Fenntartott felhasználónevek

## Alapvető korlátozások


### OpenPanel

Az OpenPanel a következő szabályokat alkalmazza az OpenPanel felhasználónév létrehozásakor vagy módosításakor:

A felhasználónevekben **csak** kisbetűk (`a–z`) és számjegyek (`0-9`) szerepelhetnek.
- A felhasználónevek nem tartalmazhatnak **3** karakternél kevesebbet.
- A felhasználónevek nem tartalmazhatnak **20** karakternél többet.

### OpenAdmin
Az OpenAdmin a következő szabályokat alkalmazza a rendszergazdai felhasználónév létrehozásakor vagy módosításakor:

A felhasználónevekben **csak** kisbetűk (`a–z`) és számjegyek (`0-9`) szerepelhetnek.
- A felhasználónevek nem tartalmazhatnak **5** karakternél kevesebbet.
- A felhasználónevek nem tartalmazhatnak **30** karakternél többet.

A jelszavak **csak** kisbetűket (`a–z`) és számjegyeket (`0-9`) tartalmazhatnak.
- A jelszavak nem tartalmazhatnak **6** karakternél kevesebbet.
- A jelszavak nem tartalmazhatnak **30** karakternél többet.

:::info
Ha adminisztrátori vagy viszonteladói fiókokat az OpenAdmin segítségével hoz létre, a jelszavak a szóköz kivételével minden karaktert tartalmazhatnak.
:::

### SSH
Az OpenAdmin a következő szabályokat alkalmazza az SSH-jelszó módosításakor:

A jelszavak **csak** kisbetűket (`a–z`) és számjegyeket (`0-9`) tartalmazhatnak.
- A jelszavak nem tartalmazhatnak **8** karakternél kevesebbet.
- A jelszavak nem tartalmazhatnak **20** karakternél többet.

### MySQL / MariaDB
Az OpenPanel a következő szabályokat alkalmazza adatbázisok és felhasználók létrehozásakor:

Az adatbázisnevek **csak** tartalmazhatnak kisbetűket (`a-z`) és számjegyeket (`0-9`).
- Az adatbázis neve nem tartalmazhat **1** karakternél kevesebbet.
- Az adatbázis neve legfeljebb **64** karaktert tartalmazhat.


Az adatbázis-felhasználók **csak** használhatnak kisbetűket (`a-z`) és számjegyeket (`0-9`).
- A felhasználónevek nem tartalmazhatnak **1** karakternél kevesebbet.
- A felhasználónevek nem tartalmazhatnak **32** karakternél többet.


Az adatbázis felhasználói jelszavai **csak** tartalmazhatnak kisbetűket (`a–z`), aláhúzásjeleket (`_`) és számjegyeket (`0-9`).
- A jelszavak nem tartalmazhatnak **8** karakternél kevesebbet.
- A jelszavak nem tartalmazhatnak **32** karakternél többet.


### FTP
Az OpenPanel a következő szabályokat alkalmazza az FTP-alfiókok létrehozásakor:

Az FTP-alfelhasználóknak **ponttal (`.`) kell végződniük**, amelyet az OpenPanel-felhasználónév követ – például: `ftpuser.openpaneluser`.
- A felhasználónevek nem tartalmazhatnak **3** karakternél kevesebbet.
- A felhasználónevek nem tartalmazhatnak **32** karakternél többet.


Az FTP-felhasználó jelszavainak tartalmazniuk kell **legalább egy** nagybetűt (`A-Z`), kisbetűt (`a-z`), számjegyet (`0-9`) és speciális szimbólumokat (`!`, `.`, `,`, `@`, `#`, `_`, `-`).
- A jelszavak nem tartalmazhatnak **8** karakternél kevesebbet.
- A jelszavak nem tartalmazhatnak **32** karakternél többet.

### E-mailek

Az OpenPanel a következő szabályokat alkalmazza az e-mail fiókok létrehozásakor:

Az e-mail fiókok **csak** használhatnak kisbetűket (`a–z`), kötőjelet (`-`), aláhúzásjelet (`_`) és számjegyeket (`0-9`).
Az e-mail fiókoknak ** kell tartalmazniuk** egy „@” szimbólumot, amelyet a domain név követ – például: „account@example.net”.
- A felhasználónevek nem tartalmazhatnak **1** karakternél kevesebbet.
- A felhasználónevek nem tartalmazhatnak **32** karakternél többet.

Az e-mail fiókok jelszavai **csak** nagybetűket (`A-Z`), kisbetűket (`a-z`), aláhúzásjelet (`_`) és számjegyeket (`0-9`) tartalmazhatnak.
- A jelszavak nem tartalmazhatnak **8** karakternél kevesebbet.
- A jelszavak nem tartalmazhatnak **32** karakternél többet.


## Fenntartott felhasználónevek

Az OpenPanel fenntart néhány felhasználónevet a rendszer számára, és ezeket nem használhatja OpenPanel fiókokhoz. A fenntartott felhasználónevek listája idővel bővülhet, és az OpenPanel új verziói is bővülhetnek a listán.

Az OpenPanel a következő fájlban ellenőrzi, hogy le kell-e foglalni vagy korlátozni kell-e egy felhasználónevet:

```bash
/etc/openpanel/openadmin/config/forbidden_usernames.txt
```

Jelenleg lefoglalt felhasználónevek:



- 1000
- admin
- apache
- apache2
- biztonsági mentés
- busybox
- cron
- dokkoló
- rugalmas keresés
- végrehajtó
- fájlböngésző
- ftp
- httpd
- Litespeed
- openlitespeed
- lsws
- mariadb
- memcached
- minecraft
- mssql
- mysql
- mysqld
- nginx
- openadmin
- nyitott panel
- nyitottság
- opensearch
- pgadmin
- php
- php-fpm-5.6
- php-fpm-7.0
- php-fpm-7.1
- php-fpm-7.2
- php-fpm-7.3
- php-fpm-7.4
- php-fpm-8.0
- php-fpm-8.1
- php-fpm-8.2
- php-fpm-8.3
- php-fpm-8.4
- php-fpm-8.5
- phpmyadmin
- podman
- postgres
- reboot
- redis
- restart
- root
- shutdown
- test
- user_service
- varnish
- vsftpd
- www-data

To reserve a certain username, simply add it to the forbidden_usernames.txt file.

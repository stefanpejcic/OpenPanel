# PHP-bővítmény telepítése az OpenPanel-ben

Az OpenPanel a [`shinsenter/php`](https://github.com/shinsenter/php) Docker-képeket használja PHP-szolgáltatásokhoz. Ezek a képek számos általánosan használt PHP-bővítményt tartalmaznak, és egyszerű módot kínálnak további bővítmények telepítésére, ha szükséges.

## Előre telepített PHP-bővítmények

A legtöbb népszerű PHP-bővítmény alapértelmezés szerint előre telepítve van, lehetővé téve a projektek gyors elindítását extra konfiguráció nélkül:

```
apcu
bcmath
calendar
Core
ctype
curl
date
dom
exif
fileinfo
filter
gd
gettext
gmp
hash
iconv
igbinary
intl
json
lexbor
libxml
mbstring
msgpack
mysqli
mysqlnd
openssl
pcntl
pcre
PDO
pdo_mysql
pdo_pgsql
pdo_sqlite
pgsql
Phar
posix
random
readline
redis
Reflection
session
SimpleXML
sodium
SPL
sqlite3
standard
tidy
tokenizer
uri
uuid
xml
xmlreader
xmlwriter
yaml
Zend OPcache
zip
zlib
```

és az OpenPanel hozzáadja a "Zend OPcache" kiterjesztést.

----

## Egyéni PHP-bővítmények hozzáadása

A PHP-képek tartalmaznak egy `phpaddmod' nevű segédparancsot a PHP-bővítmények telepítéséhez: telepíti a kiterjesztést, és szerkeszti a php.ini fájlt, hogy engedélyezze azt.


### PHP-bővítmény telepítése egyetlen felhasználó számára

Ha egyetlen felhasználó számára szeretne PHP-bővítményt telepíteni, nyisson meg egy terminált a kívánt PHP-verzióhoz.

### Példafeltevés

* **Felhasználónév**: `stefan`
* **PHP verzió**: `8.5`
* **Szolgáltatás neve**: `php-fpm-8.5`

Állítsa be ezeket az értékeket a környezetének megfelelően.

---

### 1. lehetőség: OpenPanel terminál (Enterprise Edition)

Ha a **Docker** modul engedélyezve van:

1. Nyissa meg az OpenPanel **Containers → Terminal** menüpontját
2. Válassza ki a kívánt PHP verziót
3. Futtassa a következő parancsokat

Példa: a **GMP** bővítmény telepítése:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### 2. lehetőség: OpenAdmin Terminal

1. Lépjen a **Felhasználók → (felhasználó kiválasztása) → Szolgáltatások** lehetőségre.
2. Nyissa meg a kívánt PHP-verzió terminált
3. Futtassa a következő parancsokat

Példa: a **GMP** bővítmény telepítése:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### 3. lehetőség: SSH-hozzáférés (root)

1. Csatlakozzon a szerverhez SSH-n keresztül "root"-ként
2. Adja meg a kívánt felhasználóhoz és verzióhoz tartozó PHP-tárolót:

```
docker --context=USERNAME_HERE exec -it php-fpm-8.5 bash
```

3. Telepítse a bővítményt

Példa: a **GMP** bővítmény telepítése:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### Ellenőrizze a telepítést

A bővítmény telepítését a következő futtatással ellenőrizheti:

```
php -m
```

Vagy az alkalmazás vagy a webhely ellenőrzésével.

---

## PHP-bővítmények telepítése minden új felhasználó számára

Ahhoz, hogy egy PHP-bővítmény automatikusan elérhető legyen **minden újonnan létrehozott felhasználó számára**, két lehetősége van:

1. **Módosítsa az indítási parancsot** a `docker-compose.yml' sablonban a bővítmény telepítéséhez a tároló indításakor.

* Ez megnöveli a PHP tároló indítási idejét

2. **Használjon egyéni PHP Docker-képet** a szükséges bővítményekkel előre telepítve

* Termelési környezetekhez ajánlott

> Ezek a megközelítések haladó felhasználók számára készültek. Ha segítségre van szüksége, forduljon az OpenPanel közösségi fórumhoz vagy a Discord csatornához.

---

## PHP-bővítmények telepítése az összes meglévő felhasználóhoz

Ha PHP-bővítményt szeretne telepíteni **minden meglévő felhasználóhoz**, a következőket teheti:

* Futtasson egy egyéni Bash-szkriptet, amely minden felhasználóhoz és PHP-verzióhoz telepíti a bővítményt
* Váltson át minden felhasználót egy egyéni PHP Docker-képre, amely már tartalmazza a kiterjesztést

Mindkét megközelítéshez szükséges a Docker és az OpenPanel belső részeinek fejlett ismerete.

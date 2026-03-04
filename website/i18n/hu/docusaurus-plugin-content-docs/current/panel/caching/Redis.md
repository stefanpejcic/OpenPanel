---
sidebar_position: 2
---

# REDIS

![redis_disabled.png](/img/panel/v2/redismain.png)

A REDIS egy robusztus és tartós objektum-gyorsítótár megoldás, amelyet arra terveztek, hogy hatékonyan tárolja a gyakran használt webhelyadatokat a RAM memóriájában.

A RAM-ból származó adatok tárolására és kiszolgálására való képességével minimálisra csökkenti az időigényes adatbázis-lekérdezések szükségességét, jelentősen javítva a webhely teljesítményét.

## Engedélyezés / letiltása

Lehetősége van szükség szerint engedélyezni vagy letiltani a REDIS szolgáltatástárolót. A letiltása azonnal törli (törli) az összes meglévő Redis-adatot a memóriából.

A REDIS szolgáltatástároló engedélyezése elindítja a Redis tárolószolgáltatást az alapértelmezett porton, amely a _6379_.

![redis_enabled.png](/img/panel/v2/redisenabled.png)

## Állítsa be a memóriakorlátokat

Inicializáláskor a Redis tároló alapértelmezett memóriakorlátokat állít be, célszerű az Ön használati esetének és igényeinek megfelelő memóriakorlátokat beállítani.

Ezeket a korlátozásokat a /containers felületen állíthatja be, amely a Docker/containers felhasználói panel navigációján keresztül érhető el.

![redis_limits.png](/img/panel/v2/redislimits.png)

:::info
A memóriakorlát módosításához újra kell indítani a Redis-tárolót az új korlátozások alkalmazásához, ami az összes meglévő adat eltávolítását eredményezi a gyorsítótárból.
:::

## Csatlakozás a REDIS-hez

A REDIS-példányhoz való kapcsolat létrehozásához használja a következő adatokat:

- Szerver címe: **redis** (nem 127.0.0.1)
- Port: **6379** (az alapértelmezett REDIS port)

A REDIS szerverrel való kapcsolat teszteléséhez használhatja a terminál 'telnet' parancsát, vagy tekintse meg az alábbi példaszkripteket.

### Tesztelje a kapcsolatot a PHP-vel

1. A FileManager segítségével lépjen be webhelye könyvtárába
2. Hozzon létre egy új fájlt _redis-test.php_ néven
3. Adja hozzá a következő kódot az újonnan létrehozott fájlhoz, és mentse:
```php
<?php 
   //Connecting to Redis server on localhost 
   $redis = new Redis(); 
   $redis->connect('redis', 6379); 
   echo "Connection to server successfully"; 
   //check whether server is running or not 
   echo "Server is running: ".$redis->ping(); 
?>
```

4. Böngészőjében keresse meg webhelyét, és adja hozzá például a /redis-test.php fájlt. Ha webhelye example.com, akkor nyissa meg az example.com/redis-test.php webhelyet.

Látnia kell a _Server is running..._ üzenetet, amely jelzi, hogy a REDIS szolgáltatás aktív, és a kapcsolat létrejött.


### WordPress bővítmények

A REDIS gyorsítótár beépítéséhez a WordPress webhelyébe egy WordPress beépülő modul szükséges. Az alábbiakban felsorolunk néhány WordPress-bővítményt, amelyeket a REDIS gyorsítótárazásához értékeltünk:

- [Redis Object Cache](https://wordpress.org/plugins/redis-cache/)
- [A Redis használata](https://plugins.club/premium-wordpress-plugins/use-redis/)
   
## Naplók megtekintése

Lehetősége van a REDIS szolgáltatásnaplóinak megtekintésére. Ezzel azonosíthatja a szolgáltatási hibákat, vagy például megállapíthatja, hogy elérte-e a memóriakorlátokat.

![redis_log.png](/img/panel/v2/redislogs.png)

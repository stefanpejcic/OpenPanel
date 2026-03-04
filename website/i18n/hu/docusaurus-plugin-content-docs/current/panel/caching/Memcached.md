---
sidebar_position: 3
---

# Memcached

![memcached_disabled.png](/img/panel/v2/memcmain.png)

A Memcached egy nagy teljesítményű, elosztott memória-gyorsítótár rendszer. Gyakran használják a dinamikus adatbázis-vezérelt webhelyek és alkalmazások felgyorsítására az adatok memóriában való gyorsítótárazásával.

Általában arra használják, hogy csökkentsék az adatbázis-kiszolgáló terhelését és javítsák a webhelyek válaszkészségét a gyakran elért adatok gyorsítótárazásával, például az adatbázis-lekérdezések eredményei vagy az API-válaszok.

## Engedélyezés / letiltása

Lehetősége van szükség szerint engedélyezni vagy letiltani a Memcached szolgáltatást. A letiltása azonnal törli az összes meglévő Memcached adatot a memóriából.

A Memcached szolgáltatás engedélyezése a szolgáltatás az alapértelmezett Memcached porton indul el, amely _11211_.

![memcached_enabled.png](/img/panel/v2/memcenabled.png)

## Állítsa be a memóriakorlátokat

Inicializáláskor a Memcached tároló alapértelmezett memóriakorlátokat állít be, célszerű az Ön használati esetének és igényeinek megfelelő memóriakorlátokat beállítani.

Ezeket a korlátozásokat a /containers felületen állíthatja be, amely a Docker/containers felhasználói panel navigációján keresztül érhető el.

:::info
A memóriakorlát módosítása szükségessé teszi a szolgáltatás újraindítását az új korlátozások alkalmazásához, ami az összes meglévő gyorsítótáradat eltávolítását eredményezi.
:::

![memcached_limits.png](/img/panel/v2/memclimits.png)

## Csatlakozás a Memcachedhez

A Memcached példányhoz való kapcsolat létrehozásához használja a következő adatokat:

- Szerver címe: **memcached** (nem 127.0.0.1)
- Port: **11211** (az alapértelmezett Memcached port)

A Memcached szerverrel való kapcsolat teszteléséhez a következő eszközöket vagy szkripteket használhatja.

### Teszt kapcsolat PHP-vel

1. A Fájlkezelő segítségével keresse meg webhelye könyvtárát.
2. Hozzon létre egy új fájlt _memcached-test.php_ néven.
3. Adja hozzá a következő PHP-kódot az újonnan létrehozott fájlhoz, és mentse el:

```php
<?php 
   // Connect to Memcached server on localhost 
   $memcached = new Memcached(); 
   $memcached->addServer('memcached', 11211); 
   echo "Connection to server successful"; 
   // Check whether the server is running or not 
   echo "Server is running: ".$memcached->getVersion(); 
?>
```

Nyissa meg webhelyét egy böngészőben, és fűzze hozzá a /memcached-test.php fájlt. Például, ha webhelye az example.com, nyissa meg az example.com/memcached-test.php címet

Látnia kell a "Szerver fut..." üzenetet, amely azt jelzi, hogy a Memcached szolgáltatás aktív, és a kapcsolat létrejött.


### WordPress beépülő modulok

A Memcached gyorsítótárazás megvalósításához a WordPress webhelyén szükség van egy dedikált bővítményre. Íme néhány WordPress-bővítmény, amelyeket a Memcached gyorsítótárazáshoz teszteltünk:

- [Memcached Object Cache](https://wordpress.org/plugins/memcached/)

## Naplók megtekintése

Lehetősége van hozzáférni a Memcached szolgáltatásnaplókhoz. Ezzel azonosíthatja a szolgáltatási hibákat, vagy ellenőrizheti a memóriahasználatot és a korlátokat.

![memcached_log.png](/img/panel/v2/memclogs.png)


---
sidebar_position: 1
---

# Gyorsítótár

Az OpenPanel által kínált készenléti funkciók tömbjének felhasználása jelentősen növelheti webhelye teljesítményét és biztonságát. Az olyan funkciók, mint a PHP-FPM, a Redis, a Memcached és az Opcache, jelentősen növelhetik webhelye sebességét és hatékonyságát.

## REDIS

A Redis állandó objektum-gyorsítótár háttér-gyorsítótár-kiszolgálóként szolgál, elsősorban az adatbázisokkal és webhelyekkel kapcsolatos hívások és lekérdezések felgyorsítására a RAM-memóriában való tárolással. A RAM a nagy sebességű teljesítményéről ismert, még az NVMe-t és az UFS-t is felülmúlja, így a Redis gyorsítótárazás hatékony eszköz a webhely teljesítményének optimalizálásához. A Redis gyorsítótár kihasználása jelentős előnyökkel járhat webhelye számára.

[REDIS beállítások és használat](/docs/panel/caching/Redis)

## Memcached

Ezt a memórián belüli (RAM) objektum-gyorsítótárat kifejezetten az adatbázis-terhelés csökkentésére tervezték, így ideális dinamikus webhelyekhez. Csak az adatbázishoz kapcsolódó lekérdezéseket gyorsítótárazza.

Nem javasoljuk, hogy más gyorsítótárral, például a Redis-szel együtt használja. Azonban hatékonyan használható a PHP OPcache mellett a webhely teljesítményének javítása érdekében.

[Memcached konfiguráció és használat](/docs/panel/caching/Memcached)

## Elaszticsearch

Növelje webhelye keresési lehetőségeit és általános teljesítményét az Elasticsearch segítségével. Az OpenPanel zökkenőmentes integrációt biztosít az Elasticsearch hatékony keresőmotorral, amely lehetővé teszi az információk hatékony és gyors visszakeresését.

Konfigurálja és optimalizálja az Elasticsearch beállításait, hogy a webhely igényeihez igazítsa. Ismerje meg az Elasticsearch által kínált különféle funkciókat és funkciókat, hogy a legtöbbet hozza ki ebből a robusztus keresőmotorból.

[Elasticsearch beállításai és használata](/docs/panel/caching/elasticsearch)

## Opcache
Az OPcache értékes eszköz a PHP teljesítményének javítására. Úgy működik, hogy az előre lefordított szkriptkódot az osztott memóriában tárolja, így nincs szükség arra, hogy a PHP minden kérésnél újratöltse és elemezze a szkripteket. Egyszerűbben fogalmazva, az OPcache gyorsítótárazza a korábban végrehajtott PHP kódot, csökkentve a CPU terhelését és javítva a webhely teljesítményét.

**Ez a funkció alapértelmezés szerint engedélyezve van PHP használatakor, és nincs szükség további beállításokra.**

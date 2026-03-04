# 404 Hibaelhárítási útmutató

Ez az útmutató ismerteti a webhelyeken előforduló **404 nem található** hibák leggyakoribb okait és azok megoldását.

---

## 1. A fájl nem létezik

A legegyszerűbb és leggyakoribb ok: a fájl nincs ott.
Ha statikus fájlt kér, ellenőrizze, hogy az valóban létezik-e a megadott elérési úton vagy dokumentumgyökérben.

**Példa:**
Ha az "example.com/index.html" címet kéri, ellenőrizze, hogy az "index.html" létezik-e az "example.com" tartomány dokumentumgyökérében.

**Hogyan ellenőrizhető:**

* Nyissa meg a **Domainek > [Saját domain] > *kattintson a* Dokumentumgyökér** elemre.
* Nyissa meg a Fájlkezelőt, és ellenőrizze, hogy a fájl a megfelelő helyen van-e.
![screenshot](https://i.postimg.cc/TRPPtj2P/image.png)
---

## 2. Szabályok újraírása

Ha a kérés dinamikus útvonalra vonatkozik (pl. szép WordPress állandó hivatkozásokra, például `example.com/about`), a problémát a `htaccess` fájl vagy a VirtualHosts konfiguráció rosszul konfigurált átírási szabályai okozhatják.

**Apache vagy LiteSpeed ​​esetén:**

* Nyissa meg a Fájlkezelőt, és ellenőrizze, hogy van-e `.htaccess` fájl a tartomány dokumentumgyökérében.
![képernyőkép](https://i.postimg.cc/b8nwVwP1/image.png)
* Ha létezik, tekintse át a tartalmát.
![screenshot](https://i.postimg.cc/338rwjvZ/image.png)
* WordPress esetén egyszerűen lépjen a [**WP Admin > Settings > Permalinks**](https://www.google.com/search?q=wordpress+recreate+htaccess) oldalra, és kattintson a **Mentés** gombra (még változtatás nélkül is). Ez frissíti a `.htaccess` átírási szabályait.
* Ezután indítsa újra a webszervert a .htaccess fájl módosításainak alkalmazásához.

**Nginx vagy OpenResty esetén:**

* A `.htaccess` fájlok nem használatosak. Az újraírási szabályok közvetlenül a VirtualHost konfigurációjában vannak beállítva.
* Lépjen a **Domains > [Your Domain] > Edit VHosts** menüpontra.
![screenshot](https://i.postimg.cc/mbSrwXWJ/image.png)
* Tekintse át és javítsa az átírási szabályokat, majd mentse el a fájlt. Ez automatikusan újratölti a webszervert a változtatások alkalmazásához.
![screenshot](https://i.postimg.cc/RCSpt0Q2/image.png)

A módosítások után tesztelje újra webhelyét.

---

## 3. Docker Mount Problémák

Ritka esetekben a Docker fájlillesztő rendszere (amely inode-okat használ) eltéréseket okozhat. Ha egy felhasználó vagy program áthelyezett vagy átnevezett egy saját könyvtárat, előfordulhat, hogy a PHP szolgáltatástároló már nem fogja látni ugyanazokat a fájlokat, mint a Fájlkezelő.

Ez azt jelenti, hogy a fájl megjelenhet a Fájlkezelőben, de nem érhető el a PHP számára.

**Hogyan ellenőrizhető:**

* Ha rendelkezik Docker-hozzáféréssel:

1. Lépjen a **Docker > Terminal** elemre.
2. Válassza ki a PHP szolgáltatást.
![screenshot](https://i.postimg.cc/qphbySyB/image.png)
3. Fuss:

``` bash
ls fájlnév_itt.php
     ```

Ha a fájl nem jelenik meg, indítsa újra a PHP szolgáltatást az újracsatlakozáshoz.

* A PHP tároló újraindítása:

* Lépjen a **Docker > Containers** oldalra.
* Tiltsa le a PHP szolgáltatást, várja meg az oldal újratöltését, majd engedélyezze újra.
![képernyőkép](https://i.postimg.cc/wx8Dm4XP/image.png)

* Ha nem rendelkezik Docker-hozzáféréssel:

* Lépjen kapcsolatba tárhelyszolgáltatójával, és kérje a PHP szolgáltatás újraindítását.

Ha végzett, ellenőrizze újra a fájlt a terminál segítségével.

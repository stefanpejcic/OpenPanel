# 502 Hibaelhárítási útmutató

Ez az útmutató ismerteti az **502 Bad Gateway** hibák leggyakoribb okait és azok kijavítását.

---

## 1. A PHP-FPM nem fut

Az 502 Bad Gateway hiba gyakori az Nginx használatakor. Ez azt jelenti, hogy a webszerver (Nginx) nem kapott választ a háttérrendszertől (például PHP-FPM, Node.js vagy Python).

**A PHP-FPM ellenőrzésének lépései:**

1. Ellenőrizze, hogy fut-e a PHP szolgáltatás.
Ha a Docker funkció engedélyezve van, lépjen a **Docker > Tárolók** elemre, és ellenőrizze a PHP-tároló állapotát.
Ha inaktív, indítsa el.
![képernyőkép](https://i.postimg.cc/wx8Dm4XP/image.png)

---

## 2. Indítsa újra az Nginxet

Még ha fut is a PHP-FPM, előfordulhat, hogy az Nginx korábban elindult, és az első próbálkozásra nem tudott csatlakozni, ami gyorsítótári hibákat okoz.

**Javítás:** Indítsa újra az Nginxet.

* Ha a Docker funkciót használja:
Lépjen a **Docker > Containers** elemre, tiltsa le az Nginxet, majd engedélyezze újra.
![screenshot](https://i.postimg.cc/jRdPP7fs/image.png)

* Ha nem fér hozzá a Docker funkcióhoz, egyszerűen adjon hozzá egy másik domaint az Nginx újraindításához:
Lépjen a **Domainek > Új hozzáadása** elemre, írja be a domain nevet, és kattintson a "Hozzáadás" gombra.
![screenshot](https://i.postimg.cc/Bs9FNyF7/image.png)

---

## 3. Tiltsa le a Lakk gyorsítótárat

Néha a Varnish gyorsítótáraz egy 502-es hibát, még akkor is, ha a háttérprobléma megoldódott. Ha a Varnish Cache-t használja:

**Lépések:**

1. Lépjen a **Gyorsítótár > Lakk** elemre, és tiltsa le az érintett tartományban.
![screenshot](https://i.postimg.cc/0krVf4Jy/image.png)
2. Töltse be újra a webhelyet a böngészőjében.

A gyorsítótárat úgy is megkerülheti, hogy paramétereket ad hozzá az URL-hez:

* Példa: "example.net?test" vagy "example.net?cachebypass"

Ha a Varnish letiltása megoldja a problémát, állítsa le és indítsa el az összes gyorsítótárban lévő bejegyzés törléséhez.

---

## 4. Ellenőrizze a VirtualHostokat

A rosszul konfigurált VirtualHost miatt az Nginx proxykérelmeket küldhet egy már nem létező háttérrendszerhez. ha hozzáfér az *edit_vhost* szolgáltatáshoz:

**Lépések:**

1. Nyissa meg a **Domains > [Select Domain] > Edit VHosts** menüpontot.
![screenshot](https://i.postimg.cc/mbSrwXWJ/image.png)
2. Ellenőrizze a proxy célját (pl. PHP-szolgáltatás nevét vagy tárolóját).
![screenshot](https://i.postimg.cc/rVPMn7nn/image.png)
3. Győződjön meg arról, hogy a szolgáltatásnév helyesen van írva, és továbbra is létezik.

---

Ezekkel a lépésekkel képes lesz azonosítani és kijavítani az **502 Bad Gateway** hibák leggyakoribb okait.

# 403 Hibaelhárítási útmutató

Ez az útmutató a webhelyeken előforduló **403 tiltott** hibák leggyakoribb okait és azok megoldását ismerteti.

---

## 1. Coraza WAF

Az egyik leggyakoribb ok az, hogy a **Coraza WAF** blokkolja a hozzáférést a **ModSecurity CoreRuleSet** szabályai szerint.

1. Lépjen a **Speciális > WAF** elemre a hosting vezérlőpulton.
![WAF letiltása](https://i.postimg.cc/fZw2Skqv/waf-status.png)

2. Ellenőrizze, hogy a WAF engedélyezve van-e a tartományban. Ideiglenesen kapcsolja ki, majd tesztelje a webhelyet, hogy megbizonyosodjon arról, hogy a WAF okozza-e.

3. Ha a WAF a probléma:

* **Nem ajánlott:** Hagyja letiltva a WAF-ot a domainben.
* **Jobb lehetőség:** Határozza meg pontosan a kiváltott szabályt, és csak azt a szabályt tiltsa le.

4. A szabály megtalálásához:

* Kattintson a **Naplók megtekintése** lehetőségre a ModSecurity naplók megnyitásához.
* Keresse meg 403-as kérését IP, idő vagy kérés részletei alapján.
* Vegye figyelembe a **szabályazonosítót**.

5. Nyissa meg a **Szabályok kezelése** részét a domainhez a WAF részben:

* Letiltás szabályazonosítóval: [SecRuleRemoveByID](https://coraza.io/docs/seclang/directives/#secruleremovebyid)
* Címke tiltása: [SecRuleRemoveByTag](https://coraza.io/docs/seclang/directives/#secruleremovebytag)

![WAF szerkesztési szabályok](https://i.postimg.cc/GcSm9Xzm/2025-08-13-11-58.png)

6. Engedélyezze újra a WAF-ot, és tesztelje újra a webhelyet.

---

## 2. WordPress

A tartalomkezelő rendszerek, például a **WordPress** .htaccess fájlokat használnak, amelyek blokkolhatják a hozzáférést.

1. Ideiglenesen nevezze át vagy távolítsa el a ".htaccess" fájlt webhelye gyökérkönyvtárában.
2. Hozzon létre egy egyszerű tesztfájlt (pl. `index.php`), és nyissa meg a böngészőben.
3. Ha betöltődik, a probléma a CMS-hez kapcsolódik.
4. Lehetséges javítások:

* Ideiglenesen tiltsa le a WordPress bővítményeket.
* Állítsa vissza az alapértelmezett `.htaccess` fájlt.
* Kérjen segítséget CMS-fórumaiban vagy az [OpenPanel-közösségben](https://community.openpanel.org/).

---

## 3. Fájlengedélyek

A helytelen fájltulajdonlás vagy engedélyek 403-as hibákat is kiválthatnak.

1. Nyissa meg a **Fájlkezelőt**, és keresse meg a domain könyvtárát.

2. Kattintson a **Opciók** lehetőségre, és engedélyezze a **Tulajdonos** és **Csoport** oszlopokat.
![Fájlkezelő tulajdonos](https://i.postimg.cc/cZ4FdrY4/2025-08-13-11-52.png)

3. Győződjön meg arról, hogy minden fájlnak ugyanaz a tulajdonosa és csoportja.

4. Ellenőrizze a fájlok és mappák engedélyeit.

5. Ha a tulajdonjog helytelen, használja a **Fájlok > Engedélyek javítása** lehetőséget az alapértelmezett értékek visszaállításához.

---

## 4. Érzékeny fájlok

Alapértelmezés szerint a webszerver blokkolja a hozzáférést az érzékeny fájlokhoz:

```
.git
composer.json
composer.lock
auth.json
config.php
wp-config.php
vendor
```

Ha biztos abban, hogy ezeket a fájlokat elérhetővé kívánja tenni (nem ajánlott), módosítsa a tartomány VHost konfigurációját.

### OpenResty / Nginx esetén:

```nginx
# <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
location ~* ^/(\.git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor) {
    deny all;
    return 403;
}
# <!-- END EXPOSED RESOURCES PROTECTION -->
```

### Apache esetén:

```apache
# <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
<Directory <DOCUMENT_ROOT>>
    <FilesMatch "\.(git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor)">
        Require all denied
    </FilesMatch>
</Directory>
# <!-- END EXPOSED RESOURCES PROTECTION -->
```


---
sidebar_position: 1
---

# WordPress Manager

![wp_manager_grid.png](/img/panel/v2/wpmanager.png)

A WordPress Manager az Ön többfunkciós eszköze az OpenPanelben, amellyel WordPress webhelyeket telepíthet és kezelhet – anélkül, hogy valaha is be kellene jelentkeznie a wp-adminba. Több webhely kezelését gyorsítja, egyszerű és hatékony.

## WordPress-webhelyek kezelése

A WordPress Manager segítségével módosíthatja a beállításokat, biztonsági másolatokat készíthet, frissítheti a bővítményeket, átkapcsolhatja a hibakeresést és sok mást – mindezt közvetlenül az OpenPanelről. Nem kell több irányítópultot megnyitnia, vagy több tucat bejelentkezést megjegyezni.
Tökéletes ügynökségek, fejlesztők és bárki számára, aki több WordPress webhelyet kezel egyszerre.

### WP Manager

A fő WP Manager oldalon a következőket teheti:

- [Telepítések megtekintése](#wp-manage): tekintse meg a domaint, a WordPress verzióját, a telepítés dátumát és a rendszergazdai e-mail-címet.
- [Webhelyadatok frissítése](#refresh-website-data): ha domaint váltott, manuálisan frissítette a WordPresst vagy módosította az adminisztrátori e-mail-címet.
- [Témák és bővítménykészletek kezelése](#themes-and-plugins-sets): határozza meg, hogy mely témák és bővítmények legyenek automatikusan telepítve minden új webhelyen.
- [WordPress telepítése](#install-wordpress): néhány kattintással állítson be egy friss WordPress-telepítést.
- [Meglévő telepítések keresése](#scanning-importing-installations): észleli és importálja a kézzel telepített WordPress-webhelyeket.
- [Váltás táblázat/rács nézetre] (#grid-vs-table-view): megjeleníti a webhelyeket Rács (alapértelmezett) vagy táblázat módban.

### Telepítse a WordPress-t

A WordPress telepítése gyors és automatikus. Az OpenPanel mindenről gondoskodik – a WordPress letöltését a WordPress.org webhelyről, az adatbázis létrehozását, a domainnel való összekapcsolását és az új webhely konfigurálását.

1. Először adja meg **domainnevét**.
2. Nyissa meg a **Webhelykezelőt** az oldalsávon, és kattintson az **+ Új webhely** elemre.
3. Válassza az I**WordPress telepítése** lehetőséget.

![new_site_popup.png](/img/panel/v2/wpinstall.png)

Ezután töltse ki az űrlapot:

- Webhely neve
- Webhely leírása (opcionális)
- Domain név (opcionálisan almappa)
- Admin felhasználónév
- Admin jelszó
- WordPress verzió


Kattintson a **Telepítés indítása** gombra, és kész.

![wp_install.png](/img/panel/v2/wpinstall2.png)

📘 Olvassa el a teljes útmutatót: [A WordPress® telepítése OpenPanel segítségével](/docs/articles/websites/how-to-install-wordpress-with-openpanel/#install-wordpress-via-wp-manager)

### Beolvasási (importáló) telepítések

Ha már manuálisan telepítette a WordPress-t, importálhatja a WP Managerbe.
A rendszer átvizsgálja a tárhely fájljait a `wp-config.php` keresésére, és automatikusan hozzáadja a talált webhelyeket.

📘 Olvassa el a teljes útmutatót: [Hogyan lehet migrálni a WordPress® telepítést OpenPanelre] (/docs/articles/websites/how-to-upload-wordpress-website-to-openpanel/)

### Témák és beépülő modulok

Belefáradt, hogy minden alkalommal ugyanazt a beállítást telepítse?
Hozzon létre **témakészletet** és **bővítménykészletet**, amelyek automatikusan vonatkoznak az új WordPress-telepítésekre.

Például beállíthat egy alapértelmezett kombinációt, például:

- Elementor téma + gyermek téma
- Elementor plugin
- Klasszikus szerkesztő beépülő modul

Minden alkalommal, amikor új webhelyet telepít – bumm, az Ön által preferált beállításokkal készen áll.

📘 Olvassa el a teljes útmutatót: [WordPress Plugin & Theme Sets in OpenPanel](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)

### Webhelyadatok frissítése

Ha manuálisan módosította webhelyét (például frissítette a WordPress magját vagy megváltoztatta az adminisztrátori e-mail címet), kattintson az **Adatok frissítése** lehetőségre, hogy mindent szinkronizáljon a WP Managerrel.

### Rács és táblázatnézet

Webhelyeit **képernyőképekkel ellátott rácsban** vagy **egyszerű táblázat** nézetben tekintheti meg.
Bármikor válthat nézetet egy gombbal.

---

## Site Manager

![wp_manager_site.png](/img/panel/v2/wpmanage.png)


### Automatikus bejelentkezés a wp-adminba

A **Bejelentkezés rendszergazdaként** használatával egy kattintással biztonságosan hozzáférhet WordPress irányítópultjához – nincs szükség jelszóra.

![wp_manager_autologin](/img/panel/v2/wpautolog.png)

### Ideiglenes link

Tekintse meg webhelye előnézetét még azelőtt, hogy a domain csatlakozik, vagy az SSL készen áll.
Az ideiglenes hivatkozások 15 percig tartanak.

Kattintson az **Élő előnézet** elemre, hogy létrehozzon egyet:

![website_temporary_url_openpanel.gif](/img/panel/v2/wppreview.png)

### Képernyőkép

A webhely képernyőképei 24 óránként automatikusan frissülnek.
Hamarabb kell? Kattintson a frissítés ikonra a képernyőkép felett.

### Verziók

* **WordPress-verzió** – A WordPress-verziót a rendszer lekéri az adatbázisból, és egy AJAX-kéréssel ellenőrzi magának a webhelynek, biztosítva a megjelenített verzió pontosságát. Ha elérhető frissítés, egy jelvény jelenik meg a verziószám mellett.
* **PHP-verzió** – A PHP-verzió beolvasása a tartomány VirtualHost konfigurációs fájljából történik, garantálva, hogy a megjelenített verzió megegyezik a tartományhoz ténylegesen beállított verzióval.
* **MySQL/MariaDB verzió** – Megjeleníti, hogy a webhely MySQL-t vagy MariaDB-t használ-e, valamint a közvetlenül a termináltól kapott verziószámot.
* **Létrehozva** – Azt a dátumot és időt jelzi, amikor a webhelyet először hozzáadták a WP Managerhez.

![general](/img/panel/v2/general.png)

### Sebesség

A webhely teljesítményét a **Google PageSpeed ​​Insights** segítségével naponta nyomon követjük. Mind a mobil, mind az asztali eszközök esetében megtekintheti az ellenőrzési időt, valamint olyan kulcsfontosságú mutatókat, mint a **First Contentful Paint**, **Speed ​​Index** és **Time to Interactive**.

Ezenkívül [hozzáadhatja saját PageSpeed ​​Insights API-kulcsát] (/docs/articles/websites/google-pagespeed-insights-api-key/#adding-the-api-key-in-openpanel) az adatgyűjtés testreszabásához.

![sebesség](/img/panel/v2/speed.png)

### Gyorsítótár

A gyorsítótár modul megjeleníti az aktuális [wp-gyorsítótár típusát](https://developer.wordpress.org/cli/commands/cache/type/) a webhelyen, és lehetőséget ad a gyorsítótár törlésére.

![wp_cache](/img/panel/v2/wp_cache.png)

### Tűzfal

Ha a CorazaWAF engedélyezve van a szerveren, és fiókja hozzáfér a WAF funkcióhoz, akkor megjelenik egy *Tűzfal* widget, amely megjeleníti a domain aktuális állapotát, a módosítási lehetőséget, valamint az elmúlt órában elutasított/kihívott kérések számát.

![wp_waf](/img/panel/v2/wp_waf.png)


### Áttekintés

Az *Áttekintés* lapon megtekintheti:
- Fájlok: Mappa elérési útja és Mappa mérete
- Adatbázis: méret, gazdagép, név, táblázat előtag, felhasználó, jelszó és hivatkozás a phpMyAdmin megnyitásához

![áttekintés](/img/panel/v2/overview.png)

### Opciók

Az *Opciók* lap megjeleníti a WordPress aktuális beállításait, és lehetővé teszi azok módosítását.

Elérhető opciók:

- Webhely URL-je
- Kezdőlap URL-je
- Webhely neve
- Blog leírása
- Rendszergazda e-mail
- Új felhasználó regisztrációjának engedélyezése
- SEO láthatóság engedélyezése
- Pingback engedélyezése

![options](/img/panel/v2/options.png)

### Karbantartási mód

A karbantartási mód engedélyezése vagy letiltása közvetlenül a WP Manager alkalmazásból.
A karbantartás.php fájlt akár közvetlenül a panelről is szerkesztheti.

![wp_manager_maintenance](/img/panel/v2/wpmaint.png)

### Biztonság

Tartsa biztonságban webhelyét a beépített biztonsági eszközökkel.

Innen a következőket teheti:
- Keverje össze a WordPress sókat
- Ellenőrizze az alapvető fájl integritását
- Szükség esetén telepítse újra a WordPress magot

![wp_manager_security.png](/img/panel/v2/wpsec.png)

### Frissítések

Szabályozhatja, hogy a WordPress hogyan kezelje a mag, a beépülő modulok és a témák frissítéseit.
Alapértelmezés szerint csak a kisebb alapvető frissítések vannak automatikusan engedélyezve.

![wp_manager_site_edit_2.png](/img/panel/v2/wpupdate.png)

Ha elérhető egy újabb WordPress alapverzió, megjelenik a „Kattintson a WordPress mag frissítéséhez” gomb, amelyre kattintva végrehajtja a WordPress frissítését a legújabb elérhető verzióra.

### Hibakeresés

Kapcsolja be a WordPress beépített hibakereső eszközeit (WP_DEBUG, WP_DEBUG_LOG stb.) közvetlenül a WP Managerből.

Ezek kiválóan alkalmasak tesztelési vagy fejlesztési helyszínekre – termelésre nem ajánlott.
A részletekért tekintse meg a [Debugging in WordPress] (https://wordpress.org/documentation/article/debugging-in-wordpress/) részt, ahol további információra van szüksége ezekről a lehetőségekről.

![wp_manager_site_edit_3.png](/img/panel/v2/wpdebug.png)

### Biztonsági mentések

Bármikor készíthet és állíthat vissza biztonsági másolatot – fájlokat, adatbázisokat vagy mindkettőt.

Biztonsági másolat létrehozása:
- Válassza ki, hogy miről szeretne biztonsági másolatot készíteni (fájlok, adatbázis vagy mindkettő).
- Kattintson a *Biztonsági másolat létrehozása** lehetőségre.

![wp_manager_site_backup_1.png](/img/panel/v2/wpbackup.png)

Biztonsági másolat visszaállítása:
A visszaállításhoz kattintson a Visszaállítás gombra, válassza ki a biztonsági mentés dátumát, és erősítse meg.

### Eltávolítás

Szeretné leállítani a webhely kezelését a WP Managerben (törlés nélkül)?

Használja a **Leválasztás** funkciót – a fájlok és az adatbázis érintetlenek maradnak.

![detach](/img/panel/v2/detach.png)

Egy webhely – fájlok, adatbázis és minden – teljes eltávolításához kattintson az **Eltávolítás** lehetőségre, majd erősítse meg.

![uninstall](/img/panel/v2/uninstall.png)

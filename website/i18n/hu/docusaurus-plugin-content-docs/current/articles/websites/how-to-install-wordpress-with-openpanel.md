# A WordPress® telepítése az OpenPanel segítségével

A [WordPress](https://wordpress.org/)® egy hatékony, webalapú tartalomkezelő rendszer (CMS), amely megkönnyíti webhelyek és blogok készítését. Ez az útmutató két módszert mutat be a WordPress OpenPanel-fiókra történő telepítéséhez: **automatikus telepítés a WP Manageren keresztül** és **kézi telepítés**.

A WordPress telepítése előtt győződjön meg arról, hogy hozzáadott egy domaint, ahol a WordPress webhelyet tárolni fogja.

Domain hozzáadása:
**OpenPanel > Domains > New Domain hozzáadása**

![újDomain.png](/img/panel/v2/wpgDomain.png)

## Telepítse a WordPress-t a WP Manageren keresztül

Az OpenPanel Site Manager segítségével néhány kattintással telepítheti a WordPress-t.

1. **Nyissa meg a Webhelykezelőt**, és nyomja meg az **+Új webhely** gombot.
![SiteManager_1.png](/img/panel/v2/wpgSitemanager1.png)
2. Kattintson a **WordPress telepítése** lehetőségre.
![SiteManager_2.png](/img/panel/v2/wpgSitemanager2.png)
3. Töltse ki a **Webhely neve** mezőt, opcionálisan adjon hozzá egy **Webhely leírását**, és válassza ki a domainjét.
Kattintson a **Telepítés indítása** gombra a kezdéshez.
![SiteManager_3.png](/img/panel/v2/wpgSitemanager3.png)
5. A telepítés befejezése után átirányítjuk a **WP Manager** oldalra, ahol kezelheti az összes telepített WordPress webhelyet.
![SiteManager_4.png](/img/panel/v2/wpgSitemanager4.png)

## A WordPress manuális telepítése

A hivatalos WordPress telepítési archívum letöltésének legegyszerűbb módja az OpenPanel Fájlkezelő eszköze.

Először nyissa meg a domain mappáját, ahová új WordPress-példányt telepít.

![Manually_1.png](/img/panel/v2/wpgManual1.png)

Amikor a domainek könyvtárában tartózkodik, nyomja meg a Feltöltés gombot a jobb felső sarokban.

![Manually_2.png](/img/panel/v2/wpgManual2.png)

A Feltöltés oldalon nyomja meg a "Letöltés helyett az URL-ről" gombot.

![Manually_3.png](/img/panel/v2/wpgManual3.png)

Illessze be a https://wordpress.org/latest.zip hivatkozást az URL mezőbe, és nyomja meg a „Letöltés” ​​gombot.

![Manually_4.png](/img/panel/v2/wpgManual4.png)

![Manually_5.png](/img/panel/v2/wpgManual5.png)

Amikor a letöltés befejeződött, térjen vissza a Fájlkezelőbe, és csomagolja ki az archívumot.

![Manually_6.png](/img/panel/v2/wpgManual6.png)

Nyissa meg a "wordpress" mappát, válassza ki az Összes fájlt a "Select All" gombbal, és helyezze át őket az új domain könyvtárába az "Áthelyezés" gombbal.

![Manually_7.png](/img/panel/v2/wpgManual7.png)

![Manually_8.png](/img/panel/v2/wpgManual8.png)

![Manually_9.png](/img/panel/v2/wpgManual9.png)

![Manually_10.png](/img/panel/v2/wpgManual10.png)

Hozzon létre új adatbázist és adatbázis-felhasználót az Adatbázisvarázsló eszközzel, további információ az Adatbázisvarázsló útmutatójában -> https://openpanel.com/docs/panel/mysql/wizard/.

![Manually_11.png](/img/panel/v2/wpgManual11.png)

Kezdje el szerkeszteni a wp-config-sample.php fájlt úgy, hogy kijelöli és megnyomja a "Szerkesztés" gombot.

![Manually_12.png](/img/panel/v2/wpgManual12.png)

Cserélje le a DB_NAME, DB_USER és DB_PASSWORD értékeit az Adatbázisvarázslóban beállított értékekkel, állítsa a DB_HOST értékét „mysql”-re, és mentse a változtatásokat a jobb felső sarokban.

![Manually_config.png](/img/panel/v2/wpgManualFinal.png)

Végül nevezze át a wp-config-sample.php fájlt wp-config.php-re a Fájlkezelő Átnevezés gombjával.

![Manually_configR.png](/img/panel/v2/wpgManualRename.png)

![Manually_configR2.png](/img/panel/v2/wpgManualRename2.png)

Az új WordPress-példányod elkészült!

Fejezze be a telepítést úgy, hogy a böngészőn keresztül hozzáfér a domainjéhez, és végrehajtja a WordPress webes telepítővarázslót.


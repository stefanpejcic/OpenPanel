# Branding & White-Label

Az OpenPanelben minden moduláris, és könnyen módosítható vagy letiltható anélkül, hogy a többi funkciót megzavarná.

Az OpenPanel testreszabásához a következő lehetőségek állnak rendelkezésére:

- [Személyre szabott üzenet megjelenítése felhasználónként](#personalized-messages)
- [Funkciók és oldalak engedélyezése/letiltása az OpenPanel felületről](#enabledisable-features)
- [Egyéni ikonok hozzáadása az Irányítópult oldalon](/docs/articles/dev-experience/add-custom-icons-in-openpanel-dashboard)
- [Alapértelmezett, felfüggesztett felhasználó és felfüggesztett domain oldalak testreszabása](@customize-templates)
- [A felület honosítása](#localize-the-interface)
- [Egyéni márkajelzés beállítása](#set-custom-branding)
- [Egyéni színséma beállítása](#set-a-custom-color-scheme)
- [Cserélje ki az útmutató cikkeket tudásbázisával](#replace-how-to-articles-with-your-knowledge-base)
- [Bármelyik oldal testreszabása](#edit-any-page-template)
- [A bejelentkezési oldal testreszabása](#customize-login-page)
- [Egyéni CSS- vagy JS-kód hozzáadása a felülethez](#create-custom-pages)
- [Egyéni modul létrehozása az OpenPanelhez](#create-custom-module)
- [Selfhosted-temporary-links-api/] (/docs/articles/dev-experience/selfhosted-temporary-links-api/)
- [Selfhosted-screenshots-api/](/docs/articles/dev-experience/selfhosted-screenshots-api/)



## Személyre szabott üzenetek

A rendszergazdák az **OpenAdmin > Felhasználók** oldalon beállíthatnak egy egyéni üzenetet, amely minden OpenPanel felhasználó számára megjelenjen.

![egyéni img](https://i.postimg.cc/9CCgHGG2/2025-06-11-12-26.png)

## Funkciók engedélyezése/letiltása

Az adminisztrátorok engedélyezhetik vagy letilthatják az OpenPanel felület egyes funkcióit (oldalait) tervenként vagy felhasználónként.

Az engedélyezést követően a funkció azonnal elérhetővé válik minden felhasználó számára, és megjelenik az OpenPanel felület oldalsávjában, a keresési eredményekben és az irányítópult ikonjain.

## Előre telepített szolgáltatások beállítása

Az OpenPanel a docker-összeállítási fájlokat használja minden egyes felhasználó alapjaként. A fájlokat összeállító docker-képek alapján különböző szolgáltatások állíthatók be tervenként/felhasználónként.


## Lokalizálja a felületet

Az OpenPanel honosításra kész, és könnyen lefordítható bármilyen nyelvre.

Az OpenPanel az EN nyelvi beállítással együtt kerül szállításra, [a rendszergazda további területi beállításokat is telepíthet](https://dev.openpanel.com/localization.html#How-to-translate).


## Egyéni márkajelzés beállítása

Az egyéni márkanév és logó az [OpenAdmin > Beállítások > OpenPanel](/docs/admin/settings/openpanel/#branding) oldalon állítható be.

Az OpenPanel oldalsávján és a bejelentkezési oldalakon látható egyéni név beállításához adja meg a kívánt nevet a "Márkanév" opcióban. Alternatív megoldásként, ha helyette logót szeretne megjeleníteni, adja meg az URL-t az „Embléma képe” mezőben, és mentse el a változtatásokat.


## Sablonok testreszabása

Testreszabhatja a felhasználók számára megjelenített összes sablont:

- [Domain VHost sablon] (/docs/admin/services/nginx/#domain-vhost-template)
- [Alapértelmezett céloldal](/docs/admin/services/nginx/#default-landing-page)
- [Felfüggesztett felhasználói sablon](/docs/admin/services/nginx/#suspended-user-template)
- [Felfüggesztett domain sablon] (/docs/admin/services/nginx/#suspended-domain-template)
- [Hibaoldalak](/docs/admin/services/nginx/#error-pages)

## Hozzon létre OpenPanel modult

Egyéni modul (bővítmény) létrehozásához az OpenPanelhez kövesse ezt az útmutatót: [Példamodul](https://dev.openpanel.com/modules/#Example-Module)

## Állítson be egyéni színsémát

Egyéni színséma beállításához az OpenPanel felülethez, szerkessze az "/etc/openpanel/openpanel/custom_code/custom.css" fájlt, és állítsa be a kívánt színsémát.

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

Állítsa be az egyéni css kódot, mentse és indítsa újra az openpanelt a módosítások alkalmazásához:

```bash
cd /root && docker compose up -d openpanel
```

Példa:

![custom_css_code](https://i.postimg.cc/YprhHZhg/2024-06-18-15-04.png)





## Cserélje ki az útmutató cikkeket tudásbázisával

Az [OpenPanel Dashboard oldal](/docs/panel/dashboard) megjeleníti a [How-to articles](/docs/panel/dashboard/#how-to-guides) az OpenPanel Dokumentumokból, de ezek módosíthatók, hogy helyette a tudásbázis cikkeit jelenítse meg.

Szerkessze az "/etc/openpanel/openpanel/conf/knowledge_base_articles.json" fájlt, és állítsa be a hivatkozásokat:

```json
{
    "how_to_topics": [
        {"title": "How to install WordPress", "link": "https://openpanel.com/docs/panel/applications/wordpress#install-wordpress"},
        {"title": "Publishing a Python Application", "link": "https://openpanel.com/docs/panel/applications/pm2#python-applications"},
        {"title": "How to edit Nginx / Apache configuration", "link": "https://openpanel.com/docs/panel/advanced/server_settings#nginx--apache-settings"},
        {"title": "How to create a new MySQL database", "link": "https://openpanel.com/docs/panel/databases/#create-a-mysql-database"},
        {"title": "How to add a Cronjob", "link": "https://openpanel.com/docs/panel/advanced/cronjobs#add-a-cronjob"},
        {"title": "How to change server TimeZone", "link": "https://openpanel.com/docs/panel/advanced/server_settings#server-time"}
    ],
    "knowledge_base_link": "https://openpanel.com/docs/panel/intro/?source=openpanel_server"
}
```


## Szerkesszen bármely oldalsablont

Minden OpenPanel sablonkód nyitott, és egyszerűen szerkeszthető a HTML-kód szerkesztésével.

A sablonok egy "openpanel" nevű Docker-tárolóban találhatók, ezért először meg kell találnia azt a sablont, amely tartalmazza a szerkeszteni kívánt kódot.

Például az oldalsáv szerkesztéséhez és az OpenPanel logó elrejtéséhez kövesse az alábbi lépéseket:

1. Hozzon létre egy új mappát/fájlt helyileg a módosított kódhoz.
``` bash
mkdir /root/custom_template/
   ```
2. Másolja ki a meglévő sablonkódot.
``` bash
docker cp openpanel:/usr/local/panel/templates/partials/sidebar.html /root/custom_template/sidebar.html
   ```
3. Szerkessze a kódot.

4. Állítsa be az OpenPanel-t a sablon használatához.
Szerkessze a `/root/docker-compose.yml` fájlt, és állítsa be, hogy a fájl felülírja az eredeti sablont:
``` bash
nano /root/docker-compose.yml
   ```
és az [openpanel > kötetek] alatti fájlban (https://github.com/stefanpejcic/openpanel-configuration/blob/180c781bfb7122c354fd339fbee43c1ce6ec017f/docker/compose/new-docker-compose:yml)
``` bash
- /root/custom_theme/sidebar.html:/usr/local/panel/templates/partials/sidebar.html
   ```
6. Indítsa újra az OpenPanel alkalmazást az új sablon alkalmazásához.
``` bash
cd /root && docker compose up -d openpanel
   ```


## A bejelentkezési oldal testreszabása


Az OpenPanel bejelentkezési oldal sablonkódja a `/usr/local/panel/templates/user/login.html` címen található a docker-tárolóban.

A bejelentkezési oldal szerkesztéséhez:

1. Hozzon létre egy új mappát/fájlt helyileg a módosított kódhoz.
``` bash
mkdir /root/custom_template/
   ```
2. Másolja ki a meglévő sablonkódot.
``` bash
docker cp openpanel:/usr/local/panel/templates/user/login.html /root/custom_template/login.html
   ```
3. Szerkessze a kódot.

4. Állítsa be az OpenPanel-t a sablon használatához.
Szerkessze a `/root/docker-compose.yml` fájlt, és állítsa be, hogy a fájl felülírja az eredeti sablont:
``` bash
nano /root/docker-compose.yml
   ```
és az [openpanel > kötetek] alatti fájlban (https://github.com/stefanpejcic/openpanel-configuration/blob/180c781bfb7122c354fd339fbee43c1ce6ec017f/docker/compose/new-docker-compose:yml)
``` bash
- /root/custom_theme/login.html:/usr/local/panel/templates/user/login.html
   ```
6. Indítsa újra az OpenPanel alkalmazást az új bejelentkezési sablon alkalmazásához.
``` bash
cd /root && docker compose up -d openpanel
   ```


## Adjon hozzá egyéni CSS- vagy JS-kódot

Ha egyéni CSS-kódot szeretne hozzáadni az OpenPanel felülethez, szerkessze a `/etc/openpanel/openpanel/custom_code/custom.css` fájlt:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.css
```

Ha egyéni JavaScript kódot szeretne hozzáadni az OpenPanel felülethez, szerkessze a `/etc/openpanel/openpanel/custom_code/custom.js` fájlt:

```bash
nano /etc/openpanel/openpanel/custom_code/custom.js
```

Ha egyéni kódot szeretne beszúrni az OpenPanel felület `<head>` címkéjébe, módosítsa az `/etc/openpanel/openpanel/custom_code/in_header.html` címen található fájl tartalmát, és helyezze bele az egyéni kódot:

```bash
nano /etc/openpanel/openpanel/custom_code/in_header.html
```

Ha egyéni kódot szeretne beszúrni az OpenPanel felület `<láb>` címkéjébe, módosítsa az `/etc/openpanel/openpanel/custom_code/in_footer.html` címen található fájl tartalmát, és helyezze bele az egyéni kódot:

```bash
nano /etc/openpanel/openpanel/custom_code/in_footer.html
```


## E-mail sablonok testreszabása

A felhasználói értesítésekhez használt e-mail sablon módosítása: https://community.openpanel.org/d/189-customizing-the-email-template-used-for-user-notifications


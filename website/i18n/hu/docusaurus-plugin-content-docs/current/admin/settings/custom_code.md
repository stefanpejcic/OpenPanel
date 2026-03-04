---
sidebar_position: 6
---

# Egyéni kód

Az OpenPanel modularitásáról híres, így teljes irányítást biztosít a platform igényeinek megfelelő testreszabásához.

Az OpenAdmin > Beállítások > Egyéni kód menüponton keresztül elérhető Egyéni kód szakasz lehetővé teszi az OpenPanel felhasználói felület viselkedését és megjelenését kiterjesztő vagy módosító egyéni kód beszúrását és kezelését.

## Egyéni CSS
Szúrja be saját CSS-stílusait, amelyeket az OpenPanel felhasználói felületének minden oldalán alkalmazni fog.

[Dokumentáció megtekintése](https://dev.openpanel.com/customize.html#Custom-CSS)

## Egyéni JS
Adjon hozzá egyéni JavaScriptet az összes oldalhoz. Ez ideális a felhasználói felület funkcióinak bővítéséhez, widgetek hozzáadásához vagy harmadik féltől származó eszközök integrálásához.

[Dokumentáció megtekintése](https://dev.openpanel.com/customize.html#Custom-JS)

## Kód a fejlécben
Szúrjon be egyéni kódot közvetlenül minden oldal head címkéjébe. Hasznos metacímkékhez, elemző szkriptekhez és globális beállításokhoz.

[Dokumentáció megtekintése](https://dev.openpanel.com/customize.html#Code-in-Header)

## Kód a láblécben
Szúrjon be egyéni kódot közvetlenül az összes oldal láblécébe. Ezt általában a szkriptek, az elemzések vagy a késleltetett JavaScript nyomon követésére használják.

[Dokumentáció megtekintése](https://dev.openpanel.com/customize.html#Code-in-Footer)



## Útmutatók
Az *OpenPanel > Irányítópult* oldalon megjelenő Tudásbázis-cikkek szerkesztése.

Alapértelmezett:
```json
{
    "how_to_topics": [
        {"title": "How to install WordPress", "link": "https://openpanel.com/docs/panel/applications/wordpress#install-wordpress"},
        {"title": "How to enable REDIS Caching", "link": "https://openpanel.com/docs/panel/caching/Redis/#connect-to-redis"},
        {"title": "How to create DNS records", "link": "https://openpanel.com/docs/panel/domains/dns/#create-record"},
        {"title": "How to create a new MySQL database", "link": "https://openpanel.com/docs/panel/databases/#create-a-mysql-database"},
        {"title": "How to add a Cron Job", "link": "https://openpanel.com/docs/panel/advanced/cronjobs#add-a-cronjob"},
        {"title": "How to add a custom SSL Certificate", "link": "https://openpanel.com/docs/panel/domains/ssl/#custom-ssl"}
    ],
    "knowledge_base_link": "https://openpanel.com/docs/panel/intro/?source=openpanel_server"
}
```

## Egyéni szakasz

Az **OpenPanel** *Irányítópultjához* hozzáadhat egy **egyéni szakaszt** ikonalapú elemekkel.

az egyéni szakasz a következő mezőket támogatja:

* **`section_title`** *(karakterlánc)*:
Az egyéni panel tetején megjelenő cím.

* **`section_position`** *(karakterlánc)*:
Meghatározza, hogy az Ön szakasza hol jelenjen meg a többi beépített szakaszhoz képest.
Elfogadható értékek:

* `fájlok_előtt`
* "a_domainek előtt".
* `mysql_előtt`
* `before_postgresql`
* "a_jelentkezések előtt".
* "e-mailek előtt".
* `előtt_gyorsítótár`
* `fore_php`
* "kikötő előtt".
* `előre_haladó`
* "számla előtt".

* **`elemek`** *(objektumtömb)*:
A kattintható elemek listája ikonokkal ellátott kártyaként. Minden elem rendelkezik:

* **`címke`** *(karakterlánc)* – A kártyán látható szöveg.
* **`icon`** *(karakterlánc)* – A [Bootstrap Icons] (https://icons.getbootstrap.com/) ikonosztálya. Példa: "bi-bi-person-fill-gear".
* **`url`** *(karakterlánc)* – A hivatkozás, amelyre az elemre kattintva navigálhat.

Példa:
```json
{
  "section_title": "Billing Account",
  "section_position": "before_domains",
  "items": [
    {
      "label": "Manage Profile",
      "icon": "bi bi-person-fill-gear",
      "url": "https://panel.hostio.rs/clientarea.php?action=details"
    },
    {
      "label": "Manage Billing Information",
      "icon": "bi bi-credit-card",
      "url": "https://panel.unlimited.rs/clientarea.php?action=details"
    },
    {
      "label": "View Email History",
      "icon": "bi bi-envelope-open",
      "url": "https://panel.unlimited.rs/clientarea.php?action=emails"
    },
    {
      "label": "News & Announcements",
      "icon": "bi bi-megaphone-fill",
      "url": "https://panel.unlimited.rs/index.php?rp=/announcements"
    },
    {
      "label": "Knowledgebase",
      "icon": "bi bi-book-half",
      "url": "https://panel.unlimited.rs/index.php?rp=/knowledgebase"
    },
    {
      "label": "Server Status",
      "icon": "bi bi-hdd-network",
      "url": "https://panel.unlimited.rs/serverstatus.php"
    },
    {
      "label": "Invoices",
      "icon": "bi bi-receipt",
      "url": "https://panel.unlimited.rs/clientarea.php?action=invoices"
    },
    {
      "label": "Support Tickets",
      "icon": "bi bi-life-preserver",
      "url": "https://panel.unlimited.rs/supporttickets.php"
    },
    {
      "label": "Open Ticket",
      "icon": "bi bi-journal-plus",
      "url": "https://panel.unlimited.rs/submitticket.php"
    },
    {
      "label": "Register New Domain",
      "icon": "bi bi-globe",
      "url": "https://panel.unlimited.rs/cart.php?a=add&domain=register"
    },
    {
      "label": "Transfer Domain",
      "icon": "bi bi-arrow-repeat",
      "url": "https://panel.unlimited.rs/cart.php?a=add&domain=transfer"
    }
  ]
}

```

## PageSpeed ​​API kulcs

Ha be van állítva, ezt az API-kulcsot a rendszer az adatok lekérésére használja a Google PageSpeed ​​Insights szolgáltatásból.
---

## WordPress bővítménykészlet

Sorolja fel azokat a WordPress-bővítményeket, amelyeket automatikusan telepíteni szeretne az összes új WordPress-webhelyre.
Soronként egy elemet írjon be. Támogatott formátumok:

* **wp\_org\_slug** – a beépülő modul slugja a WordPress.org beépülő modul oldaláról
* **URL** – közvetlen link egy online tárolt `.zip` beépülő fájlhoz

## WordPress témakészlet

Sorolja fel az összes új WordPress-webhelyen automatikusan telepítendő WordPress-témákat.
Soronként egy elemet írjon be. Támogatott formátumok:

* **wp\_org\_slug** – a téma slug a WordPress.org témák oldaláról
* **URL** – közvetlen link egy online tárolt `.zip` témafájlra

## Tiltott felhasználónevek

A nem használható felhasználónevek listája.

## Korlátozott domainek

A nem használható domainek listája.

## Frissítés után
Határozzon meg egyéni bash parancsokat, amelyek automatikusan futnak minden OpenPanel frissítés után. Ideális testreszabások visszaállításához vagy automatizálási szkriptek indításához.

[Dokumentáció megtekintése](https://dev.openpanel.com/customize.html#After-update)

Ez a hatékony testreszabási réteg segít abban, hogy az OpenPanel zökkenőmentesen illeszkedjen a környezetébe.

![openadmin egyéni kód](/img/admin/custom_code.png)

---
sidebar_position: 1
---

# Felhasználói csomagok

A tárhelycsomagok korlátokat szabnak a felhasználók számára.

## Sorolja fel a hosting terveket

<Tabs>
<TabItem value="openadmin-plan-list" label="OpenAdminnal" alapértelmezett>


A meglévő tervek listázásához lépjen a Tervek oldalra:

![openadmin plans](/img/admin/tremor/plans_list.png)


| Mező | Leírás |
| ------------------- | -------------------------------------------------------------------------- |
| **Terv neve** | Megjelenítési név, amelyet a felhasználók látni fognak az OpenPanel irányítópultjaikon.            |
| **Memória** | Fizikai memória (RAM) GB-ban, amely a felhasználó számára van lefoglalva ebben a tárhelycsomagban.     |
| **CPU** | A felhasználónak szánt CPU magok száma ezen a tárhelycsomagon.             |
| **Lemez** | Az összes felhasználói fájl számára lefoglalt lemezterület GB-ban.           |
| **Inodes** | Korlátozza a felhasználó számára engedélyezett fájlok teljes számát.   |
| **Port sebesség** | Maximális postázási sebesség a felhasználók számára Mbit/s-ban.     |
| **Domainek** | A csomagban felhasználónként engedélyezett domain nevek teljes száma.                  |
| **Webhelyek** | A csomagban szereplő összes webhely (WordPress, NodeJS, Python) felhasználónként.   |
| **Adatbázisok** | A tervben felhasználónként engedélyezett MySQL/MariaDB adatbázisok teljes száma.              |
| **E-mail fiókok** | A felhasználó által a tervben létrehozható e-mail fiókok teljes száma.              |
| **Postafiókkvóta** | Maximális postafiókméret az e-mail fiókokhoz, amelyeket a felhasználó beállíthat ebben a tervben.              |
| **FTP-fiókok** | A felhasználó által a tervben létrehozható ftp-fiókok teljes száma.             |
| **Funkciókészlet** | A [Feature Sets](/docs/admin/settings/openpanel/#enable-features) meghatározza, hogy a felhasználók mely oldalakat érhetik el az OpenPanel felületéről.               |



</TabItem>
<TabItem value="CLI-plan-list" label="OpenCLI-vel">

Az összes futó tárhelycsomag (terv) felsorolásához:

```bash
opencli plan-list
```

Példa kimenet:
```bash
[root@fajlovi ~]# opencli plan-list
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+
| id | name           | description            | domains_limit | websites_limit | email_limit | ftp_limit | disk_limit | inodes_limit | db_limit | cpu  | ram  | bandwidth | feature_set | max_email_quota |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+
|  1 | Standard plan  | Small plan for testing |             0 |             10 |           0 |         0 | 5 GB       |      1000000 |        0 | 2    | 2g   |        10 | basic       | 0               |
|  2 | Developer Plus | 4 cores, 6G ram        |             0 |             10 |           0 |         0 | 20 GB      |      2500000 |        0 | 4    | 6g   |       100 | default     | 0               |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+

```

Az adatokat JSON-ként is formázhatja:

```bash
opencli plan-list --json
```

</TabItem>
</Tabs>

## Készítsen tervet

<Tabs>
<TabItem value="openadmin-plan-new" label="OpenAdminnal" alapértelmezett>

Új tárhelycsomag létrehozásához kattintson az **'Új létrehozása'** gombra, és állítsa be a kívánt korlátokat:

![openadmin plans create](/img/admin/tremor/plans_create.png)


* **Név** – Bármilyen karaktert tartalmazhat.
* **Leírás** – Belső megjegyzés az adminisztrátoroknak, csak az OpenAdminban látható.
* **Lemez** – Tárhely GB-ban. Használja a „0”-t a korlátlanul.
* **Inodes** – Inode-ok száma. Használja a „0”-t a korlátlanul.
* **CPU** – Az összes felhasználói szolgáltatáshoz kiosztott CPU magok száma. Állítsa 0-ra a korlátlan használathoz. Nem haladhatja meg a szervermagok teljes számát.
* **Memória** – Az összes felhasználói szolgáltatáshoz lefoglalt fizikai memória mennyisége GB-ban. Állítsa 0-ra a korlátlan használathoz. Nem haladhatja meg a szerver teljes memóriáját.
* **Port sebesség** – A felhasználói szolgáltatások maximális sebessége Mbit/s-ban *(Elavult és nem kötelező)*.
* **Adatbázisok** – Az adatbázisok maximális száma (MySQL, MariaDB, PostgreSQL). Használja a „0”-t a korlátlanul.
* **Webhelyek** – A webhelyek maximális száma a Site Managerben (WordPress, WebsiteBuilder, NodeJS/Python). Használja a „0”-t a korlátlanul.
* **FTP-fiókok** – FTP-alszámlák maximális száma. Használja a „0”-t a korlátlanul.
* **E-mail fiókok** – Az e-mail alfiókok maximális száma. Használja a „0”-t a korlátlanul.
* **Postafiókkvóta** – Maximális postafiókméret az e-mail fiókokhoz, amelyeket a felhasználó beállíthat ebben a tervben.
* **Feature Set** – Az OpenPanel felhasználói felületen elérhető szolgáltatásokat meghatározó szolgáltatáskészlet neve.


</TabItem>
<TabItem value="CLI-plan-new" label="OpenCLI-vel">
    
Új terv létrehozásához futtassa a következő parancsot:

```bash
opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT>
```

Példa:
```bash
opencli plan-create name=New Plan description=This is a new plan emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G
```

</TabItem>
</Tabs>


## Terv módosítása

A tervkorlátok módosításához kattintson a **Szerkesztés** gombra az **OpenAdmin > Felhasználói csomagok** oldalon, és állítsa be az új korlátokat.

![openadmin plans edit](/img/admin/tremor/plans_edit_1.png)

![openadmin plans edit limits](/img/admin/tremor/plans_edit_2.png)


Az új limitek azonnal alkalmazásra kerülnek a csomagot használó összes fiókra.

## A tervben szereplő felhasználók listázása

<Tabs>
<TabItem value="openadmin-plan-usage" label="OpenAdminnal" alapértelmezett>

Az összes olyan felhasználó megtekintéséhez, aki jelenleg egy tárhelycsomagot használ, egyszerűen rendezze a felhasználók táblázatát a csomagnév szerint, vagy írja be a csomag nevét a keresőmezőbe.

![openadmin plans usage](/img/admin/tremor/plans_usage_1.png)

![openadmin plans usage](/img/admin/tremor/plans_usage_2.png)

</TabItem>
<TabItem value="CLI-plan-usage" label="OpenCLI-vel">
    
Sorolja fel az összes olyan felhasználót, aki jelenleg egy csomagot használ:

```bash
opencli plan-usage
```

Példa:
```bash
[root@fajlovi ~]# opencli plan-usage 'Standard plan'
+----+----------------------------------+----------------------+---------------+---------------------+
| id | username                         | email                | plan_name     | registered_date     |
+----+----------------------------------+----------------------+---------------+---------------------+
|  3 | forums                           | stefan@openpanel.com | Standard plan | 2025-05-08 19:25:47 |
| 19 | radovan                          | radovan@jecmenica.rs | Standard plan | 2025-05-29 07:47:15 |
+----+----------------------------------+----------------------+---------------+---------------------+
```

Az adatokat JSON-ként is formázhatja:

```bash
opencli plan-usage --json
```
</TabItem>
</Tabs>

## Terv törlése

<Tabs>
<TabItem value="openadmin-plan-delete" label="OpenAdminnal" alapértelmezett>
    
Tárhelycsomag törléséhez kattintson a **Törlés** linkre a kívánt csomag mellett.

![openadmin plans delete](/img/admin/tremor/plans_delete.png)


</TabItem>
<TabItem value="CLI-plan-delete" label="OpenCLI-vel">

Tárhelycsomag törlése:

```bash
opencli plan-delete <PLAN_NAME> 
```

Példa:
```bash
opencli plan-delete 'Standard plan'
```
</TabItem>
</Tabs>

Megjegyzés: Egy csomag nem törölhető, ha hozzá vannak rendelve felhasználók.

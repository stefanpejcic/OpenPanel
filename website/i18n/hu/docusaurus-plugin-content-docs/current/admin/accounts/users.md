---
sidebar_position: 1
---

# Felhasználók

Az OpenPanel egyetlen **Felhasználó** nevű felhasználói szerepkörrel rendelkezik, amely csak a Docker-tárolóját tudja kezelni, és örökli a rendszergazda által megadott beállításokat.


## Felhasználók listázása


<Tabs>
<TabItem value="openadmin-users" label="OpenAdmin" alapértelmezett>
  
Az összes OpenPanel-felhasználó eléréséhez lépjen a Felhasználók oldalra.
  
A Felhasználók oldalon egy táblázat látható a felhasználói adatokkal és a kezelésükhöz szükséges gombokkal.
  
![openadmin felhasználók oldala](/img/admin/openadmin_users_list.gif)
  
További oszlopok jeleníthetők meg az „Oszlopok megjelenítése” gombbal.

A felfüggesztett felhasználók piros színnel vannak kiemelve, és a felfüggesztett felhasználókkal semmilyen művelet nem hajtható végre.

</TabItem>
<TabItem value="CLI-users" label="OpenCLI">

Az összes felhasználó listázásához használja a következő parancsot:

```bash
opencli user-list
```

Példa kimenet:
```bash
opencli user-list
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
| id | username                         | email                | plan_name      | server           | owner | registered_date     |
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
|  3 | forums                           | stefan@openpanel.com | Standard plan  | forums           | NULL  | 2025-05-08 19:25:47 |
|  7 | pcx3                             | stefan@pejcic.rs     | Developer Plus | pcx3             | NULL  | 2025-05-09 12:26:20 |
|  9 | openpanelwebsite                 | info@openpanel.com   | Standard plan  | openpanelwebsite | NULL  | 2025-05-09 14:47:27 |
| 19 | SUSPENDED_20250529173435_radovan | radovan@jecmenica.rs | Standard plan  | radovan          | NULL  | 2025-05-29 07:47:15 |
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
```

Az adatokat JSON-ként is formázhatja:

```bash
opencli user-list --json
```
</TabItem>
<TabItem value="API-users" label="API">

Az összes felhasználó listázásához használja a következő API-végpontot:

```bash
curl -X GET http://PANEL:2087/api/users -H "Authorization: Bearer JWT_TOKEN_HERE"
```

</TabItem>
</Tabs>


## Felhasználók létrehozása


<Tabs>
<TabItem value="openadmin-users-new" label="OpenAdmin" alapértelmezett>

Új felhasználó létrehozásához kattintson az 'Új felhasználó' gombra a Felhasználók oldalon. Megjelenik egy új rész egy űrlappal, ahol beállíthatja az e-mail címet, felhasználónevet, erős jelszót generálhat, és tárhely-tervet rendelhet a felhasználóhoz.

![új felhasználó hozzáadása openadmin](/img/admin/2025-06-09_08-20.png)

</TabItem>
<TabItem value="CLI-users-new" label="OpenCLI">

Új felhasználó létrehozásához futtassa a következő parancsot:

```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_NAME>
```
Példa:
```bash
opencli user-add filip password1234 filip@openadmin.com default_plan_apache
```

:::tipp
Adja meg a "generate" jelszót egy erős véletlenszerű jelszó generálásához.
:::
</TabItem>
<TabItem value="API-users-new" label="API">

Új felhasználó létrehozásához használja a következő API-hívást:

```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer JWT_TOKEN_HERE" -d '{"email": "EMAIL_HERE", "username": "USERNAME_HERE", "password": "PASSWORD_HERE", "plan_name": "PLAN_NAME_HERE"}' http://PANEL:2087/api/users
```

Példa:
```bash
curl -X POST "http://PANEL:2087/api/users" -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGcBns" -H "Content-Type: application/json" -d '{"username":"stefan","password":"strongishpassword1234","email":"stefan@pejcic.rs","plan_name":"default_plan_nginx"}'
```

Példa válasz:

```json
{
  "response": {
    "message": "Successfully added user stefan password: strongishpassword1234"
  },
  "success": true
}
```
  
</TabItem>
</Tabs>

- Az OpenPanel felhasználónévnek 3-16 karakter hosszúnak kell lennie, és csak betűket és számokat tartalmazhat.
- Az OpenPanel jelszónak 6-30 karakter hosszúnak kell lennie, és bármilyen karaktert tartalmazhat, kivéve az idézőjeleket (`'`) és a dupla idézőjeleket (`"`).


## Egy felhasználó

Egy felhasználó részletes információinak megtekintéséhez és beállításainak szerkesztéséhez kattintson a felhasználónevére a felhasználók táblázatában.


### Statisztika

A Statisztika az alapértelmezett lap, az aktuális használati statisztikákat jeleníti meg:

- Használt tárolóhely
- Inodes használt
- CPU használat
- Memóriahasználat
- A futó konténerek száma
- Lemez I/O
- Hálózati I/O
- PID-k száma
- Az időstatisztika használatának legutóbbi frissítése
- Történelmi használat

A „Dokkerhasználati előzmények betöltése” lehetőségre kattintva megjelenik egy táblázat a felhasználó korábbi erőforrás-használatáról: Dátum, futó konténerek száma, CPU% és memória, Net I/O és Block I/O.

![felhasználói statisztikák](/img/admin/user_usage.png)


### Szolgáltatások

A Szolgáltatások lapon megjelenik az összes felhasználói szolgáltatás (docker konténer):

- A szolgáltatás neve
- Docker kép neve és címke
- Jelenlegi CPU használat
- Lefoglalt CPU a szolgáltatáshoz
- Aktuális memóriahasználat
- Lefoglalt memória a szolgáltatáshoz
- Jelenlegi állapot: Engedélyezett vagy Letiltva
- Terminálhivatkozás a docker exec parancsok futtatásához a szolgáltatásban.

![docker szolgáltatások](/img/admin/docker_services.png)

### Tárolás

A Tárolás lap a [docker system df](https://docs.docker.com/reference/cli/docker/system/df/) parancs adatait jeleníti meg.

- Kötetek
- Konténerek
- Képek

### Áttekintés

Az Áttekintő oldal részletes felhasználói információkat jelenít meg, és lehetővé teszi a rendszergazdának, hogy egyéni üzenetet állítson be kifejezetten ehhez a felhasználóhoz.

![felhasználói áttekintés](/img/admin/2025-06-09_08-34.png)

Megjelenített információ:

- Felhasználói azonosító
- E-mail cím
- IP-cím
- Földrajzi hely az IP-hez
- Szerver neve
- Docker kontextus
- 2FA állapot
- Beállítási idő
- Egyéni üzenet a felhasználó számára


### Tevékenység

Megjeleníti a [felhasználói tevékenységnaplót] (/docs/panel/account/account_activity/).

- Dátum
- Művelet végrehajtva
- IP-cím

![felhasználói tevékenység](/img/admin/login_log.png)

### Szerkesztés
A Szerkesztés lapon a rendszergazdák szerkeszthetik a felhasználói információkat:

- Felhasználónév
- E-mail cím
- Jelszó
- IP cím
- Tárhely csomag

![felhasználói szerkesztés](/img/admin/edit_user.png)

### Felfüggesztés

<Tabs>
<TabItem value="openadmin-user-suspend" label="OpenAdminnal" alapértelmezett>

A fiók felfüggesztése azonnal letiltja a felhasználó hozzáférését az OpenPanelhez. Ez a művelet magában foglalja a felhasználó Docker-tárolójának szüneteltetését, valamint az e-mailekhez, webhelyekhez és egyéb kapcsolódó szolgáltatásokhoz való hozzáférés visszavonását. Kérjük, vegye figyelembe az azonnali hatást, mielőtt továbblép.

Egy felhasználó felfüggesztéséhez kattintson a Felfüggesztés hivatkozásra az adott felhasználói oldalon, és a megerősítéshez írja be a felhasználónevet, majd kattintson a "Fiók felfüggesztése" gombra.

![felhasználó felfüggesztése](/img/admin/openadmin_suspend_user.gif)

</TabItem>
<TabItem value="CLI-user-suspend" label="OpenCLI-vel">

A felhasználó hozzáférésének felfüggesztéséhez (ideiglenes letiltásához) futtassa a következő parancsot:

```bash
opencli user-suspend <USERNAME>
```
Példa:

```bash
opencli user-suspend filip
```


</TabItem>
</Tabs>

### Felfüggesztés visszavonása

<Tabs>
<TabItem value="openadmin-user-unsuspend" label="OpenAdminnal" alapértelmezett>

Egy felhasználó felfüggesztéséhez kattintson a Felfüggesztés feloldása gombra az adott felhasználónál.

</TabItem>
<TabItem value="CLI-user-unsuspend" label="OpenCLI-vel">
    
A felfüggesztés feloldásához (a felhasználó hozzáférésének engedélyezéséhez) futtassa a következő parancsot:

```bash
opencli user-unsuspend <USERNAME>
```

Példa:
```bash
opencli user-unsuspend filip
```

</TabItem>
</Tabs>


### Jelszó visszaállítása

<Tabs>
<TabItem value="openadmin-users-reset" label="OpenAdmin" alapértelmezett>

Egy felhasználó jelszavának visszaállításához kattintson a "Szerkesztés" fülre, állítsa be az új jelszót a Jelszó mezőben, majd kattintson a Mentés gombra.

![új felhasználó hozzáadása openadmin](/img/admin/reset_password.png)


</TabItem>
<TabItem value="CLI-users-reset" label="OpenCLI">

Az OpenPanel-felhasználók jelszavának visszaállításához használja a "user-password" parancsot:

```bash
opencli user-password <USERNAME> <NEW_PASSWORD>
```

Használja a `--ssh' jelzőt a tárolóban lévő SSH-felhasználó jelszavának módosításához.

Példa:

```bash
opencli user-password filip Ty7_K8_M2 --ssh
```

</TabItem>
<TabItem value="API-users-reset" label="API">

Egy OpenPanel-felhasználó jelszavának visszaállításához használja a következő API-hívást:

```bash
curl -X PATCH http://PANEL:2087/api/users/USERNAME_HERE -H "Content-Type: application/json" -H "Authorization: Bearer JWT_TOKEN_HERE" -d '{"password": "NEW_PASSWORD_HERE"}'
```
</TabItem>
</Tabs>



### Átnevezés

<Tabs>
<TabItem value="openadmin-user-username" label="OpenAdminnal" alapértelmezett>

Felhasználó átnevezéséhez kattintson a felhasználó 'Információ szerkesztése' linkjére, majd módosítsa a címet a 'Felhasználónév' mezőben, majd kattintson a 'Változások mentése' gombra.


</TabItem>
<TabItem value="CLI-user-email" label="OpenCLI-vel">

Egy felhasználó felhasználónevének megváltoztatásához futtassa a következő parancsot:

```bash
opencli user-rename <old_username> <new_username>
```

Példa:

```bash
#opencli user-rename stefan pejcic
User 'stefan' successfully renamed to 'pejcic'.
```
</TabItem>
</Tabs>


### Csomag módosítása

<Tabs>
<TabItem value="openadmin-user-plan" label="OpenAdminnal" alapértelmezett>

Egy felhasználó csomagjának módosításához kattintson a felhasználó 'Szerkesztés' linkjére, majd válassza ki az új csomagot, és kattintson a 'Változások mentése' gombra.

</TabItem>
<TabItem value="CLI-user-plan" label="OpenCLI-vel">

Egy felhasználó csomagjának módosításához futtassa a következő parancsot:

```bash
opencli user-change_plan <USERNAME> '<NEW_PLAN_NAME>'
```
</TabItem>
</Tabs>


### E-mail módosítása

<Tabs>
<TabItem value="openadmin-user-email" label="OpenAdminnal" alapértelmezett>

Egy felhasználó e-mail címének megváltoztatásához kattintson a felhasználó "Információ szerkesztése" linkjére, majd módosítsa a címet az "E-mail cím" mezőben, és kattintson a "Változtatások mentése" gombra.

</TabItem>
<TabItem value="CLI-user-email" label="OpenCLI-vel">

A felhasználó e-mail címének megváltoztatásához futtassa a következő parancsot:

```bash
opencli user-email <USERNAME> <NEW_EMAIL>
```

Példa:

```bash
#opencli user-email stefan stefan@pejcic.rs
Email for user stefan updated to stefan@pejcic.rs.
```
</TabItem>
</Tabs>



### Jelentkezzen be az OpenPanelbe

Az OpenPanel fiókba való automatikus bejelentkezéshez kattintson az oldal jobb felső sarkában található **OpenPanel** gombra.
 

### Felhasználó törlése


<Tabs>
<TabItem value="openadmin-user-delete" label="OpenAdminnal" alapértelmezett>

Egy felhasználó törléséhez kattintson az adott felhasználó törlés gombjára, majd írja be a „törlés” szót a megerősítési módba, végül kattintson a „Megszakítás” gombra.


</TabItem>
<TabItem value="CLI-user-delete" label="OpenCLI-vel">
    
Egy felhasználó és minden adatának törléséhez futtassa a következő parancsot:

```bash
opencli user-delete <USERNAME>
```

adjon hozzá "-y" jelzőt a prompt letiltásához.

Példa:
```bash
opencli user-delete filip -y
```

</TabItem>
</Tabs>


:::veszély
Ez a művelet visszafordíthatatlan, és véglegesen törli az összes felhasználói adatot.
:::


---
sidebar_position: 3
---

lapok importálása a '@theme/Tabs'-ból;
TabItem importálása innen: '@theme/TabItem';


# Rendszergazdák

Az adminisztrációs panelnek három felhasználói szerepköre van:


| Szerep | Leírás |
| ------------------- | -------------------------------------------------------------------------- |
| **Super Admin** | Korlátlan jogosultságokkal rendelkezik, OpenPanel telepítéskor jött létre. |
| **Adminisztrátor** | Korlátozott jogosultságokkal rendelkezik, nem fér hozzá az összes OpenAdmin felhasználói felület oldalához, és nem szerkesztheti a SuperAdmin felhasználót. |
| **Viszonteladó** | Korlátozott jogosultságokkal rendelkezik. |


## Adminisztrátor felhasználók kezelése


<Tabs>
<TabItem value="openadmin-admin-users" label="OpenAdminnal" alapértelmezett>
  
Az OpenAdmin felülethez hozzáféréssel rendelkező adminisztratív felhasználókat a **Rendszergazdák** oldalon kezelheti.
  
Minden adminisztrátor felhasználó esetében megtekintheti és kezelheti a következő adatokat: Felhasználónév, Állapot Szerep, Utolsó bejelentkezési IP, Utolsó bejelentkezési idő, Műveletek.

</TabItem>

<TabItem value="CLI" label="OpenCLI-vel">

Az admin felhasználók listázásához használja a következő parancsot:

```bash
opencli admin list
```

</TabItem>
</Tabs>

## Állítsa vissza a rendszergazdai jelszót


<Tabs>
<TabItem value="openadmin-admin-reset" label="OpenAdminnal" alapértelmezett>

Az adminisztrátori jelszó visszaállításához kattintson az adott felhasználó Szerkesztés elemére a *Beállítások > Fiókok > Rendszergazdák* oldalon, majd állítsa be az új jelszót.

</TabItem>
<TabItem value="cli-reset" label="OpenCLI-vel">

Adminisztrátor jelszavának visszaállítása:

```bash
opencli admin password <username> <new_password>
```

Példa, jelszó visszaállítása a rendszergazdai felhasználóhoz:
```bash
opencli admin password admin Pyl7_L2M1
```

</TabItem>
</Tabs>


## Új adminisztrátor létrehozása

<Tabs>
<TabItem value="openadmin-admin-new" label="OpenAdminnal" alapértelmezett>

Új adminisztrátor létrehozásához kattintson az "Új létrehozása" gombra a *Beállítások > Fiókok > Rendszergazdák* oldalon, állítsa be a felhasználónevet és a jelszót, majd kattintson a *Mentés* gombra.


</TabItem>
<TabItem value="cli-new" label="OpenCLI-vel">

Új rendszergazdai fiókok létrehozása:

```bash
opencli admin new <username> <password>
```

Példa:
```bash
opencli admin new filip Pyl7_L2M1
```

</TabItem>
</Tabs>





## Admin felhasználó átnevezése

<Tabs>
<TabItem value="openadmin-admin-rename" label="OpenAdminnal" alapértelmezett>

Adminisztrátor átnevezéséhez válassza ki a **Beállítások > Fiókok > Rendszergazdák** oldalon, majd kattintson a Szerkesztés gombra, és állítsa be az új felhasználónevet.

![openadmin admin átnevezése](/img/admin/openadmin_admin_rename.png)


</TabItem>
<TabItem value="cli-rename" label="OpenCLI-vel">

Az adminisztrátori felhasználó átnevezése:

```bash
opencli admin rename <username> <new_username>
```

Példa:
```bash
opencli admin rename filip filip2
```
</TabItem>
</Tabs>


## Az Admin felhasználó felfüggesztése

<Tabs>
<TabItem value="openadmin-admin-suspend" label="OpenAdminnal" alapértelmezett>

Adminisztrátor felhasználó felfüggesztésének feloldásához válassza ki a felhasználót a **Beállítások > Fiókok > Rendszergazdák** oldalon, majd kattintson a Szerkesztés gombra, majd a **Felfüggesztés feloldása** elemre.

</TabItem>
<TabItem value="cli-suspend" label="OpenCLI-vel">

```bash
opencli admin suspend <username>
```

Példa:
```bash
opencli admin suspend filip
```
---

Az adminisztrátori felhasználó felfüggesztésének feloldása:
```bash
opencli admin unsuspend <username>
```

Példa:
```bash
opencli admin unsuspend filip
```

</TabItem>
</Tabs>


## Admin felhasználó törlése

<Tabs>
<TabItem value="openadmin-admin-delete" label="OpenAdminnal" alapértelmezett>

Válassza ki a felhasználót a **Beállítások > Fiókok > Rendszergazdák** oldalon, kattintson a törlés gombra, majd erősítse meg.

</TabItem>
<TabItem value="cli-delete" label="OpenCLI-vel">

A terminálról:

Adminisztrátor törlése:
```bash
opencli admin delete <username>
```

Példa:
```bash
opencli admin delete filip
```

</TabItem>
</Tabs>


:::info
A Super Admin felhasználó nem törölhető.
:::



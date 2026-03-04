---
sidebar_position: 2
---

# MySQL felhasználók

Ez a szakasz felsorolja az összes MySQL-felhasználót, és lehetőséget kínál a felhasználó jelszavának visszaállítására vagy a felhasználó törlésére.

A Felhasználók oldalon a következő lehetőségek állnak rendelkezésre:

- **Új felhasználó létrehozása**
- **Felhasználó hozzárendelése az adatbázishoz**
- **Felhasználó eltávolítása az adatbázisból**
- **A felhasználó jelszavának visszaállítása**
- **Felhasználó törlése**

## Hozzon létre egy új adatbázis-felhasználót

A MySQL-felhasználók nélkülözhetetlenek ahhoz, hogy ellenőrizzék, ki férhet hozzá az adatbázisokhoz, és ki léphet interakcióba azokkal, biztosítva az adatbiztonságot és szabályozott hozzáférést a webhely információihoz.

Új adatbázis-felhasználó létrehozásához kattintson az "Új felhasználó" gombra, és adja meg az új felhasználó nevét és jelszavát.

![databases_new_user.png](/img/panel/v2/databases_new_user.png)

## Felhasználó hozzárendelése egy adatbázishoz

Ahhoz, hogy egy MySQL felhasználó csatlakozhasson egy adatbázishoz, hozzá kell adni (hozzá kell rendelni) ahhoz az adatbázishoz. Egy felhasználó hozzárendelése minden jogosultságot biztosít az adatbázishoz. Ha egy felhasználót egy adott adatbázishoz szeretne rendelni, kattintson a "Hozzárendelés adatbázishoz" gombra, és válassza ki a felhasználónevet és az adatbázist.

![databases_assign.png](/img/panel/v2/databases_assign.png)

## Felhasználó eltávolítása az adatbázisból

Egy felhasználó adatbázisból való eltávolításához egyszerűen kattintson a "Felhasználó eltávolítása az adatbázisból" gombra, és az új módban válassza ki az adatbázisból eltávolítandó felhasználónevet.

![databases_remove_user.png](/img/panel/v2/databases_remove.png)

Egy felhasználó eltávolítása azonnal eltávolítja az adott felhasználónak az adatbázishoz való összes engedélyét, és akkor hasznos, ha ideiglenesen le szeretné tiltani a felhasználó hozzáférését az adatbázishoz anélkül, hogy ténylegesen törölné a felhasználót.

## Felhasználói jelszó módosítása

Ha módosítania kell egy felhasználó jelszavát, egyszerűen kattintson az adott felhasználó melletti "Jelszó módosítása" gombra. Megjelenik egy modál, és beírhatja az új jelszót a mezőbe, majd kattintson a „Jelszó módosítása” gombra az új jelszó mentéséhez.

![databases_reset_password.png](/img/panel/v2/databases_usrpw.png)

## Felhasználó törlése

MySQL felhasználó törléséhez kattintson a Felhasználók táblázatban a felhasználó melletti törlés gombra, majd kattintson a megerősítés gombra ugyanazon a gombra:

![databases_delete_user.png](/img/panel/v2/databases_delusr.png)

:::veszély
⚠️ Egy MySQL-felhasználó törlése azonnal eltávolítja a felhasználót, és visszavonja az adatbázisok összes jogosultságát.
:::

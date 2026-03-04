# Google PageSpeed ​​Insights API-kulcs

Az OpenPanel **Site Manager** automatikusan ellenőrzi a Google PageSpeed ​​rendszert az Ön webhelyein naponta. Ha azonban ezt az ellenőrzést egy több mint 100 webhelyet tartalmazó megosztott tárhelyszerveren futtatja, a Google letilthatja a hozzáférést az API-kérések korlátai miatt. Ha ez megtörténik, nem jelennek meg a PageSpeed ​​adatok, és ehhez hasonló hibaüzenet jelenhet meg a Site Manager hálózat lapján:

```
API Error: Quota exceeded for quota metric ‘Queries’ and limit ‘Queries per day’ of service ‘pagespeedonline.googleapis.com’ for consumer ‘project_number:583797351490’. Aborting due to API error.
```

Ennek megakadályozása érdekében az OpenPanel-felhasználók létrehozhatnak egy **ingyenes Google PageSpeed ​​Insights API-kulcsot**, és beállíthatják azt webhelyeikhez.

---

## Google PageSpeed ​​Insights API-kulcs generálása

1. Nyissa meg a [Google Cloud Console-t] (https://console.cloud.google.com/apis/api/pagespeedonline.googleapis.com/), és válassza ki a projektet.
2. Kattintson az **Engedélyezés** gombra.
3. Lépjen a **Hitelesítési adatok > Hitelesítési adatok létrehozása > API-kulcs** elemre.
4. Várja meg, amíg az *“API-kulcs létrehozása…”* folyamat befejeződik, majd másolja ki az új API-kulcsot.

---

## Az API-kulcs hozzáadása az OpenPanelhez

1. Nyissa meg az **OpenPanel > Site Manager > [Select a website]** menüpontot.
2. Kattintson a **"Kattintson a PageSpeed ​​Insights API kulcs hozzáadásához"** elemre a PageSpeed ​​adatok szakasz alatt.
![api kulcs hozzáadása](/img/panel/v2/add_api_key.png)
3. Illessze be a kimásolt API-kulcsot, és kattintson a **Mentés** gombra.
![api kulcs hozzáadása](/img/panel/v2/add_key.png)

> Mentés után az API-kulcs a fiókjában található összes webhelyhez használatos lesz.

---

## Globális API-kulcs hozzáadása az OpenAdminban

A rendszergazdák beállíthatnak egy globális API-kulcsot, amely minden olyan felhasználóra vonatkozik, aki nem állította be saját kulcsát.

1. Nyissa meg az **OpenAdmin > Beállítások > Egyéni kód** menüpontot.
2. Illessze be API-kulcsát a **PageSpeed ​​API-kulcs** mezőbe.
3. Kattintson a **Mentés** gombra.

![admin key](/img/panel/v2/admin_add_key.png)

---

A konfigurálást követően az API-kulcs biztosítja, hogy a PageSpeed-adatok helyesen legyenek lekérve az összes webhelyen anélkül, hogy a Google kvótakorlátját meghaladná.

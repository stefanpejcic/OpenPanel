# Domain az alapértelmezett oldalt mutatja

Amikor hozzáad egy domaint az **OpenPanelhez**, először a következő üzenet jelenik meg:

> **Kész, kész, internet 🎉**
> Ennek a domainnek jelenleg nincs webhelye. Kérjük, nézzen vissza később.

Ez az alapértelmezett oldal addig jelenik meg, amíg egy "index.php" vagy "index.html" fájl megjelenik a tartomány dokumentumgyökérében.

![Alapértelmezett oldal](https://i.postimg.cc/Zn8gbHm6/2025-08-13-12-20.png)

**Megoldás:**
Töltse fel webhelye fájljait (beleértve az "index.php" vagy az "index.html" fájlt is) a dokumentum gyökérkönyvtárába. A feltöltés után webhelye lecseréli az alapértelmezett oldalt.

---

## Gyorsítótárazási probléma

Ha a **Lakk gyorsítótárazás** engedélyezve van, és még azelőtt hozzáfér a tartományhoz, hogy annak tartalma lenne, akkor előfordulhat, hogy az alapértelmezett oldal gyorsítótárba kerül.
Ez azt jelenti, hogy még a webhely feltöltése után is megjelenhet a gyorsítótárazott alapértelmezett oldal.

**Javítás:**

* Ideiglenesen tiltsa le, majd engedélyezze újra a Varnish gyorsítótárat a tartományhoz, vagy
* Teljesen tiltsa le, majd engedélyezze újra a Varnish alkalmazást az összes gyorsítótárazott tartalom törléséhez.

**A Lakk letiltása egy adott domainnél:**

1. Lépjen az **OpenPanel > Gyorsítótár > Lakk** elemre.
2. Tiltsa le a Lakk funkciót az adott tartományban

![Domain Varnish-gyorsítótárának letiltása](https://i.postimg.cc/dwSGj2qk/2025-08-13-12-25.png)

**Az összes lakk-gyorsítótár ürítése:**
Egyszerűen kattintson a **Letiltás**, majd az **Engedélyezés** lehetőségre ugyanazon az oldalon.

---

## Az alapértelmezett oldal testreszabása

A rendszergazdák módosíthatják a tartalom nélküli domaineken megjelenő alapértelmezett oldalt.

Testreszabás:

1. Az **OpenAdminban** lépjen a **Domainek > Domainsablonok szerkesztése** lehetőségre.
2. Módosítsa az alapértelmezett oldal HTML-kódját

![Alapértelmezett oldal szerkesztése az OpenAdminban](https://i.postimg.cc/JRx0Qm3T/2025-08-13-12-28.png)


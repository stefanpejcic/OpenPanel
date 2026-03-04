# 504 Hibaelhárítási útmutató

Ez az útmutató ismerteti az **504 Gateway Timeout** hibák leggyakoribb okait és azok kijavítását.

---

## 1. Indítsa újra a PHP-FPM-et

Az 502 Bad Gateway hiba gyakori az Nginx használatakor. Ez azt jelenti, hogy a webszerver (Nginx) nem kapott időben választ a háttérrendszertől (például PHP-FPM, Node.js vagy Python).

**A PHP-FPM ellenőrzésének lépései:**

1. Ellenőrizze, hogy fut-e a PHP szolgáltatás.
Ha a Docker funkció engedélyezve van, lépjen a **Docker > Tárolók** elemre, és ellenőrizze a PHP-tároló állapotát.
Ha inaktív, indítsa el.
![képernyőkép](https://i.postimg.cc/wx8Dm4XP/image.png)
Ha aktív, állítsa le és indítsa újra. Ez leállítja az esetlegesen elakadt háttérfolyamatokat.

---

## 2. Várj

Ha nem Ön a webhely tulajdonosa, és nincs hozzáférése az OpenPanel-fiókhoz, csak annyit tehet, hogy értesíti a tulajdonost a problémáról, és megvárja, amíg a webhely újra válaszol.

---

## 3. Frissítse a WHMCS-t

504-es hiba néha előfordulhat a WHMCS-ben, ha nem frissítették rendszeresen. A megoldás a WHMCS frissítése a legújabb verzióra.
[Útmutató a WHMCS frissítéséhez](https://docs.whmcs.com/8-13/system/updates/updating-whmcs/)


Ezekkel a lépésekkel képes lesz azonosítani és kijavítani az **504 Gateway Timeout** hibák leggyakoribb okait.

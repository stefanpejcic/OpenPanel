# Setup FTP

[OpenPanel Enterprise](https://openpanel.com/enterprise/) supports FTP access. Once enabled, a shared FTP server becomes available for all users.

Follow these steps to enable FTP in OpenPanel:

---

## 1. Activate the Enterprise License

Ensure that you're running the **Enterprise Edition** of OpenPanel. FTP support is only available in this version.

[Upgrading to OpenPanel Enterprise and activating License](/docs/articles/license/upgrade_to_openpanel_enterprise_and-activate_license/)

---

## 2. Enable the FTP Module

OpenPanel uses a modular system where features can be individually enabled or disabled.

To enable the FTP module:

* Go to **OpenAdmin > Settings > Modules**
* Find the **FTP** module and click **Activate**

![FTP Module Activation](https://i.postimg.cc/zGBwhJVz/2025-07-09-11-34.png)

---

## 3. Start the FTP Service

Az FTP szolgáltatás a "vsftpd" szervert használja.

Az indításhoz:

* Lépjen az **OpenAdmin > Szolgáltatások > FTP-fiókok** oldalra.
* Kattintson az **FTP-kiszolgáló indítása** gombra

![Start FTP Server](https://i.postimg.cc/pdMTHCNw/2025-07-09-11-35.png)

---

## 4. Engedélyezze az FTP szolgáltatást a Tárhelycsomagban

Az FTP-hozzáférést kifejezetten engedélyezni kell a tárhelyterv szolgáltatáskészletében.

Ehhez tegye a következőket:

* Lépjen az **OpenAdmin > Tárhelycsomagok > Funkciókezelő** oldalra.
* Válassza ki a módosítani kívánt szolgáltatáskészletet
* Engedélyezze az **ftp** funkciót

![FTP funkció engedélyezése](https://i.postimg.cc/mrSbGGy9/2025-07-09-11-38.png)

---

## 5. Indítsa újra az OpenPanel-t

Az OpenPanel újraindítása biztosítja, hogy az újonnan engedélyezett modul és szolgáltatáskészletek azonnal érvénybe lépjenek minden felhasználó számára.

Újraindításhoz:

* Lépjen az **OpenAdmin > Szolgáltatások > Állapot** oldalra.
* Kattintson az **Újraindítás** lehetőségre az **OpenPanel** mellett

![Az OpenPanel újraindítása](https://i.postimg.cc/pd1PdJ3V/2025-07-09-11-40.png)

---

## Az FTP beállítása kész

Az FTP mostantól minden olyan tárhelycsomaghoz engedélyezett, amely tartalmazza az **ftp** funkciót. Az érintett felhasználók az **OpenPanel > Fájlok > FTP-fiókok** alatt érhetik el.

[![2025-07-09-11-42.png](https://i.postimg.cc/rsjn8QYQ/2025-07-09-11-42.png)](https://postimg.cc/y3JXjXKZ)

---

## Hibaelhárítás

Ha a felhasználó a következő hibát látja:

> `Az FTP szolgáltatás még nem indult el, kérjük, lépjen kapcsolatba a rendszergazdával az engedélyezéséhez.`

Ez azt jelenti, hogy az FTP szolgáltatás nem fut. A probléma megoldásához keresse fel újra a **3. lépést: Indítsa el az FTP szolgáltatást**.

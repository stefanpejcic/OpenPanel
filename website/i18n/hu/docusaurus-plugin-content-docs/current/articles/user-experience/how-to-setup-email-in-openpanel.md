# Setup Email

[OpenPanel Enterprise](https://openpanel.com/enterprise/) supports Emails. Once enabled, a shared Email server becomes available for all users.

Follow these steps to enable Emails in OpenPanel:

---

## 1. Activate the Enterprise License

Ensure that you're running the **Enterprise Edition** of OpenPanel. Email support is only available in this version.

[Upgrading to OpenPanel Enterprise and activating License](/docs/articles/license/upgrade_to_openpanel_enterprise_and-activate_license/)

---

## 2. Engedélyezze az E-mail modult

Az OpenPanel moduláris rendszert használ, ahol a funkciók egyenként engedélyezhetők vagy letilthatók.

Az E-mail modul engedélyezése:

* Lépjen az **OpenAdmin > Beállítások > Modulok** oldalra.
* Keresse meg az **E-mailek** modult, és kattintson az **Aktiválás** gombra.

![E-mail modul aktiválása](https://i.postimg.cc/9FvhmLSX/2025-07-09-17-12.png)

---

## 3. Telepítse az E-mail szolgáltatást

Az E-mail szolgáltatás a [`docker-mailserver`](https://docker-mailserver.github.io/docker-mailserver/latest/) veremét használja.

Az indításhoz:

* Lépjen az **OpenAdmin > E-mailek > E-mail fiókok** oldalra.
* Másolja ki az "opencli email-server install" oldalon megjelenő parancsot
* Illessze be a terminálba, és várja meg, amíg a folyamat befejeződik.

[![2025-07-09-17-14.png](https://i.postimg.cc/YCzMd1np/2025-07-09-17-14.png)](https://postimg.cc/G4tW2snN)

---

## 4. Engedélyezze az e-mailek funkciót a hosting tervekben

Az e-mailekhez való hozzáférést kifejezetten engedélyezni kell a tárhelyterv szolgáltatáskészletében.

Ehhez tegye a következőket:

* Lépjen az **OpenAdmin > Tárhelycsomagok > Funkciókezelő** oldalra.
* Válassza ki a módosítani kívánt szolgáltatáskészletet
* Engedélyezze az **e-mailek** funkciót

[![2025-07-09-17-17.png](https://i.postimg.cc/NjfzwmGG/2025-07-09-17-17.png)](https://postimg.cc/zV6jCLR4)

---

## 5. Indítsa újra az OpenPanel-t

Az OpenPanel újraindítása biztosítja, hogy az újonnan engedélyezett modul és szolgáltatáskészletek azonnal érvénybe lépjenek minden felhasználó számára.

Újraindításhoz:

* Lépjen az **OpenAdmin > Szolgáltatások > Állapot** oldalra.
* Kattintson az **Újraindítás** lehetőségre az **OpenPanel** mellett

![Az OpenPanel újraindítása](https://i.postimg.cc/pd1PdJ3V/2025-07-09-11-40.png)

---

## Kész

Az e-mail mostantól az **e-mailek** funkciót tartalmazó összes tárhelycsomaghoz engedélyezve van. Az érintett felhasználók elérhetik a webmailt, és az e-mail fiókok űrlapja az **OpenPanel > E-mailek** alatt érhető el.

[![2025-07-09-17-19.png](https://i.postimg.cc/44QPNySY/2025-07-09-17-19.png)](https://postimg.cc/144wvmPS)

---

## Hibaelhárítás

Ha a levelezőszerver nem indul el, próbálja meg manuálisan elindítani a terminálról:

```bash
cd /usr/local/mail/openmail && docker compose up mailserver
```

Tekintse át az indítási naplókat az esetleges hibákért, javítsa ki azokat, és amint a szolgáltatás megfelelően fut, állítsa le, és indítsa újra leválasztott módban a `-d` jelzővel.

A leggyakoribb okok, amelyek miatt a levelezőszerver nem indul el, a következők:

* **Egy másik szolgáltatás, amely már használja a 25-ös portot** – például az Exim.
**Megoldás:** Tiltsa le vagy állítsa le az ütköző szolgáltatást.

* **Tűzfal- vagy iptables-hiba**, amelyek blokkolják a docker indítását.
**Megoldás:** Indítsa újra a Docker szolgáltatást a hálózati szabályok visszaállításához.

* **Érvénytelen konfiguráció** a levelezőszerverhez.
**Megoldás:** Állítsa vissza az alapértelmezett levelezőszerver .env fájlt.

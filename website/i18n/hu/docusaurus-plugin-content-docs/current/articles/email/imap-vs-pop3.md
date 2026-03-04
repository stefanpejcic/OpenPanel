# IMAP vs POP3

📬 POP3 vs IMAP: E-mail hozzáférési protokollok

A **POP3** és **IMAP** két különálló protokoll a szerveren tárolt e-mailek elérésére.

* A **POP3 (Post Office Protocol v3)** bevezetése **[1996](https://www.ietf.org/rfc/rfc1939.txt)**. Korlátozott internetkapcsolattal és minimális szervertárhellyel rendelkező környezetekhez tervezték. Letölti az e-maileket egy helyi eszközre, majd általában eltávolítja őket a szerverről – így ideális régebbi telefonos kapcsolatokhoz.

* Az **IMAP (Internet Message Access Protocol)** később, **[2003-ban] (https://datatracker.ietf.org/doc/html/rfc3501)** jelent meg, és az állandó szélessávú kapcsolatok (például kábel vagy DSL) mellett fejlődött. Az IMAP a szerveren tárolja az e-maileket, és minden eszközön szinkronizálja állapotukat (olvasott, olvasatlan, megválaszolt, címkézett stb.). Ez ideálissá teszi modern használatra telefonokon, táblagépeken és asztali számítógépeken.

---

## Főbb különbségek

| Funkció | **IMAP** | **POP3** |
| ---------------------- | -------------------------------- | ----------------------------------------------- |
| **Mail Storage** | A **szerveren** marad | Letöltve a **helyi eszközre** |
| **Szinkronizálás** | Igen – minden eszközön | Nem – elérést követően eltávolítva a szerverről* |
| **Elküldött levél** | **szerveren** tárolva | **eszközön** tárolva |
| **Törölt levél** | A kukába megy (ki kell üríteni) | Csak az eszközről eltávolítva (nincs hatással a szerverre) |
| **Katasztrófa-helyreállítás** | Igen – a szerver biztonsági másolataival | Nem – csak helyben tárolva |
| **Offline hozzáférés** | Nem – internet szükséges | Igen – letöltés után |

> * A POP3 **beállítható** úgy, hogy üzeneteket hagyjon a szerveren, de ez gyakran duplikált e-mailekhez vezet a különböző eszközökön, és nem ajánlott. A megfelelő szinkronizáláshoz használja az IMAP protokollt.

---

## Melyik a megfelelő számomra?

**Válassza az IMAP-ot** – ez a modern szabvány, és a legtöbb felhasználó számára a legjobb megoldás. Lehetővé teszi a teljes szinkronizálást minden eszközön, és biztosítja, hogy e-mailjei biztonságban és elérhetők maradjanak a szerveren.

**Csak akkor válassza a POP3-at**, ha:

* **korlátozott internet-hozzáférése van**
* **el kell** hozzáférned e-mailjéhez **offline**
* A szerver tárolása erősen korlátozott

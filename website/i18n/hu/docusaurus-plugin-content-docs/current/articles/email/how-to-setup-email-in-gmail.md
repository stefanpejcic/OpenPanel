# Hogyan állítsunk be e-mailt a Gmailben

Ez az útmutató elmagyarázza, hogyan adhatja hozzá **OpenPanel által létrehozott e-mail fiókját** a Gmailhez webböngészővel.

> **Megjegyzés:** Ha egy másik eszközön vagy szolgáltatáson állít be e-mailt, tekintse meg a fő útmutatót:
> [Az e-mail kliens beállítása](/docs/articles/email/how-to-setup-your-email-client)

---

## Új fiók hozzáadása a Gmailben

1. Nyissa meg a **Gmailt** a böngészőjében

2. Kattintson a **⚙️ Beállítások ikonra → Az összes beállítás megtekintése → Fiókok és importálás fülre**

3. A **„E-mailek ellenőrzése más fiókokból”** részben kattintson a **Levelezési fiók hozzáadása** lehetőségre.

4. Adja meg **teljes e-mail címét**, majd kattintson a **Tovább** gombra.

5. Válassza az **E-mailek importálása másik fiókomból (POP3)** lehetőséget, majd kattintson a **Tovább** gombra.

---

## Bejövő levelek (POP3) beállításai

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Bejövő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a bejövő levelekhez | 995 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Jelszó |
| További | Opcionálisan: "Hagyjon másolatot a szerveren" biztonsági mentéshez | Bejelölve / Nem ellenőrizve |

A folytatáshoz kattintson a **Fiók hozzáadása** lehetőségre.

---

## Kimenő levelek (SMTP) beállításai

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Kimenő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a kimenő levelekhez | 465 / 587 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A levélküldéshez használt hitelesítési módszer | Jelszó |

Kattintson a **Fiók hozzáadása** lehetőségre a beállítás befejezéséhez.

---

OpenPanel e-mailje mostantól integrálva van a Gmaillel. Gmail-fiókjából közvetlenül **küldhet és fogadhat e-maileket**.

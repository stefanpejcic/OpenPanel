# Hogyan állítsunk be e-mailt a Thunderbird-en

Ez az útmutató elmagyarázza, hogyan állíthatja be az **OpenPanel által létrehozott e-mail fiókot** a Thunderbird levelezőalkalmazásban.

> **Megjegyzés:** Ha egy másik eszközön vagy szolgáltatáson állít be e-mailt, tekintse meg a fő útmutatót:
> [Az e-mail kliens beállítása](/docs/articles/email/how-to-setup-your-email-client)

---

## Új fiók hozzáadása

1. Nyissa meg a **Thunderbird → Fájl → Új → Meglévő levelezési fiók** lehetőséget.

2. Adja meg a következő adatokat:
- **Az Ön neve**: Az Ön megjelenített neve
- **E-mail cím**: Az Ön teljes e-mail címe
- **Jelszó**: Az e-mail fiók jelszava
   
3. Kattintson a **Tovább** gombra.

> A Thunderbird megpróbálja automatikusan észlelni a beállításait. Ha hiba történik, hagyja figyelmen kívül, és adja meg manuálisan az adatokat.

---

## Válassza ki a Fiók típusát

A Thunderbird engedélyezi az **IMAP** vagy **POP** használatát.
- Ha nem biztos benne, válassza az **IMAP** lehetőséget (ajánlott).
- További információ: [📬 IMAP vs POP3](/docs/articles/email/imap-vs-pop3/)

---

## Bejövő levelek beállításai

### IMAP (ajánlott)

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Bejövő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a bejövő levelekhez | 993 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Normál jelszó |

### POP

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Bejövő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a bejövő levelekhez | 995 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Normál jelszó |

A bejövő levelek beállításainak megadása után kattintson a **Kész** gombra.

---

## Kimenő levelek beállításai (SMTP)

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Kimenő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a kimenő levelekhez | 465 / 587 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Normál jelszó |

---

E-mail fiókja készen áll a Thunderbirdben való használatra.

# Az e-mail beállítása Androidon (Gmail alkalmazás)

Ez az útmutató elmagyarázza, hogyan állíthatja be az **OpenPanel által létrehozott e-mail fiókot** az Android Gmail alkalmazásában.

> **Megjegyzés:** Ha egy másik eszközön vagy szolgáltatáson állít be e-mailt, tekintse meg a fő útmutatót:
> [Az e-mail kliens beállítása](/docs/articles/email/how-to-setup-your-email-client)

---

## Új fiók hozzáadása

1. Nyissa meg a **Gmail alkalmazást**
Ugrás ide:
**Menü → Beállítások → Fiók hozzáadása → Személyes (IMAP/POP) → Következő**

2. Adja meg **e-mail címét** → **Következő**

3. Válassza az **IMAP** vagy **POP** lehetőséget
- Ha nem biztos benne, válassza az **IMAP** lehetőséget (ajánlott).
- További információ: [📬 POP3 vs IMAP: E-mail hozzáférési protokollok](/docs/articles/email/imap-vs-pop3/)

4. Adja meg **e-mail jelszavát** → **Következő**

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

---

### POP

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Bejövő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a bejövő levelekhez | 993 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Normál jelszó |

Miután megadta a bejövő levelek beállításait, érintse meg a **Tovább** gombot.

---

## Kimenő levelek beállításai (SMTP)

Győződjön meg arról, hogy a **„Bejelentkezés szükséges”** engedélyezve van.

| Beállítás | Leírás | Példa |
|----------------|---------------------------------------------------------------------------------|
| Felhasználónév | Az Ön teljes e-mail címe | user@domain.tld |
| Jelszó | E-mail fiókod jelszava | ******** |
| Szerver | Kimenő levelek szerver címe | mail.domain.tld |
| Kikötő | Portszám a kimenő levelekhez | 465 / 587 |
| Biztonság típusa | Titkosítási módszer a biztonságos kapcsolathoz | SSL/TLS |
| Hitelesítés | A bejelentkezéshez használt hitelesítési módszer | Jelszó |

A kimenő levelek beállításainak megadása után érintse meg a **Tovább** gombot.

---

OpenPanel e-mail fiókja be van állítva, és készen áll a használatra a Gmail alkalmazásban.

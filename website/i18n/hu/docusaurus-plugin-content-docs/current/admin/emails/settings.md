---
sidebar_position: 3
---

# E-mail beállítások

Az E-mail beállítások szakasz lehetővé teszi a MailServer verem különféle paramétereinek konfigurálását a hatékony és biztonságos e-mail-kezelés érdekében.

:::info
Az e-mailek csak az [OpenPanel Enterprise Edition](/enterprise) felületen érhetők el
:::


## MailServer állapota

A Mail Server szolgáltatás állapota az oldal tetején jelenik meg, ahol a rendszergazdák szükség szerint elindíthatják, leállíthatják vagy újraindíthatják a szolgáltatást.

## Fiókok

Megjeleníti a szerveren jelenleg aktív e-mail fiókok teljes számát. Ez magában foglalja az összes fiókot a rendszeren konfigurált összes tartományban.


## Webmail

- Állapot - megjeleníti a webmail szolgáltatás aktuális állapotát
- Aktuális szoftver - megjeleníti az aktuális kiválasztott klienst
- Webmail kliens kiválasztása - Válassza ki azt a webmail klienst, amellyel a felhasználók kapcsolatba lépnek. A szolgáltatás újraindul az esetleges módosítások végrehajtásához.
- Webmail tartomány beállítása - A webmail szolgáltatáshoz használandó tartomány beállítása. A webmail elérhető lesz ezen a tartományon, és a /webmail minden felhasználói domainben átirányít erre a tartományra.


## Szolgáltatások engedélyezése

A rendszergazdák különféle szolgáltatásokat állíthatnak be és konfigurálhatnak igényeik szerint.

Szolgáltatások konfigurálása a MailServer veremhez:

| Szolgáltatás | Leírás |
|------------------------------------------------------------ ------------------------------------------------------------|
| **Amavis** | Amavis tartalomszűrő (a ClamAV és a SpamAssassin számára használatos).                      |
| **DNS-blokkoló listák** | Engedélyezi a DNS-blokkoló listákat a Postscreenben.                                       |
| **Rspamd** | Rspamd engedélyezése vagy letiltása.                                                    |
| **SpamAssassin** | Elemzi a bejövő leveleket, és spam-pontszámot rendel hozzá.                            |
| **MTA-STS** | Lehetővé teszi az MTA-STS támogatást a kimenő levelekhez.                                  |
| **OpenDKIM szolgáltatás** | Engedélyezi az OpenDKIM szolgáltatást e-mail-aláíráshoz.                             |
| **OpenDMARC szolgáltatás** | Engedélyezi az OpenDMARC szolgáltatást az e-mail tartományalapú üzenetek hitelesítéséhez. |
| **POP3** | Engedélyezi a POP3 szolgáltatást az e-mailek lekéréséhez.                               |
| **IMAP** | Engedélyezi az IMAP szolgáltatást az e-mailek lekéréséhez.                               |
| **ClamAV** | Engedélyezi a ClamAV víruskereső szolgáltatást.                                       |
| **fail2ban** | Lehetővé teszi a fail2ban szolgáltatás számára az IP-címek kitiltását gyanús tevékenység alapján.       |
| **Csak SMTP** | Ha engedélyezve van, csak a Postfix szolgáltatás indul el, és a felhasználók nem fogadhatnak bejövő e-maileket. |
| **Feladó átírási séma** | Engedélyezi a küldő átírási sémát, amely az e-mailek továbbításához szükséges (magyarázat: [postsrsd](https://github.com/roehling/postsrsd/blob/main/README.rst). |


A szolgáltatás módosításai megszakítják a jelenlegi e-mail forgalmat, és újraindítják a levelezőszervert.




## Relay Hosts

A **Relay Hosts** funkció lehetővé teszi egy SMTP továbbítási szolgáltatás (más néven továbbítási gazdagép vagy smarthost) konfigurálását a kimenő e-mailek továbbítására (továbbítására) harmadik felek nevében. Ez a szolgáltatás nem kezeli a levelezési tartományokat, de segít az e-mailek külső SMTP-kiszolgálón keresztül történő irányításában.

Ez a funkció olyan szervezetek számára hasznos, amelyeknek a kimenő e-mail forgalmukat megbízható harmadik féltől származó szolgáltatáson vagy SMTP-kiszolgálón kell átirányítaniuk a jobb kézbesítés és biztonság érdekében.

A következő paraméterek használhatók a közvetítő gazdagép beállításainak konfigurálásához:

- **DEFAULT_RELAY_HOST**
Alapértelmezett továbbítási gazdagép a kimenő e-mailekhez. Ennek meg kell egyeznie a **RELAY_HOST**-val.
- Példa: `mail.example.com`

- **RELAY_HOST**
Az SMTP közvetítő gazdagép, amelyen keresztül minden kimenő e-mail át lesz irányítva.
- Példa: `mail.example.com`

- **RELAY_PORT**
Az SMTP közvetítő gazdagéphez való csatlakozáshoz használt port.
- Példa: "25".

- **RELAY_USER (opcionális)**
A közvetítő gazdagéppel történő hitelesítéshez szükséges felhasználónév. Ha ez be van állítva, biztonságos kapcsolatokra lesz szükség a kimenő levélforgalomhoz.
- Példa: `relay_user`

- **RELAY_PASSWORD**
A közvetítő gazdagépével történő hitelesítéshez használt jelszó, a **RELAY_USER** mellett használatos.
- Példa: `relay_password`

Ha a **RELAY_USER** és a **RELAY_PASSWORD** is be van állítva, minden kimenő levélforgalomhoz biztonságos kapcsolatra lesz szükség, és a hitelesítő adatok megadása kötelező.

A konfigurálás után kattintson a **Relay mentése** gombra a beállítások alkalmazásához, és megkezdje a kimenő e-mailek átirányítását a megadott továbbítási gazdagépen.




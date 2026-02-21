---
sidebar_position: 2
---

# Folyamatkezelő

A **Process Manager** felület lehetővé teszi a szerverén futó összes folyamat megfigyelését. Közvetlenül az OpenPanel felületéről kereshet, tekinthet meg részletes parancsinformációkat, és leállíthatja (leállíthatja) az egyes folyamatokat.

A folyamatok **CPU-használat szerint vannak rendezve**, ami megkönnyíti az erőforrás-igényes feladatok azonosítását és végrehajtását.

## Főbb jellemzők

- 🔍 **Keresés:** Gyorsan szűrheti a folyamatokat tárolónév, PID vagy parancs alapján.
- 🧾 **Részletes információ:** A kulcsfontosságú folyamatok metaadatainak megtekintése, például:
- Konténer neve
- UID / PID / PPID
- CPU %
- Kezdés időpontja (STIME)
- TTY (terminál)
- Teljes végrehajtási idő
- Teljes parancs (bővíthető)
- 🛑 **Folyamatok leállítása:** Minden nem kritikus folyamat kényszerített leállítása.

:::veszély
⚠️ **Figyelem:** Az olyan alapszolgáltatások leállítása, mint a "MySQL", "PHP-FPM" vagy "Nginx/Apache", a webhelyek offline állapotba kerülését eredményezi. Csak olyan folyamatokat szakítson meg, amelyekben biztos vagy.
:::

---

## Hogyan kell használni

1. **Lépjen a** `Speciális > Folyamatkezelő` pontra az OpenPanel oldalsávon.
2. A **keresőmezővel** kereshet meg egy adott folyamatot PID, parancs vagy tárolónév alapján.
3. Kattintson a **"Teljes parancs megtekintése"** lehetőségre a régóta futó parancssorok kibontásához.
4. Kattintson a **Kill** gombra a folyamat azonnali leállításához.

## Interfész részletei

A táblázat minden sora a következőket tartalmazza:

| oszlop | Leírás |
|--------|--------------|
| **Konténer** | A tároló/szolgáltatás, ahol a folyamat fut |
| **UID** | A folyamat tulajdonosának felhasználói azonosítója |
| **PID** | Egyedi folyamatazonosító |
| **PPID** | Szülő folyamatazonosító |
| **CPU %** | CPU-használat százalékos |
| **STIME** | Folyamat kezdési időpontja |
| **TTY** | Kapcsolódó terminál ("?": leválasztott/háttér) |
| **IDŐ** | Teljes CPU-idő elhasznált |
| **CMD** | A végrehajtott parancs |
| **Akció** | Gomb a folyamat leállításához |

---

## Kill Process Behavior

Ha a **Kill** gombra kattint, a következő történik:

1. Megjelenik egy értesítés: _„PID megszűnése: xxxx...”_
2. A rendszer egy `POST' kérést küld a háttérrendszernek a `PID-vel' a befejezéshez.
3. Siker vagy kudarc esetén utólagos pohárköszöntőt kap az eredménnyel.

> Ez a felület csak szűrt felhasználói szintű folyamatokat jelenít meg. A belső vagy rendszerszintű karbantartási parancsok (például `/etc/entrypoint.sh` vagy `ps -eo`) automatikusan kizárásra kerülnek.

---

## Tippek

- Gyanús vagy magas CPU-használati folyamatok esetén ellenőrizze a **teljes parancsot**, mielőtt intézkedik.
- A **Folyamatok frissítése** gombbal bármikor újratöltheti a folyamatlistát.
- Egy szülőfolyamat (PPID) leállítása leállíthatja annak utódfolyamatait is.

---

Van még kérdése? Forduljon a kiszolgáló rendszergazdájához, vagy tekintse meg a rendszernaplókat, hogy mélyebb betekintést kapjon az ismétlődő folyamatokba.

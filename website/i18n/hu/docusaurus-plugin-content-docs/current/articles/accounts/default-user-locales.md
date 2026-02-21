# Alapértelmezett nyelv

Az **OpenPanel** teljes mértékben honosításra kész, de jelenleg csak angolul ("en") lesz az alapértelmezett nyelv.
További nyelveket a rendszergazda telepíthet az ***OpenAdmin → Beállítások → Helyek*** menüpontból, és ezek azonnal elérhetővé válnak a felhasználói felületen.

Ha az Ön által preferált nyelv még nem érhető el, fordítással segíthet: 👉 [OpenPanel Translations on GitHub](https://github.com/stefanpejcic/openPanel-translations)

---

## Az alapértelmezett terület beállítása

Az OpenPanelben **öt** különböző forrás határozza meg, hogy melyik területi beállítás kerül alkalmazásra a felhasználóra.
Ellenőrzésük a következő prioritási sorrendben történik:

| Prioritás | Forrás | Példa elérési út / adatok |
| -------- | ------------------------------ | ------------------------------------------------------ |
| 1️⃣ | Munkamenet (`session['locale']`) | "fr" |
| 2️⃣ | Felhasználó-specifikus fájl | `/home/<kontextus>/locale` |
| 3️⃣ | Böngésző legjobb találata | `Accept-Language: es-ES,es;q=0,9` |
| 4️⃣ | Alapértelmezett területi beállításfájl | `/etc/openpanel/openpanel/default_locale` → pl. "de" |
| 5️⃣ | Hardcoded végső tartalék | "en" |

---

### 1. Session Locale

Amikor egy felhasználó meglátogatja az OpenPanel-t, a rendszer először ellenőrzi a szekcióban a tárolt területi beállításokat.
A munkamenet területi beállítása a felhasználó bejelentkezésekor jön létre.
Például, ha a felhasználó a „de” (német) lehetőséget választja a bejelentkezési oldalon, akkor ez a területi beállítás az aktuális munkamenethez lesz beállítva.

> **Megjegyzés:** Ez a beállítás felülír minden más területi forrást.

---

### 2. Felhasználó-specifikus fájl

A bejelentkezés után a felhasználók megváltoztathatják preferált nyelvüket – ha a „locale” modul és funkció engedélyezve van számukra.
Lépjen a következőhöz: ***OpenPanel → Fiók → Nyelv módosítása*** ([Dokumentumok hivatkozás](/docs/panel/account/language/)) – ez a nézet felsorolja az adminisztrátor által telepített összes nyelvet.
A felhasználó által választott területi beállítás felülírja a böngésző beállításait és a rendszer alapértelmezett beállításait.

A felhasználó preferenciái egy felhasználónkénti fájlban tárolódnak:

```
/home/<context>/locale
```

---

### 3. A böngésző legjobb találata

Ha a felhasználó még nem állított be nyelvi beállítást, az OpenPanel ellenőrzi a böngésző [`Accept-Language` fejlécét] (https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept-Language).
Ha az előnyben részesített nyelv telepítve van a kiszolgálón, akkor a rendszer ideiglenesen alkalmazza az adott munkamenethez.

> **MEGJEGYZÉS**: Ez csak az **aktuális munkamenetet** érinti, és nem marad felhasználói preferenciaként.
> Csak az adott munkamenetre vonatkozóan írja felül a rendszergazda alapértelmezett területi beállítását.

Példa fejléc:

```
Accept-Language: es-ES,es;q=0.9
```

---

### 4. Default Locale

The Administrator can set a global default locale by creating a configuration file at:

```
/etc/openpanel/openpanel/default_locale
```

For example, to set Spanish (`es`) as the default:

```bash
echo 'es' > /etc/openpanel/openpanel/default_locale
```

---

### 5. Fallback Locale

If no other locale setting is found, OpenPanel defaults to English (`en`), which is included by default.

---

## Checking Which Locale Is Active

To verify which locale is being used for a user, you can temporarily enable developer mode and check the logs.

1. Enable `dev_mode`:

   ```bash
   opencli config update dev_mode on
   ```

2. Repeat the user action in the browser.

3. Tail the logs:

   ```bash
   docker logs -f openpanel
   ```

Look for log lines similar to:

```
APP - Using locale..
```

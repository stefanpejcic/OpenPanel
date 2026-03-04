# Billentyűparancsok

Az OpenAdmin felhasználói felületén billentyűparancsokkal lehet navigálni.

## Elérhető billentyűparancsok

Az összes elérhető billentyűparancs megtekintéséhez: `Ctrl` + `K`

| Parancsikon | Akció | URL elérési útja |
|------------------------------------------------------------------------------------------------|
| `Ctrl` + `Escape` | Kijelentkezés | "/kijelentkezés" |
| "Ctrl" + "Shift" + "H" | Irányítópult | "/műszerfal" |
| "Ctrl" + "Shift" + "N" | Értesítések | `/view_notifications` |
| "Ctrl" + "Shift" + "U" | Felhasználók | "/felhasználók" |
| "Ctrl" + "Shift" + "R" | Viszonteladók | "/viszonteladók" |
| "Ctrl" + "Shift" + "A" | Adminisztrátorok | "/adminisztrátorok" |
| "Ctrl" + "Shift" + "P" | Tervek | "/tervek" |
| "Ctrl" + "Shift" + "F" | Funkciókezelő | "/jellemzők" |
| "Ctrl" + "Shift" + "D" | Domainek | "/domains" |
| "Ctrl" + "Shift" + "E" | E-mail fiókok | "/e-mailek/fiókok" |
| "Ctrl" + "Shift" + "S" | Szolgáltatások állapota | "/szolgáltatások" |
| "Ctrl" + "Shift" + "C" | ConfigServer tűzfal (CSF) | "/biztonság/tűzfal" |
| "Ctrl" + "Shift" + "W" | CorazaWAF beállítások | "/biztonság/waf" |
| "Ctrl" + "Shift" + "L" | Naplók megtekintése | "/szolgáltatások/naplók" |
| "Ctrl" + "Shift" + "B" | Alapszintű hitelesítés | `/security/basic_auth` |
| `Ctrl` + `Shift` + `1` (csak számbillentyűzet) | Licenc | "/licenc" |

> ℹ️ Használja a „Cmd” parancsot macOS rendszeren a „Ctrl” helyett.

---

## Parancsikonok szerkesztése

Az OpenPanel 1.7.41 verziójából a rendszergazdák szerkeszthetik és testreszabhatják a parancsikonokat. A parancsikonok szerkesztéséhez szerkessze a következő fájlt: `/etc/openpanel/openadmin/config/shortcuts.json`.

Alapértelmezett tartalom:

```json
{
  "ctrl+shift+h": "/dashboard",
  "ctrl+shift+n": "/notifications",
  "ctrl+shift+u": "/users",
  "ctrl+shift+r": "/resellers",
  "ctrl+shift+a": "/administrators",
  "ctrl+shift+p": "/plans",
  "ctrl+shift+f": "/features",
  "ctrl+shift+d": "/domains",
  "ctrl+shift+e": "/emails/accounts",
  "ctrl+shift+s": "/services",
  "ctrl+shift+c": "/security/firewall",
  "ctrl+shift+w": "/security/waf",
  "ctrl+shift+l": "/services/logs",
  "ctrl+shift+b": "/security/basic_auth",
  "ctrl+escape": "/logout"
}
```

A fájl szerkesztése után indítsa újra az OpenAdmin programot: `service admin restart`.

---

## Parancsikonok letiltása

A parancsikonok teljes letiltásához törölje a következő fájlt: `/etc/openpanel/openadmin/config/shortcuts.json`

---
sidebar_position: 6
---

# Webszerver váltása

A **Docker > Webszerver váltása** oldal lehetővé teszi az aktuális webszerver közötti váltást a rendelkezésre álló lehetőségek között: **Nginx**, **OpenResty** és **Apache**.

Az aktuálisan aktív webszerver az oldal jobb felső sarkában jelenik meg.

## Követelmények

A funkció eléréséhez:

- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.

## Használat

A webszerver váltása előtt győződjön meg a következőkről:

- **Az összes meglévő domaint el kell távolítani.**
- A jelenlegi webszerver-tárolót **le kell állítani**, mielőtt elindítaná az újat.

> ⚠️ Ha már beállított tartományokat, **készítsen biztonsági másolatot az összes konfigurációról**, távolítsa el az összes tartományt egyenként, majd folytassa a webszerver váltásával.
> Az állásidő elkerülése érdekében a legjobb, ha ezt a módosítást **a domain hozzáadása előtt** hajtja végre.

### A váltás lépései

1. Az OpenPanel menüben lépjen a **Docker > Switch Web Server** elemre.
2. A legördülő menüből válassza ki a használni kívánt új webszervert.
3. Kattintson a **Switch** gombra a folyamat elindításához.

Megerősítés után:

- A meglévő webszerver-tároló leáll, és az adatai törlődnek.
- Elindul az új webszerver.
- Ezután újra felveheti domainjeit az új webszerver-konfiguráció alá.

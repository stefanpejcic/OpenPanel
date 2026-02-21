---
sidebar_position: 5
---

# Képcímke módosítása

A **Docker > Képcímke módosítása** oldal lehetővé teszi a szolgáltatások által használt Docker-képcímke (verzió) frissítését.

## Követelmények

A funkció eléréséhez:

- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.

## Használat

A Docker-kép címkéjének módosítása:

1. Először ellenőrizze a képhez elérhető címkéket a [Docker Hubon] (https://hub.docker.com/).
2. Az OpenPanel menüben válassza a **Docker > Képcímke módosítása** lehetőséget.
3. A **Szolgáltatás kiválasztása** alatt válassza ki azt a szolgáltatást, amelynek képcímkéjét módosítani szeretné.
4. Írja be az **New Image Tag** értéket a beviteli mezőbe.
5. Kattintson a **Címke módosítása** gombra a módosítás alkalmazásához.

Megerősítés után:

- A szolgáltatás automatikusan leáll.
- Az új képcímke le lesz húzva.
- A szolgáltatás a frissített képpel újraindul.

> ⚠️ Győződjön meg arról, hogy az új címke létezik, és kompatibilis a jelenlegi konfigurációjával, hogy elkerülje a szolgáltatás fennakadását.

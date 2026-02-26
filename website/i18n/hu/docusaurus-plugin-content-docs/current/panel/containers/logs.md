---
sidebar_position: 3
---

# Naplók

A **Docker > Naplók** oldalon közvetlenül az OpenPanel felületéről tekintheti meg a tárolónaplókat (`docker logs').

## Követelmények

A funkció eléréséhez:

- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.

## Hozzáférés a naplókhoz

1. Az OpenPanel menüben lépjen a **Docker > Naplók** elemre.
2. Kattintson a **Select Container** elemre az összes elérhető szolgáltatás listájának megjelenítéséhez.
3. Válassza ki azt a szolgáltatást, amelynek naplóit meg szeretné tekinteni.
4. Az alábbiakban megjelenik a kiválasztott tároló naplókimenete.

Opcionálisan módosíthatja a megjelenített naplósorok számát a naplók panel jobb felső sarkában található legördülő menü segítségével.

> 💡 A naplók lekérése a „docker logs” használatával történik, és a tároló stdout és stderr adatfolyamainak valós idejű kimenetét mutatják.

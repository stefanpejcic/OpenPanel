---
sidebar_position: 2
---

# Terminál

A **Terminal** oldal webalapú terminált (`docker exec`) biztosít a futó konténerekkel való közvetlen interakcióhoz az OpenPanel felületen keresztül.

## Követelmények

A terminál eléréséhez:

- A **Docker** modult a rendszergazdának engedélyeznie kell **szerveren**.
- Fiókjában engedélyezni kell a **Docker** funkciót.

## Hozzáférés a terminálhoz

1. Az OpenPanel menüben lépjen a **Docker > Terminal** elemre.
2. Kattintson a **Szolgáltatás kiválasztása** elemre a jelenleg futó szolgáltatások listájának megjelenítéséhez.
3. Kattintson az elérni kívánt szolgáltatásra.
4. Megnyílik a terminálablak, amely lehetővé teszi a tárolón belüli parancsok futtatását.

A terminál jobb felső sarkában található választó segítségével válthat a shell típusa között az "sh" és a "bash" között.

---

> 💡 Ez a funkció a 'docker exec' parancsot használja a motorháztető alatt, így valós időben közvetlen hozzáférést biztosít a konténer shell környezetéhez.

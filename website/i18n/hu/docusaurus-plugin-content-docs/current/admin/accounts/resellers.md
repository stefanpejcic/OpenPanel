---
sidebar_position: 2
---

# Viszonteladók

A **Viszonteladók** funkció lehetővé teszi a rendszergazdák számára, hogy viszonteladói felhasználókat hozzanak létre és kezeljenek az OpenPanelen belül. A viszonteladók alrendszergazdaként működhetnek, és a root rendszergazda által meghatározott korlátokon belül kezelhetik saját felhasználói fiókjukat.

Ez a funkció azon tárhelyszolgáltatók számára hasznos, akik az irányítást külső viszonteladókra kívánják átruházni, miközben továbbra is fenntartják az általános ellenőrzést és az elszigeteltséget.

---

## Viszonteladói kezelőfelület

A felület megjeleníti a meglévő viszonteladói felhasználók táblázatát a következő oszlopokkal:

- **Felhasználónév**
A viszonteladó felhasználó egyedi azonosítója.

- **Állapot**
Azt jelzi, hogy a viszonteladói fiók aktív vagy felfüggesztve.

- **Utolsó bejelentkezési IP**
Az az IP-cím, amelyről a viszonteladó utoljára hozzáfért a panelhez.

- **Utolsó bejelentkezési idő**
A viszonteladó általi utolsó sikeres bejelentkezés időbélyege.

- **Folyószámlák**
A viszonteladó által jelenleg kezelt felhasználói fiókok száma.

- **Maximális fiókok**
A viszonteladó által létrehozható felhasználói fiókok maximális száma.

- **Engedélyezett tervek**
Felsorolja a viszonteladó rendelkezésére álló tárhelycsomagokat, amelyeket hozzárendelhet a felhasználókhoz.

---

A viszonteladói felhasználók csak a root rendszergazda által hozzájuk rendelt funkciókhoz és fiókkezelő eszközökhöz férhetnek hozzá. Nem léphetik túl a viszonteladói beállításaikban meghatározott korlátokat.

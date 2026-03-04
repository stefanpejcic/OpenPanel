---
sidebar_position: 4
---

# Szolgáltatási korlátok

A **Szolgáltatási korlátok** szakasz lehetővé teszi a rendszergazdák számára, hogy megtekintsék és módosítsák a rendszerszolgáltatásokhoz lefoglalt CPU- és memória-erőforrásokat.

---

A korlátok frissítéséhez írjon be új értékeket a megfelelő szolgáltatáshoz tartozó mezőkbe.

> **Megjegyzés:** A szolgáltatást le kell állítani és újra kell indítani, hogy az új korlátok életbe lépjenek.

**Elérhető szolgáltatások**:

- nyitott panel
- Caddy
- mysql
- clamav
- redis
- köt 9
- ftp

---

A limitek terminálon keresztüli szerkesztéséhez módosítsa közvetlenül a `/root/.env` fájlt.

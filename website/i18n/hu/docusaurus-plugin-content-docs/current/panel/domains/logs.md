---
sidebar_position: 21
---

# Nyers hozzáférési naplók

A **Nyers hozzáférési naplók** funkció csak akkor érhető el az OpenPanel irányítópultján, ha a **Domain Logs** modul engedélyezve van a szerveren, és felhasználói fiókja rendelkezik a szükséges engedélyekkel.

A **Domain Logs** oldalon megtekintheti a **Caddy** webszerver által a domainhez létrehozott nyers hozzáférési naplókat.

A Caddy az SSL kezelésére, a proxykonfigurációk kezelésére és a forgalom továbbítására szolgál a konténeres szolgáltatásokhoz (például Varnish, Nginx, OpenResty, Apache).

Minden naplóbejegyzés részletes információkat tartalmaz, többek között:

* **Időbélyeg** – A kérés pontos időpontja
* **Client IP** – Az az IP-cím, amelyről a kérés származik
* **Naplószint** – A naplózás súlyossága (pl. információ, figyelmeztetés, hiba)
* **HTTP-módszer** – Kérési módszer (GET, POST, PUT stb.)
* **Állapotkód** – HTTP válaszkód (200, 301, 403, 500 stb.)
* **URI** – A kért elérési út
* **Logger** – Belső Caddy logger azonosító
* **Üzenet** – Részletek a kérés kezeléséről vagy a hibákról
* **Méret** – Méret kérése bájtban
* **Időtartam** – A Caddy által a kérelem feldolgozásához és a webszerverhez való továbbításához szükséges idő

Alapértelmezés szerint oldalanként 1000 találat jelenik meg. Ezt a korlátot a kiszolgáló rendszergazdája állíthatja be.

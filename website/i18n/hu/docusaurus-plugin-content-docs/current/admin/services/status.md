---
sidebar_position: 1
---

# Szolgáltatás állapota

A Szolgáltatás állapota szakaszban megtekintheti és szabályozhatja a kiszolgálón futó rendszerszolgáltatások és tárolók állapotát.

Ez a táblázat az egyes szolgáltatások főbb részleteit tartalmazza:

* **Szolgáltatás** – A szolgáltatás megjelenített neve.
* **Állapot** – Azt jelzi, hogy a szolgáltatás aktív vagy inaktív.
* **Verzió** – Aktuális verzió (ha elérhető).
* **Valódi név** – Belső szolgáltatásnév vagy tárolónév (pl. az OpenAdmin rendszergazdája).
* **Típus** – Azonosítja, hogy a szolgáltatás rendszerfolyamat vagy tároló.
* **Port** – A szolgáltatás által használt portok listája.
* **Monitoring** – Megmutatja, hogy a szolgáltatást aktívan figyelik-e és naplózzák-e.
* **Művelet** – A szolgáltatás indításának, leállításának vagy újraindításának lehetőségei.

## Szolgáltatások szerkesztése

A **Szolgáltatások szerkesztése** gombra kattintva testreszabhatja, hogy mely szolgáltatások jelenjenek meg és melyek legyenek kezelhetők ebből a szakaszból.

A szolgáltatások JSON formátumban vannak konfigurálva:

* **név** – A szolgáltatás megjelenített neve.
* **típus** – „rendszer” vagy „tároló”.
* **valós_név** – Belső szolgáltatás- vagy tárolóazonosító.

Az alapértelmezett OpenPanel-szolgáltatásokat a SentinelAI aktívan felügyeli, és megpróbálja automatikusan diagnosztizálni és újraindítani a meghibásodott szolgáltatásokat. Ha manuálisan állít le egy szolgáltatást, ne felejtse el letiltani ezt a megfigyelési funkciót az [OpenAdmin Notifications](/docs/admin/notifications/) oldalon.


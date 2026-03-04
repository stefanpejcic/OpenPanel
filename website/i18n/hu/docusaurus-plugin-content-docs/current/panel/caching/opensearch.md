---
sidebar_position: 5
---

# OpenSearch

Az [OpenSearch](https://opensearch.org) egy nyílt forráskódú kereső- és elemzőmotor, amely teljes mértékben kompatibilis az Elasticsearch programmal. Ideális:

- Teljes szöveges keresés
- Naplóelemzés
- Valós idejű adatfeltárás

Az OpenSearch szolgáltatást az **OpenPanel > Gyorsítótár > OpenSearch** oldalon kezelheti, lehetővé téve az alkalmazások és rendszerek gyors, méretezhető keresési képességeit.

---

## Állapot

Engedélyezze vagy tiltsa le az OpenSearch szolgáltatást.

- Megjelenik az aktuális állapot.
- Használja a kapcsolót a szolgáltatás **engedélyezéséhez** vagy **letiltásához**, ha szükséges.

---

## TCP

Az OpenSearch API eléréséhez szükséges kapcsolat részletei:

- **Szerver:** `opensearch`
- **Port:** `9200`

Használja ezeket a beállításokat olyan külső eszközökben, mint a "curl", a Kibana, a Grafana vagy az alkalmazások SDK-i az adatok lekérdezéséhez vagy kezeléséhez.

---

## Naplók

Kövesse nyomon a szolgáltatási tevékenységet és a hibakeresési problémákat az OpenSearch naplóinak megtekintésével.

- Kattintson a **Szolgáltatásnapló megtekintése** lehetőségre a szolgáltatásnaplófájl megnyitásához és a kimenet valós idejű ellenőrzéséhez.

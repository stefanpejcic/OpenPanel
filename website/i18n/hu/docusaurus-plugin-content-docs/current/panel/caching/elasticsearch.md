---
sidebar_position: 4
---

# Elasztikus keresés

Az [Elasticsearch](https://www.elastic.co/elasticsearch/) egy hatékony, nyílt forráskódú kereső- és elemzőmotor, amelyet gyakran használnak:

- Teljes szöveges keresés
- Naplóelemzés
- Valós idejű adatfeltárás

Az **OpenPanel > Caching > Elasticsearch** segítségével kezelheti az Elasticsearch szolgáltatást, lehetővé téve a gyors és méretezhető keresési funkciókat az alkalmazásokban és rendszerekben.

---

## Állapot

Engedélyezze vagy tiltsa le az Elasticsearch szolgáltatást.

- Megjelenik az aktuális állapot.
- Használja a kapcsolót a szolgáltatás **engedélyezéséhez** vagy **letiltásához**, ha szükséges.

---

## TCP

Az Elasticsearch API eléréséhez szükséges csatlakozási adatok:

- **Szerver:** `elasticsearch`
- **Port:** `9200`

Használja ezeket a beállításokat olyan külső eszközökben, mint a "curl", a Kibana, a Grafana vagy az alkalmazások SDK-i az adatok lekérdezéséhez vagy kezeléséhez.

---

## Naplók

Kövesse nyomon a szolgáltatási tevékenységet és a hibakeresési problémákat az ElasticSearch naplóinak megtekintésével.

- Kattintson a **Szolgáltatásnapló megtekintése** lehetőségre a szolgáltatásnaplófájl megnyitásához és a kimenet valós idejű ellenőrzéséhez.

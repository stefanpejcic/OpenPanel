---
sidebar_position: 3
---

# API hozzáférés

Az API-hozzáférési oldalon tesztelheti az API-hívásokat, megtekintheti a kérésekre/válaszokra vonatkozó példákat, figyelheti az API-naplókat, és kísérletezhet a végpontokkal egy egyszerű felületen.

Kezdésnek:

1. **API-hozzáférés engedélyezése**

Kapcsolja be az API-hozzáférés engedélyezése beállítást az API-kérések tesztelésére szolgáló felület aktiválásához.

3. **Töltse ki az igénylőlapot**

Miután engedélyezte, megjelenik egy űrlap a következő mezőkkel:

- **Módszer** – Válassza ki a HTTP-módszert.
- **URL** – Adja meg a hívni kívánt API-végpontot.
- **FELHASZNÁLÓNÉV** - Adja meg a rendszergazda felhasználónevét.
- **PASSWORD** - Adja meg a rendszergazdai jelszót.
- **TOKEN** – a felhasználónév és jelszó kombinációjával generált token.
- **ADATOK** – (Opcionális) Adja meg a végpont által igényelt adatokat vagy paramétereket.

5. **Kérés küldése**

Kattintson a Kérelem küldése gombra az API-hívás végrehajtásához.

7. **Az API-naplók és a Curl parancs megtekintése**

A kérés után a válasz, a curl parancs és a napló részletei jelennek meg alul. Ez magában foglalja a kérésinformációkat, a szerver válaszát, az állapotkódokat és az esetleges hibaüzeneteket.

A teljes API-referenciaért és a további végpontokért keresse fel részletes dokumentációnkat: https://dev.openpanel.com/openadmin-api/

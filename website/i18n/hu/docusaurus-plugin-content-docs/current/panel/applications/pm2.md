---
sidebar_position: 2
---

# Node.js és Python

Konténeres [Node.js](https://nodejs.org) és [Python](https://python.org/) alkalmazások hozhatók létre és kezelhetők az **OpenPanel Enterprise Edition** segítségével.

---

## Hozzon létre egy alkalmazást

Új Python vagy Node.js alkalmazás létrehozásához nyissa meg az **OpenPanel > AutoInstaller** lehetőséget, és válassza a **Python** vagy **NodeJS** lehetőséget.

![screenshot](https://i.postimg.cc/HmZh5ZMJ/new-tab.png)

A következő oldalon a következő beállításokat konfigurálhatja:

* **Név** – Az alkalmazás és a tároló neve az OpenPanelben.
* **Port** – Állítson be egyéni portot (például 3000 vagy 5000), ha az alkalmazás használ ilyet. Egyébként alapértelmezés szerint a 80-as port használatos.
* **Domain név / Almappa** – Az a tartomány (és opcionális almappa), ahol az alkalmazás nyilvánosan elérhető lesz.
* **Indítófájl** – Az indításkor a "node" vagy a "py" paranccsal végrehajtott fájl.
* **Egyéni indítási parancs** – Használjon egyéni indítási parancsot az alapértelmezett "node" vagy "py" helyett.
* **Típus** – Válasszon a Node.js vagy a Python közül.
* **Verzió** – Válassza ki a Docker Hub bármely elérhető verzióját.
* **Telepítés futtatása** – Az alkalmazás elindítása előtt futtassa az `npm install` vagy `pip install` parancsot.
* **CPU magok** – Az alkalmazáshoz kiosztott CPU magok száma.
* **Memória** – Az alkalmazáshoz lefoglalt memória mennyisége (GB-ban).

![képernyőkép](https://i.postimg.cc/x0PBW9qB/new-app.png)

Az űrlap kitöltése után kattintson a **Telepítés indítása** gombra.
A telepítési folyamat az űrlap alatt jelenik meg. Ha elkészült, átirányítjuk a kezelőoldalra, ahol megtekintheti az összes alkalmazását.

---

## Alkalmazások kezelése

Az alkalmazás létrehozása után az **OpenPanel > Site Manager** menüpontban kezelheti.

![screenshot](https://i.postimg.cc/vYbbVP6T/manage-apps.png)

Kattintson az alkalmazás neve melletti **Kezelés** lehetőségre a kezelési oldal megnyitásához.

![képernyőkép](https://i.postimg.cc/bzMFXdpg/single-app.png)

Ezen az oldalon olyan fontos részleteket tekinthet meg, mint például:

* **Képernyőkép** – Az alkalmazás tartományának előnézete.
* **Status** – A tároló aktuális állapota.
* **Verzió** – Node.js vagy Python verzió használatban.
* **CPU-korlát** – Konfigurált CPU-kiosztás.
* **Memory Limit** – Konfigurált memóriafoglalás.
* **Speed** – Google PageSpeed ​​Insights adatok a webhelyhez.
* **Fájlok** – A mappa jelenlegi elérési útja és mérete.
* **Tűzfal** – WAF (Web Application Firewall) állapota a tartományhoz (ha engedélyezve van).

Számos kezelési lehetőség közül választhat:

* **Műveletek** – A tároló indítása, leállítása vagy újraindítása.
* **Áttekintés** – Módosítsa az indítófájlt vagy parancsot, a munkakönyvtárat, a csomagtelepítési beállításokat (NPM/PIP), a verziót és az erőforráskorlátokat (CPU, memória, PID-k).
* **Csomagok telepítése** – Tekintse meg és kezelje a "package.json" vagy "requirements.txt" fájlt, és futtasson NPM/PNPM vagy PIP telepítéseket.
* **Logs** – A tárolónaplók megtekintése hibaelhárításhoz.
* **Eltávolítás** – Az alkalmazás törlése.

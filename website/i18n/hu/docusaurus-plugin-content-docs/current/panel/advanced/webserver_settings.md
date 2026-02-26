---
sidebar_position: 3
---

# Webszerver beállításai

A **WebServer Settings** felület lehetővé teszi az OpenPanel felhasználók számára, hogy szövegszerkesztővel közvetlenül szerkesszék az elsődleges webszerver konfigurációs fájlt.

Ezek a beállítások **globálisak** és **minden tartományra** vonatkoznak (azaz befolyásolják a virtuális gazdagép konfigurációját).

> ⚠️ **Fontos:** Bármilyen változtatás előtt mindig készítsen biztonsági másolatot a konfigurációs fájlról.

- **Apache** esetén szerkesztheti a "httpd.conf" fájlt.
- **Nginx** és **OpenResty** esetén szerkesztheti az "nginx.conf" fájlt.
- **OpenLitespeed** és **Litespeed** esetén szerkesztheti az "openlitespeed.conf" fájlt.

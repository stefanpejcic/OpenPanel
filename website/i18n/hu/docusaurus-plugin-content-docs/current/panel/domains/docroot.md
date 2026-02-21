---
sidebar_position: 7
---

# Docroot módosítása

A **Change Docroot** beállítás lehetővé teszi egyéni dokumentumgyökér (docroot) beállítását a domainhez.

Alapértelmezés szerint az OpenPanel a `/var/www/html/{domain}` könyvtárból szolgál ki webhelyeket. Előfordulhat azonban, hogy a haladó felhasználók módosítani szeretnék ezt a helyet – például egy alkönyvtárat (például `/var/www/html/{domain}/public) szeretnének kiszolgálni, vagy teljesen egy másik mappára szeretnének mutatni.

Ez akkor hasznos, ha:
- Több alkalmazás tárolása ugyanazon a domain alatt.
- Statikus tartalom kiszolgálása egy adott mappából.
- Olyan keretrendszer használata, amelyhez egy "public/" vagy "htdocs/" gyökér szükséges.

## Dokumentumgyökér módosítása

Domain dokumentumgyökérének módosítása az OpenPanelben:

1. Nyissa meg a **Domainek** oldalt.
2. Válassza ki a módosítani kívánt tartományt.
3. Kattintson a **Change Docroot** lehetőségre.
4. Adja meg az új elérési utat a `/var/www/html/` karakterlánccal kezdve.
5. Mentse el a változtatásokat.

Az OpenPanel automatikusan frissíti a webszerver konfigurációját, hogy tükrözze az új docrootot.

---
sidebar_position: 5
---

# Szerkessze a VHosts fájlt

A *Virtual Host* (vagy *vhost*) egy tartományspecifikus konfigurációs fájl, amelyet a webszerver használ.

A **VHosts fájl szerkesztése** funkció haladó felhasználók és tapasztalt rendszergazdák számára készült. Lehetővé teszi az egyes tartományok konfigurációs fájljainak közvetlen megtekintését és szerkesztését.

Különféle beállításokat módosíthat, például:

- Erőforrás korlátok
- Hozzáférési korlátozások
- Egyéni átirányítások
- Teljesítményhangolás

A webszervertől függően -**Nginx**, **OpenResty**, **OpenLiteSpeed** vagy **Apache** - eltérő szintaxist használnak, hivatalos útmutatásként tekintse meg:

- [Apache Virtual Host Documentation](https://httpd.apache.org/docs/current/vhosts/)
- [Nginx Admin Guide](https://docs.nginx.com/nginx/admin-guide/web-server/)
- [Az OpenLiteSpeed ​​útmutató konfigurálása](https://docs.openlitespeed.org/config/)


:::info
⚠️ **Fontos:** A módosítások elvégzése előtt mindig készítsen biztonsági másolatot a konfigurációs fájlról.
Míg az OpenPanel alapvető ellenőrzést hajt végre a mentéskor (beleértve a szintaktikai ellenőrzéseket és az automatikus visszaállítást hiba esetén), nem szabad kizárólag erre a biztosítékra hagyatkozni. A saját biztonsági másolatok karbantartása biztosítja, hogy szükség esetén manuálisan visszaállíthassa a konfigurációt.
:::


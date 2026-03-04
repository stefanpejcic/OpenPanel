---
sidebar_position: 5
---

# ImmunifyAV

Az [ImunifyAV](https://cloudlinux.zendesk.com/hc/en-us/articles/4716287786396-Imunify360-Standalone-installation-guide-with-integration-conf-examples) fokozza a szerver biztonságát azáltal, hogy lehetővé teszi a felhasználói webhelyek fájljainak rosszindulatú keresését.

> Megjegyzés: Az Imunify, annak védjegyei és minden kapcsolódó eszköz a [CloudLinux Zug GmbH](https://cloudlinux.com/) tulajdonát képezi.

## Telepítés

Kezdő verzió 1.5.4 - Az ImunifyAV alapértelmezés szerint az OpenPanel része.

Ha régebbi verziót használ, a telepítéshez futtassa:

```bash
opencli imunify install
```

Ez a parancs telepíti a PHP legújabb verzióját, az *imunify360-agent*-et, és konfigurálja a hozzáférést az OpenAdminon keresztül.

## Kezdje

Az Imunify grafikus felület elindításához használja:

```bash
opencli imunify start
```

## Bejelentkezés

Az **OpenAdmin > Biztonság > Imunify** menüpontból érheti el az Immunify grafikus felhasználói felületet.

## Kezelés

Az Imunify lehetővé teszi a felhasználói fájlok átvizsgálását és a rosszindulatú tartalom észlelését.

----

> További használati példákért lásd: [Útmutatók > ImunifyAV beállítása](/docs/articles/security/setup-imunifyav/).


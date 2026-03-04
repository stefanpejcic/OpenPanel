---
sidebar_position: 2
---

# FTP

Az OpenPanel fiókok FTP-felhasználókat hozhatnak létre, amelyek megosztják az FTP-szolgáltatást a kiszolgáló összes többi FTP-felhasználójával. Az FTP-hez a szerver IP-címe és az alapértelmezett 21-es port használatával csatlakozhat.

:::info
Az FTP az [OpenPanel Enterprise](/enterprise/) kiadásban érhető el, útmutató adminisztrátoroknak a [How to setup FTP in OpenPanel] (/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/)
:::

## Felhasználók megtekintése

Az FTP-kapcsolat információinak megtekintéséhez és a meglévő felhasználók kezeléséhez – például jelszavak módosításához, FTP-útvonaluk módosításához (a korlátozott könyvtárhoz) vagy fiókok törléséhez – lépjen az „OpenPanel > Fájlok > FTP” menüpontra.

## Felhasználó létrehozása

Új FTP-alfelhasználó létrehozásához nyissa meg a menü "OpenPanel > Fájlok > FTP" menüpontját, és kattintson a **Fiók hozzáadása** gombra. Megjelenik egy új rész, ahol beállíthatja az FTP felhasználónevet, jelszót és elérési utat.

Kattintson a **Fiók hozzáadása** gombra a módosítások mentéséhez.

:::info
Az FTP-alfelhasználóknak **ponttal (`.`) kell végződniük**, amelyet az OpenPanel-felhasználónév követ – például: `ftpuser.openpaneluser`. [FTP-felhasználónév követelmény](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Jelszó módosítása

Az FTP-alfiók jelszavának megváltoztatásához lépjen a menü "OpenPanel > Fájlok > FTP" pontjára, és kattintson a felhasználó mellett található **Jelszó módosítása** gombra. A következő oldalon állítsa be a felhasználó új jelszavát.

A mentéshez kattintson a **Jelszó módosítása** gombra.

:::info
Az FTP felhasználói jelszavaknak **legalább egy** nagybetűt (`A-Z`), kisbetűt (`a-z`), számjegyet (`0-9`) és speciális szimbólumokat kell tartalmazniuk. [FTP-jelszavak követelményei](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Útvonal módosítása

Egy FTP-alfiók elérési útjának módosításához (a címtárban a felhasználó csak korlátozott), lépjen a menü "OpenPanel > Fájlok > FTP" pontjára, és kattintson a felhasználó mellett található **Elérési út módosítása** gombra. A következő oldalon állítson be új elérési utat a felhasználó számára.

Kattintson az **Útvonal módosítása** gombra a mentéshez.

## Felhasználó törlése

FTP-alfiók törléséhez lépjen a menü "OpenPanel > Fájlok > FTP" pontjára, és kattintson a felhasználó mellett található **Törlés** gombra. A gomb a „Megerősítés” szövegre vált, kattintson rá ismét a törlés megerősítéséhez.

## Kapcsolatok megtekintése

Az aktív kapcsolatok (munkamenetek) megtekintéséhez lépjen a menü 'OpenPanel > Fájlok > FTP' pontjára, majd kattintson a 'Kapcsolatok megtekintése' linkre az oldal jobb felső sarkában.

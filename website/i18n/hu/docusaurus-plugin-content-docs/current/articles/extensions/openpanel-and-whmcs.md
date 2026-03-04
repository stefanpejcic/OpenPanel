# WHMCS

Az OpenPanel Enterprise Edition számlázási integrációkat tartalmaz a WHMCS-szel és a [FOSSBilling](/docs/articles/extensions/openpanel-and-fossbilling) szolgáltatással.

Az OpenPanel WHMCS modul lehetővé teszi a felhasználók számára, hogy integrálják a számlázási automatizálást az OpenPanel szerverükkel.

# OpenPanel

A WHMCS beállításához az OpenPanel szerver használatához kövesse az alábbi lépéseket:

## API engedélyezése

Először győződjön meg arról, hogy az API-hozzáférés engedélyezve van-e az „OpenAdmin > API” menüben, vagy az „opencli config get api” parancs futtatásával a terminálról:
![enable_api](https://i.postimg.cc/L6vwMQ4t/image.png)
Ha az API nincs engedélyezve, kattintson az "API hozzáférés engedélyezése" gombra vagy a terminál futtatásából
```bash
opencli config update api on
```

Javasoljuk, hogy hozzon létre új adminisztrátori felhasználót az API-hoz, új felhasználó létrehozásához lépjen az *OpenAdmin > OpenAdmin Settings* oldalra, és hozzon létre új adminisztrátori felhasználót, vagy a terminál futtatásából:
```bash
opencli admin new USERNAME_HERE PASSWORD_HERE
```

## Whitelist az OpenPanelen

Az OpenPanel szerveren győződjön meg arról, hogy az OpenAdmin 2087-es portja nyitva van az "OpenAdmin > Firewall" oldalon, vagy tegye a WHMCS-szerver IP-címét az engedélyezőlistára.
az IP-cím engedélyezése a terminál futtatásából:

```bash
csf -a WHMCS_IP_HERE
```
vagy ha UFS-t használ:
```bash
ufw allow from WHMCS_IP_HERE
```

## Hozzon létre tárhelycsomagot
Tárhelycsomagokat az OpenPanel és a WHMCS szervereken is létre kell hozni.
Az OpenPanel kiszolgálón jelentkezzen be az adminisztrációs panelbe, és az 'OpenAdmin > Tervek' oldalon hozzon létre tárhelycsomagokat, amelyeket hozzárendel a WHMCS felhasználóihoz.

# WHMCS

## Telepítse az OpenPanel WHMCS modult

Jelentkezzen be az SSH-ba a WHMCS-kiszolgálóhoz
Keresse meg a `whmcs_útvonala/modules/servers`
Futtassa ezt a parancsot egy új mappa létrehozásához, és töltse le a modult:
```bash
git clone https://github.com/stefanpejcic/openpanel-whmcs-module.git openpanel
```

## Whitelist a WHMCS-en

A WHMCS-kiszolgálón győződjön meg arról is, hogy a 2087-es port meg van nyitva, vagy adja meg az OpenPanel-kiszolgáló IP-címét az engedélyezőlistára.

## WHMCS modul beállítása

A WHMCS-ről navigáljon a következőhöz: *Rendszerbeállítások > Termékek és szolgáltatások > Szerverek*
![screenshot](https://i.postimg.cc/MHWpL3tc/image.png)
Kattintson az *Új kiszolgáló létrehozása* elemre, és a modul alatt válassza az **OpenPanel** lehetőséget, majd adja meg az OpenPanel szerver IP-címét, felhasználónevét és jelszavát az OpenAdmin panelhez:
![create_whmcs_group](https://i.postimg.cc/3Jh3nqWY/image.png)

## Hozzon létre tárhelycsomagot
A WHMCS-kiszolgálón először hozzon létre egy új csoportot, majd hozzon létre új terveket ezen a csoporton. Termékek létrehozásakor feltétlenül válassza ki az OpenPanel for Module és az újonnan létrehozott csoportot
![screenshot2](https://i.postimg.cc/NLvF4GSc/image.png)

## Teszt
Hozzon létre egy rendelést, és hozzon létre egy új rendelést az OpenPanel API teszteléséhez.

---
sidebar_position: 1
---

# Kezdje el

Az OpenAdmin rendszergazdai szintű felületet kínál, ahol hatékonyan kezelheti az olyan feladatokat, mint a felhasználók létrehozása és kezelése, tárhely-tervek beállítása, szolgáltatások engedélyezése és az OpenPanel beállításainak szerkesztése.

## Követelmények

Minimális követelmények:

- Egy üres teljes virtuális gép vagy csupasz fém szerver
- Minimum 1 GB RAM és 5 GB tárhely (4 GB RAM és 50 GB ajánlott)
- AMD64(x86_64) vagy ARM(AArch64) architektúra
- IPv4 cím

Támogatott operációs rendszerek:
- **Ubuntu 24.04** (ajánlott)
- **Debian 10, 11, 12, 13**
- **AlmaLinux 9.5 és 10** (ARM CPU-hoz a 9.5 ajánlott, a 10-ben ismert probléma van [#744](https://github.com/stefanpejcic/OpenPanel/issues/744))
- **RockyLinux 9.6, 10**
- **CentOS 9.5**

AlmaLinux 10 és RockyLinux 10 esetén át kell váltania az nftables-ról az iptables-ra. Lásd: [#1472](https://github.com/docker/for-linux/issues/1472) és [#745](https://github.com/stefanpejcic/OpenPanel/issues/745#issuecomment-3451272947).

:::info
Ha külső tűzfalat használ, a következő portokat kell megnyitni: `25` `53` `80` `443` `465` `993` `2083` `2087` `32768:60999`
:::

## Telepítés

Az OpenPanel VPS és bare-metal szerverekre egyaránt telepíthető.

A telepítési folyamat körülbelül 5 percet vesz igénybe. Az openpanel telepítéséhez kövesse az alábbi lépéseket:

<Tabs>
<TabItem value="openpanel-install-on-dedicated" label="Szkript telepítése" alapértelmezett>

1. Jelentkezzen be az új szerverére;
- rootként SSH-n keresztül vagy
- mint felhasználó sudo jogosultságokkal és írja be a "sudo -i" parancsot
2. Másolja és illessze be az openpanel telepítési parancsot a terminálba
```shell
bash <(curl -sSL https://openpanel.org)
```

A telepítő szkript támogatja az [opcionális zászlókat] (/install), amelyek segítségével konfigurálható az openpanel, kihagyható bizonyos telepítési lépések, vagy egyszerűen megjeleníthetők a hibakeresési információk.

Ha bármilyen hibát észlelt a telepítőszkript futtatása közben, kérjük, másolja és illessze be a telepítési naplófájlt [a közösségi fórumokba] (https://community.openpanel.org).

</TabItem>
<TabItem value="openpanel-install-on-cloud" label="Cloud">

[Amazon Web Services (AWS)](/docs/articles/install-update/install-on-aws)

[DigitalOcean](/docs/articles/install-update/install-on-digitalocean)

[Google Cloud Platform (GCP)](/docs/articles/install-update/install-on-google-cloud)

[Microsoft Azure](/docs/articles/install-update/install-on-microsoft-azure)

[Vultr](/docs/articles/install-update/install-on-vultr)
    
</TabItem>
<TabItem value="openpanel-install-on-other" label="Egyéb">

[CloudInit](/docs/articles/install-update/install-using-cloudinit)

[Lehetséges](/docs/articles/install-update/install-using-ansible)

[Virtualizor](/docs/articles/install-update/install-on-virtualizor)

</TabItem>
</Tabs>


## Telepítés utáni lépések

Az OpenPanel telepítése után javasolt lépések:
- [Az OpenAdmin panel elérése](/docs/articles/dev-experience/how-to-access-openadmin)
- [Domain és SSL konfigurálása az OpenPanelhez](/docs/admin/settings/general/#set-domain-for-openpanel)
- [Modulok (szolgáltatások) engedélyezése az OpenPanel UI-ban](/docs/admin/settings/modules/)
- [Egyéni névszerverek konfigurálása](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [tárhelycsomagok létrehozása](/docs/admin/plans/hosting_plans#create-a-plan)
- [Új felhasználói fiókok létrehozása](/docs/admin/accounts/users/#create-users)
- [E-mail-cím beállítása a figyelmeztetések fogadására](/docs/admin/notifications/#email-alerts)
- [Frissítési beállítások módosítása](/docs/admin/settings/updates)
- [Secure OpenPanel for Production Use](/docs/articles/security/securing-openpanel/)

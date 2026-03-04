---
sidebar_position: 3
---

# Zóna sablonok szerkesztése

Ez az interfész lehetővé teszi azoknak a DNS-zónafájl-sablonoknak a szerkesztését, amelyeket az OpenPanel az új tartományok zónáinak létrehozásakor használ.
Akkor hasznos, ha egyéni DNS-konfigurációra van szüksége.

![a zóna képernyőképének szerkesztése](/img/admin/dns_templates_admin.png)

- **IPv4-sablon** – Az új domainekhez hozzárendelt IPv4-címekhez használatos.
- **IPv6-sablon** – Az új domainekhez hozzárendelt IPv6-címekhez használatos.

---

## Elérhető sablonváltozók

A következő változókat veheti fel a DNS-zóna sablonjaiba:

| Változó | Leírás |
|----------------|----------------------------------------------------------------------------|
| `{ns1}` | Az elsődleges névszerver gazdagépneve (az NS rekordban használatos).                |
| `{ns2}` | A másodlagos névszerver gazdagépneve.                                      |
| `{ns3}` | A harmadlagos névszerver gazdagépneve.                                       |
| `{ns4}` | A kvaterner névszerver gazdagépneve.                                     |
| `{server_ip}` | A tartomány IP-címe (IPv4 vagy IPv6, a sablontól függően). |
| `{rpemail}` | Kapcsolatfelvételi e-mail cím az OpenAdmin szolgáltatásból, vagy "root@domain".                    |

---

## Alapértelmezések visszaállítása

A sablon alapértelmezett verziójának visszaállítása:

1. Kattintson az **Alapértelmezés visszaállítása** gombra a megfelelő szakaszhoz.
2. Ezután kattintson a **Fájlok mentése** gombra a módosítások alkalmazásához.

> ⚠️ Az alapértelmezések a hivatalos GitHub-tárhelyből származnak. Mentés előtt mindig nézze át.

# Frissítés kihagyása

Az OpenPanel az Ön preferenciáitól függően **automatikusan** vagy **manuálisan** frissíthető.

A frissítési folyamat a következő műveleteket hajtja végre:

* Lekéri a legújabb OpenPanel UI Docker-képet a Docker Hubról
* Frissíti az OpenAdmin felhasználói felületet a legújabb GitHub-kiadásból
* Frissíti az OpenCLI parancsokat a GitHubból
* Végrehajt minden frissítési szkriptet, amely a konfiguráció módosításához biztosított
* Ellenőrzi és telepíti a frissítéseket a rendszercsomagkezelőn keresztül
* Frissítések docker komponálni a legújabb verzióra
* Ellenőrzi, hogy valamelyik kernelfrissítéshez újra kell-e indítani
* Frissítés utáni szkripteket hajt végre, ha azt a rendszergazda biztosította

Ha az **automatikus frissítés** és az **automatikus javítás** opció is engedélyezve van, a frissítések automatikusan alkalmazásra kerülnek az új verzió kiadásakor.

Azonban **meghatározott verziókat** kihagyhat, ha felsorolja őket a következő fájlban:

```
/etc/openpanel/upgrade/skip_versions
```

### Mikor kell használni?

Ez például akkor hasznos, ha egy változásnapló bejelenti egy olyan szolgáltatás eltávolítását, amelyre támaszkodik. Ennek a verziónak a kihagyásával biztosíthatja, hogy a jelenlegi beállítás változatlan maradjon.

### Verzió kihagyása

1. Hozza létre a könyvtárat (ha még nem létezik):

```bash
mkdir -p /etc/openpanel/upgrade/
```

2. Adja hozzá a kihagyni kívánt verziószámot:

```bash
echo 1.4.9 >> /etc/openpanel/upgrade/skip_versions
```

Most, amikor kiadják az `1.4.9-es verziót, a rendszer figyelmen kívül hagyja még akkor is, ha az automatikus frissítések engedélyezve vannak.

Tetszőleges számú verziót adhat hozzá:

```bash
echo 1.5.0 >> /etc/openpanel/upgrade/skip_versions
echo 1.5.1 >> /etc/openpanel/upgrade/skip_versions
```

Ennyi – egyszerű és hatékony vezérlés az OpenPanel frissítési folyamata felett.

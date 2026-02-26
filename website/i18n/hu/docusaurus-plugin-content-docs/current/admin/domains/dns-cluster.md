---
sidebar_position: 3
---

# DNS-fürt

A DNS-fürtözés lehetővé teszi a DNS-rekordok szinkronizálását egy OpenPanel-kiszolgálóról más gépekre.

A funkció használatához **a fürt összes gépén a BIND9-et kell futtatnia** – vagy az OpenPanel telepítésén keresztül, vagy önálló szolgáltatásként/tárolóként.

---

## Klaszterezés engedélyezése

A szolgáltatás aktiválásához kattintson a **DNS-fürtözés engedélyezése** lehetőségre.

---

### Slave szerverek hozzáadása

A szolga BIND9-kiszolgáló beállításához kövesse a [How-to Guides > DNS Clustering](/docs/articles/domains/how-to-setup-dns-cluster-in-openpanel/) utasításait.

Ha a rabszolga készen áll:

1. Adja meg a segédkiszolgáló IP-címét a főkiszolgálón lévő űrlapon.
2. Az OpenAdmin megpróbál csatlakozni és azonnal szinkronizálni a DNS-zónákat.
3. A felhasználók által végrehajtott összes jövőbeni tartomány- és DNS-módosítás automatikusan átkerül a szolgakiszolgálóra.

> A beállítás működésének ellenőrzéséhez ellenőrizze, hogy a DNS-zónák megjelennek-e a slave szerveren.

---

### Távolítsa el a Slave Servert

Jelenleg a slave szerverek **csak a terminálon keresztül távolíthatók el**.

---

## A fürtözés letiltása

Kattintson a **DNS-fürtözés letiltása** lehetőségre a funkció kikapcsolásához.

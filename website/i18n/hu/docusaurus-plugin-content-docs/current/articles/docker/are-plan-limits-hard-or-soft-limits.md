# Kemény határok

Minden felhasználói terv rendelkezik rögzített erőforrás-használati korlátokkal – ezek **szigorú korlátok**, ami azt jelenti, hogy a felhasználók semmilyen körülmények között nem léphetik túl azokat.

- A **Disk** és **Inode** [kvóták] (https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/storage_administration_guide/ch-disk-quota) a Linux kernel által betartandó szigorú korlátozások.
- A **CPU** és **memória** lefoglalásokat is szigorú korlátozásként kényszeríti a Docker. Ezeket az erőforrásokat szigorúan korlátozzák a túlzott felhasználás megelőzése érdekében.

> **Megjegyzés:** A CPU-használati százalékok félrevezetőek lehetnek több CPU maggal rendelkező rendszereken. Például, ha egy szolgáltatáshoz 1,5 CPU-mag van hozzárendelve, akkor a maximális kihasználtsága **150%**, nem **100%** formában jelenhet meg.
> További részletekért lásd: [Docker CLI Issue #2134](https://github.com/docker/cli/issues/2134)


# Frissítések engedélyezése

- Az "autopatch" opció lehetővé teszi a rendszergazdának, hogy automatikusan frissítse az OpenPanel kisebb verzióit. A KISEBB verziók csak biztonsági frissítéseket és hibajavításokat tartalmaznak.
- Az "autoupdate" opció lehetővé teszi a rendszergazdának, hogy engedélyezze vagy letiltja a főbb verziók automatikus frissítését. A FŐ verziók új funkciókat adnak hozzá visszafelé kompatibilis módon.

<Tabs>
<TabItem value="openadmin-admin-updates" label="OpenAdminnal" alapértelmezett>

Az automatikus frissítések engedélyezéséhez nyissa meg az **OpenAdmin > Beállítások > Frissítési beállítások** lehetőséget, és állítsa az „Automatikus frissítés” beállítást „Mindkettő” értékre:

![openadmin frissítési beállítások](/img/admin/openadmin_set_update_preferences.png)

</TabItem>
<TabItem value="CLI" label="OpenCLI-vel">

A terminálról érkező automatikus frissítések engedélyezéséhez használja a következő parancsokat:

```bash
opencli config update autoupdate yes
```

```bash
opencli config update autopatch yes
```
</TabItem>
</Tabs>

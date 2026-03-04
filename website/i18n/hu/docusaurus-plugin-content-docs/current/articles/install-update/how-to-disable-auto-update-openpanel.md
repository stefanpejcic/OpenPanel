# Frissítések letiltása

<Tabs>
<TabItem value="openadmin-admin-updates" label="OpenAdminnal" alapértelmezett>

Az automatikus frissítések letiltásához nyissa meg az **OpenAdmin > Beállítások > Frissítési beállítások** lehetőséget, és állítsa az „Automatikus frissítés” beállítást „Soha” értékre:

![openadmin frissítési beállítások](/img/admin/openadmin_set_update_preferences.png)

</TabItem>
<TabItem value="CLI" label="OpenCLI-vel">

A terminálról érkező automatikus frissítések letiltásához használja a következő parancsokat:

```bash
opencli config update autoupdate no
```

```bash
opencli config update autopatch no
```
</TabItem>
</Tabs>

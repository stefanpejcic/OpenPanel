### Enable Automatic Updates for OpenPanel

- `autopatch` option allows Administrator to automatically update OpenPanel to minor versions. MINOR versions include only security updates and bug fixes.
- `autoupdate` option allows Administrator to enable or disable automatic updates to major versions. MAJOR versions add new functionality in a backward compatible manner.

<Tabs>
  <TabItem value="openadmin-admin-updates" label="With OpenAdmin" default>

To enable automatic updates, navigate to **OpenAdmin > Settings > Update Preferences** and change 'Update automatically' option to 'Both':

![openadmin update preferences](/img/admin/openadmin_set_update_preferences.png)

  </TabItem>
  <TabItem value="CLI" label="With OpenCLI">

To enable automatic updates from the terminal use commands:

```bash
opencli config update autoupdate yes
```

```bash
opencli config update autopatch yes
```
  </TabItem>
</Tabs>

# Disable Updates

<Tabs>
  <TabItem value="openadmin-admin-updates" label="With OpenAdmin" default>

To disable automatic updates, navigate to **OpenAdmin > Settings > Update Preferences** and change 'Update automatically' option to 'Never':

![openadmin update preferences](/img/admin/openadmin_set_update_preferences.png)

  </TabItem>
  <TabItem value="CLI" label="With OpenCLI">

To disable automatic updates from the terminal use commands:

```bash
opencli config update autoupdate no
```

```bash
opencli config update autopatch no
```
  </TabItem>
</Tabs>

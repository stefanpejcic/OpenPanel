# Downgrade License

OpenPanel is available in two editions:

- **Community** - a free hosting control panel for Debian and Ubuntu OS, suitable for VPS and private use.
- **Enterprise** - offers advanced features for user isolation and management, suitable for web hosting providers.

More information: [OpenPanel Community VS Enterprise edition](/enterprise)

## Downgrade

:::info
Downgrading from the Enterprise to the Community edition will immediately remove all Enterprise features and modules from the OpenPanel UI. Existing enterprise services - such as FTP and email - will continue running, but you will no longer have the ability to manage them from the OpenPanel interface.
:::

To downgrade from Enterprise edition to Community:

- From OpenAdmin:
  Navigate to **OpenAdmin > License** and click on the 'Downgrade' button:
  
  ![remove license key](/img/guides/downgrade_license.png)

- From terminal:
  ```bash
  opencli license delete
  ```

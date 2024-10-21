# Upgrading to OpenPanel Enterprise and activating License

OpenPanel is available in two editions:

- **Community** - a free hosting control panel for Debian and Ubuntu OS, suitable for VPS and private use.
- **Enterprise** - offers advanced features for user isolation and management, suitable for web hosting providers.

More information: [OpenPanel Community VS Enterprise edition](/beta)

## Purchase License

To purchase a license for the OpenPanel Enterprise edition [click here](https://my.openpanel.com/cart.php?a=add&pid=1).

After payment, license key is automatically issued and you will see it under the Services page:

![license key](/static/img/guides/add_license.png)

## Add License Key

### To existing installation:

- From OpenAdmin:
  If you already have OpenPanel Community edition installed, navigate to **OpenAdmin > Try Enterprise** and add the key into the form:
  
  ![add license key](/static/img/guides/add_key.png)

- From terminal:
  ```bash
  opencli license enterprise-XXX
  ```
  Replace `enterprise-XXX` with your license key.
  
  ![opencli license](/static/img/guides/key_command.png)

### To new installation:

If you will be installing OpenPanel Enterprise on another server, copy the install command from your service page and paste it into new server:

![install command](/static/img/guides/key_code.png)

# Upgrading to OpenPanel Enterprise and activating License

OpenPanel is available in two editions:

- **Community** - a free hosting control panel for Debian and Ubuntu OS, suitable for VPS and private use.
- **Enterprise** - offers advanced features for user isolation and management, suitable for web hosting providers.

More information: [OpenPanel Community VS Enterprise edition](https://openpanel.co/beta)

## Purchase License

To purchase a license for the OpenPanel Enterprise edition [click here](https://my.openpanel.co/cart.php?a=add&pid=1).

After payment, license key is automatically issued and you will see it under the Services page:

![license key](https://i.postimg.cc/0ybCkTQX/2024-08-04-15-40.png)

## Add License Key

### To existing installation:

- From OpenAdmin:
  If you already have OpenPanel Community edition installed, navigate to **OpenAdmin > Try Enterprise** and add the key into the form:
  
  ![add license key](https://i.postimg.cc/P5VTZwdr/2024-08-04-15-46.png)

- From terminal:
  ```bash
  opencli license enterprise-XXX
  ```
  Replace `enterprise-XXX` with your license key.
  
  ![opencli license](https://i.imgur.com/rxqjFsy.png)

### To new installation:

If you will be installing OpenPanel Enterprise on another server, copy the install command from your service page and paste it into new server:

![install command](https://i.postimg.cc/3xFDH9jf/2024-08-04-15-43.png)

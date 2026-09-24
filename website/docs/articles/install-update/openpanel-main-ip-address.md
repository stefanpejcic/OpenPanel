---
sidebar_label: "Main IP in OpenPanel"
---

# How the Server’s Main IP Address Is Chosen (and How to Check It)

The network configuration of the server determines the server's main IP address. The Main IP address is **not** configurable through OpenPanel.

A network administrator or your hosting provider can only modify it. The main IP address is the IP address that the kernel selects as the default outbound IP address. For this reason, this is the IP address that the OpenPanel licensing system uses as the IP address of your server.

To check the current IP address:

```bash
curl -4Lk https://ip.openpanel.com
```

---

For assigning a specific IP address to a user, see: [How to set a Dedicated IP address for a user](/docs/articles/accounts/set-dedicated-ip-address-for-user)

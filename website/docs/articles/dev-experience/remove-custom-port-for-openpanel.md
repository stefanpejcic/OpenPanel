---
sidebar_label: "Removing the Port for OpenPanel"
---

# How to Access the User Panel Without a Port Number

By default, the OpenPanel interface runs in its own container, while all other websites run on Caddy. This setup ensures that the user panel remains accessible even if the main web server is down.

If you want to remove the port number for OpenPanel and access it directly via your domain (e.g., `https://yourdomain.com/`), follow these steps:

1. Set port to `443`:
   ```bash
   opencli port set 443
   ```

The command opens the port on the firewall, reloads Caddy and recreates the `openpanel` container, so there is nothing else to do.

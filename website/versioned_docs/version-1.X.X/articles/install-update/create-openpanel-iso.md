# Creating an OpenPanel ISO Image

This guide is intended for server providers offering Cloud, VPS, or Dedicated servers who want to enable fast image deployment with OpenPanel.

Creating an OpenPanel ISO image is straightforward. Follow these steps:

1. **Install and configure OpenPanel** on your server.
2. **Power off the server.**
3. **Create an image** of the server.

> ⚠️ **Note:** Your server’s hostname and email settings are preserved in the image. You will need to update these after deploying the image.

Some settings can be configured automatically during deployment, like the hostname. Others, such as email and database credentials, must be updated manually.

### Example: Update panel domain

```bash
opencli domain YOUR_DOMAIN
```

Additionally, it is recommended to **change the MySQL root password** and update it in `/etc/my.cnf` to ensure security.

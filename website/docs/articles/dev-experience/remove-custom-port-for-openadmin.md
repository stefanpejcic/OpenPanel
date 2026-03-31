#  Removing the Port for OpenAdmin

By default, the OpenAdmin interface runs on a Gunicorn web server, while all other websites run on Caddy. This setup ensures that the admin panel remains accessible even if the main web server is down.

If you want to remove the port number for OpenAdmin and access it directly via your domain (e.g., `https://yourdomain.com/`), follow these steps:

1. Create a marker file to disable the default port:
   ```bash
   touch /root/disable_2087_port
   ```
3. Restart OpenAdmin to apply the change:
   ```bash
   service admin restart
   ```

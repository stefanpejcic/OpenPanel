#  Removing the Port for OpenPanel

By default, the OpenPanel interface runs on a Gunicorn web server, while all other websites run on Caddy. This setup ensures that the admin panel remains accessible even if the main web server is down.

If you want to remove the port number for OpenPanel and access it directly via your domain (e.g., `https://yourdomain.com/`), follow these steps:

1. Set port to `443`:
   ```bash
   opencli port set 443
   ```
3. Restart services to apply the change:
   ```bash
   docker restart openpanel caddy
   ```

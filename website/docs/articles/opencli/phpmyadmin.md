# phpMyAdmin

`opencli phpmyadmin` views and sets the domain (or IP) used to access phpMyAdmin.

View the current phpMyAdmin link:
```bash
opencli phpmyadmin
```

Set a domain for phpMyAdmin access (served over HTTPS at `https://DOMAIN/PORT/`):
```bash
opencli phpmyadmin set <domain>
```

Use the server's IP instead of a domain (served over HTTP at `http://IP:PORT`):
```bash
opencli phpmyadmin set ip
```

# phpMyAdmin

`opencli phpmyadmin` views and sets the domain (or IP) used to access phpMyAdmin.

View the domain currently used for phpMyAdmin (empty if the IP is used):
```bash
opencli phpmyadmin
```

<details>
  <summary>Example output</summary>

```bash
# opencli phpmyadmin
pma.example.com
```
</details>

Set a domain for phpMyAdmin access (served over HTTPS at `https://DOMAIN/PORT/`):
```bash
opencli phpmyadmin set <domain>
```

<details>
  <summary>Example output</summary>

```bash
# opencli phpmyadmin set pma.example.com
Creating proxy for the pma.example.com domain..
Added new PHPMYADMIN domain block.
Updating PMA_ABSOLUTE_URI for all existing users...
phpmyadmin
Restarting running phpmyadmin service for podman context: stefan
Updating PMA_ABSOLUTE_URI in template for new users..
Done
```
</details>

Use the server's IP instead of a domain (served over HTTP at `http://IP:PORT`):
```bash
opencli phpmyadmin set ip
```

<details>
  <summary>Example output</summary>

```bash
# opencli phpmyadmin set ip
Updating PMA_ABSOLUTE_URI for all existing users...
Updating PMA_ABSOLUTE_URI in template for new users..
Done
```
</details>


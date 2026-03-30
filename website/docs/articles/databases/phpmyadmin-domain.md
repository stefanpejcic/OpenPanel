# Custom Domain for phpMyAdmin

In **OpenPanel**, each user has isolated services, including a dedicated **phpMyAdmin** instance accessible via a unique port.

By default, a user’s phpMyAdmin instance is accessible at:

```
http://IP:PORT
```

> **Tip:** The assigned address is visible in OpenPanel under **MySQL > phpMyAdmin**.

---

## Setting a Custom Domain for All Users

You can configure a single custom domain for all users’ phpMyAdmin instances. Users will then access phpMyAdmin via:

```
https://your-domain:PORT
```

**Example:**

| User   | Address                        |
| ------ | ------------------------------ |
| User A | https://phpmyadmin.my-domain.com/37746/ |
| User B | https://phpmyadmin.my-domain.com/37752/ |
| User C | https://phpmyadmin.my-domain.com/37778/ |

### Steps

#### 1. Add Your Domain in Caddy

Edit the Caddy configuration file:

```bash
nano /etc/openpanel/caddy/Caddyfile
```

Add a section for your phpMyAdmin domain (example: `phpmyadmin.my-domain.com`):

```caddy
# START PHPMYADMIN DOMAIN #
phpmyadmin.my-domain.com: {
    @dynamicPort path_regexp port ^/(\d+)(/.*)?$
    handle @dynamicPort {
        uri strip_prefix /{http.regexp.port.1}
        reverse_proxy localhost:{http.regexp.port.1}
    }

    handle {
        respond "Please use your unique phpmyadmin path shown on OpenPanel > MySQL > phpMyAdmin page."
    }
}
# END PHPMYADMIN DOMAIN #
```

![example](https://i.postimg.cc/MH8MzP4L/caddyfile.png)


Next, create an empty file for Caddy to manage SSL automatically:

```bash
touch /etc/openpanel/caddy/domains/phpmyadmin.my-domain.com.conf
```

Restart the Caddy service:

```bash
docker restart caddy
```

---

#### 2. Configure `PMA_ABSOLUTE_URI` for New Users

Edit the `.env` file for new users:

```bash
nano /etc/openpanel/docker/compose/1.0/.env
```

Under `# PHPMYADMIN`, add:

```bash
PMA_ABSOLUTE_URI="https://phpmyadmin.my-domain.com/{PMA_PORT}/"
```

![example](https://i.postimg.cc/8cVJTK3X/compose.png)

Save the file.

From now on, all new users will have phpMyAdmin accessible via your custom domain. This will also be reflected in the OpenPanel UI.

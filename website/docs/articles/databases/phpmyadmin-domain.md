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
| User A | https://phpmyadmin.my-domain.com:37746 |
| User B | https://phpmyadmin.my-domain.com:37752 |
| User C | https://phpmyadmin.my-domain.com:37778 |

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
    @withPort {
        host phpmyadmin.my-domain.com
    }
    reverse_proxy localhost:{http.request.port}
}
# END PHPMYADMIN DOMAIN #
```

![caddyfile](https://i.postimg.cc/G2K9YrBC/caddy-pma.png)

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
PMA_ABSOLUTE_URI="http://phpmyadmin.my-domain.com:{PMA_PORT}/"
```

Save the file.

![env file](https://i.postimg.cc/RCjqHPF8/pma-env.png)

From now on, all new users will have phpMyAdmin accessible via your custom domain. This will also be reflected in the OpenPanel UI.

---

## Allow Users to Set Their Own Custom Domains

Users can choose their own domains for phpMyAdmin instead of using IP:PORT.

**Example:**

| User   | Address                           |
| ------ | --------------------------------- |
| User A | http://IP:PORT                   |
| User B | https://phpmyadmin.their-domain.com:37752 |
| User C | http://phpmyadmin-awesome.net:37778      |

### Steps

1. Enable the **edit_vhost** module in **OpenAdmin > Settings > Modules** :

   ![module](https://i.postimg.cc/c1DBGGFw/vhost-module.png)

   and ensure it’s included in the feature sets used by hosting plans:
   ![feature](https://i.postimg.cc/PJFWW1Lx/feature.png)

3. Users will see an **Edit (pencil)** button on **MySQL > phpMyAdmin**.
   ![edit](https://i.postimg.cc/L4y3FFxz/user-pma.png)


5. To configure a custom domain, users must:

   1. Add the domain under **Domains > New**.
   2. Edit the VirtualHost for the domain to proxy traffic to `phpmyadmin:80`.
   3. Set the domain in **MySQL > phpMyAdmin**.
      ![edit](https://i.postimg.cc/gdpSRYRT/user-pma2.png)
   4. Restart the phpMyAdmin service.

Once completed, phpMyAdmin is accessible via the user’s chosen domain.

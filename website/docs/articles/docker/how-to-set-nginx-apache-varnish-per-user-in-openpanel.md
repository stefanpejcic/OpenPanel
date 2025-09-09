# Apache, Nginx, OpenResty, OpenLiteSpeed, and Varnish

Configuring Nginx, Apache, OpenResty, OpenLiteSpeed, and Varnish per user in OpenPanel.

When creating a user, you can set the webserver type to be used:

- Apache
- Apache + Varnish
- Nginx
- Nginx + Varnish
- OpenResty
- OpenResty + Varnish
- OpenLitespeed
- OpenLitespeed + Varnish


**Web servers comparison**:

| Feature / Webserver              | Apache                             | Nginx                                    | OpenResty                                | OpenLiteSpeed                                    |
| -------------------------------- | ---------------------------------- | ---------------------------------------- | ---------------------------------------- | ------------------------------------------------ |
| **Varnish Support**              | Yes                              | Yes                                    | Yes                                    | Yes                                            |
| **.htaccess Support**            | Full                             | None (requires editing VHost files)    | None (requires editing VHost files)             | Full ([with some limitations](https://docs.openlitespeed.org/config/rewriterules/))                   |
| **Multiple PHP Versions**        | Yes                        | Yes            | Yes            | No, just 1 version                            |
| **Performance (static content)** | Moderate                           | High                                     | High                                     | High                                             |
| **Performance (dynamic PHP)**    | Moderate                           | High                                     | High                                     | Very High                                        |
| **Ease of User Overrides**       | .htaccess or VHost Editor | Only via VHost Editor       | Only via VHost Editor          | .htaccess or VHost Editor      |
| **Memory Footprint**      | Moderate (~50–100 MB per process) | Low (~10–30 MB per worker)              | Low (~15–40 MB per worker)              | Moderate (~30–60 MB per process)                |
| **Docker Size**            | ~200 MB                           | ~120 MB                                 | ~150 MB                                 | ~1000 MB                                         || **Recommended Use Cases**        | Legacy apps, .htaccess-heavy sites | High-performance static or dynamic sites | Advanced Lua scripting, high concurrency | High-performance PHP hosting, WordPress, Laravel |

> **Notes:**
>
> * Memory / disk footprint values are approximate and depend on configuration and traffic.
> * Varnish can offload static content, reducing webserver load significantly.
> * With 'docker' module, users can switch webserver type anytime via **Change Webserver Type** page.
> * OpenLiteSpeed is generally best for high-performance PHP apps that need .htaccess support.

---

Administrators can select the default webserver to be used from [*OpenAdmin > Settings > User Defaults*](/docs/admin/settings/defaults/) page.

---

## Apache

[Apache](https://httpd.apache.org/) can be assigned either to all new users by default or to individual users during account creation.

### Set for a Single User

To set Apache for a single user during creation:

1. Open the **New User** form.
2. Under the **Advanced** section, select **Apache** as the webserver.

![apache for user](/img/guides/apache.gif)

> **Tip:** To enable Varnish for this user, simply turn **Varnish Cache ON** on the same page.

### Set as Default

To make Apache the default for all newly created users:

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **Apache** as the default webserver.


---

## Nginx

[Nginx](https://httpd.apache.org/) can be assigned either to all new users by default or to individual users during account creation.

### Set for a Single User

1. Open the **New User** form.
2. Under **Advanced**, select **Nginx** as the webserver.

![nginx for user](/img/guides/nginx.gif)

> **Tip:** Enable Varnish Cache ON if you want to use **Nginx + Varnish**.

### Set as Default

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **Nginx** as the default webserver.

---

## OpenResty

[OpenResty](https://openresty.org/) can be assigned either to all new users by default or to individual users during account creation.


### Set for a Single User

1. Open the **New User** form.
2. Under **Advanced**, select **OpenResty**.

![openresty for user](/img/guides/openresty.gif)

> **Tip:** Turn **Varnish Cache ON** for **OpenResty + Varnish**.

### Set as Default

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **OpenResty** as default.

---

## OpenLiteSpeed

[OpenLiteSpeed](https://openlitespeed.org/) can be assigned either to all new users by default or to individual users during account creation.

### Set for a Single User

1. Open the **New User** form.
2. Under **Advanced**, select **OpenLiteSpeed**.

![openlitespeed for user](/img/guides/openlitespeed.gif)

> **Tip:** Enable **Varnish Cache ON** to combine OpenLiteSpeed with Varnish.

### Set as Default

* Go to [**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/) and select **OpenLiteSpeed** as default.


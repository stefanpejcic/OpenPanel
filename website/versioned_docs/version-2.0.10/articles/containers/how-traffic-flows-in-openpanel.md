# How Web Traffic Flows with User Containers

Caddy is the **entry point** for all incoming web traffic.
When the first domain is added, Caddy automatically starts and begins listening on ports **80** (HTTP) and **443** (HTTPS).

Caddy is responsible for:

* **SSL certificates & renewals**
* **Redirects (HTTP → HTTPS)**
* **Web Application Firewall (CorazaWAF)**
* **Reverse proxying to user webservers**

---

## Per-User Webservers

Each user runs their own isolated **webserver instance** on a unique local port.
Supported webservers include:

* **Nginx**
* **Apache**
* **OpenResty**
* **OpenLiteSpeed**
* **Varnish**

Caddy forwards requests to the correct user’s webserver, based on domain configuration.

---

## Application Containers

The user’s webserver then proxies requests into **application containers**, such as:

* **PHP-FPM (multiple versions supported)**
* **Node.js**
* **Python / Django / Flask**
* **Docker containers**

Example flow for a PHP app:

```
Client → Caddy → Nginx → PHP-FPM container
```

Example flow for a Node.js app:

```
Client → Caddy → Nginx → Node.js container
```

---

## Handling Multiple PHP Versions

Each user runs its own **PHP-FPM container version**, ensuring compatibility with different frameworks or legacy apps. Multiple PHP versions (7.4, 8.0, 8.1, 8.2, 8.3, etc.) can coexist safely for a userL

Example:

```
Site A
Client → Caddy → Nginx → PHP-FPM-8.2

Site B
Client → Caddy → Nginx → PHP-FPM-7.4

Site C
Client → Caddy → Nginx → PHP-FPM-8.3
```
---

## Using Varnish (Optional)

If a user enables **Varnish caching**, it sits between Caddy and the webserver:

```
Client → Caddy → Varnish → Webserver → php container
```

This allows caching and performance acceleration before hitting the backend.

---

## Example Flows

### PHP (Nginx + PHP-FPM-8.2)

```
Client → Caddy          (SSL, WAF, redirects)  
       → Nginx          (per-user webserver)  
       → PHP-FPM-8.2    (specific PHP version)
```

### PHP (with Varnish + PHP-FPM-8.1)

```
Client → Caddy          (SSL, WAF, redirects)
       → Varnish        (caching)  
       → Nginx          (per-user webserver)  
       → PHP-FPM-8.1    (specific PHP version)
```

### Node.js

```
Client → Caddy          (SSL, WAF, redirects)
       → Nginx          (per-user webserver)  
       → Node.js        (user’s container)
```

### OpenLiteSpeed (with built-in LSPHP)

```
Client → Caddy          (SSL, WAF, redirects)  
       → OpenLiteSpeed  (VHost + lsphp runtime)
```

---

✅ **In short:**
**Caddy is always the first entry point** → routes traffic to a user’s **webserver** → which proxies requests to the user’s **application container (PHP version, Node.js, etc.)**.

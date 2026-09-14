# Network Isolation in OpenPanel

OpenPanel uses containers to isolate services for each user. These containers are then further segregated into local networks, providing an additional layer of isolation.

This setup ensures that containers within the same network can communicate securely, without exposing ports to the public internet.

## Available Networks

Containers can be assigned to one of three networks:

* **`db`** – containers that need database access
* **`www`** – containers that need webserver or caching access
* **`no network`** – containers with no access to other containers

Communication between containers in the same network can be done using container names as hostnames. For example, PHP containers can connect to MySQL simply by using `mysql` or `mariadb` as the hostname.

---

## `db` Network

The **db** network is designed for services that need access to databases.

By default, this network includes:

* MySQL
* MariaDB
* phpMyAdmin
* PostgreSQL
* PgAdmin
* Tor
* PHP-FPM services

If you are adding a custom service (container) that needs to connect to a database, you should attach it to the `db` network.

---

## `www` Network

The **www** network is used for services that need access to webservers, caching, or search engines.

By default, this network includes:

* OpenLiteSpeed
* Apache
* Nginx
* Varnish
* Tor
* Redis
* Memcached
* Elasticsearch
* OpenSearch
* PHP-FPM services

If you are adding a custom service (container) that needs to communicate with a webserver, Redis, or Memcached, you should attach it to the `www` network.

---

## `no network`

Containers without a network are completely isolated and cannot communicate with other containers.

These containers are typically only accessible through the terminal or the OpenPanel UI. They are ideal for services that do not require interaction with databases, webservers, or other containers.

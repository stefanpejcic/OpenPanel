# Error: Establishing a Database Connection

If you see the error **“Error establishing a database connection”** in the OpenAdmin interface, it indicates that the database connection has failed.

---

## MySQL Service

OpenPanel and OpenAdmin use a MySQL Docker service defined in `/root/docker-compose.yml`.

First, ensure that the service is running:

**From OpenAdmin:**

1. Go to **Services > Services Status**.
2. Check the status of the MySQL service.
3. Restart it from this page if necessary.

**From the terminal:**

```bash
docker ps -a
```

Look for the `openpanel_mysql` service in the output.

---

## MySQL Fails to Start

If the MySQL service fails to start, Docker will keep restarting it. You can observe this from the `docker ps -a` output.

* If the uptime is only a few seconds and the status shows **not running** or **unhealthy**, check the service logs:

```bash
docker logs -f openpanel_mysql
```

* Copy any error messages and search online. Common issues include:

  * Switching from MySQL to MariaDB
  * MySQL updates
  * MySQL not running properly on ARM CPUs

---

## Incorrect Credentials

You can test the connection directly from the terminal using the `mysql` command.

OpenAdmin and the terminal use credentials stored in `/etc/my.cnf`:

```bash
root@demo:~# cat /etc/my.cnf 
[client]
user = panel
database = panel
password = e391ac94321d110c
host = 127.0.0.1
protocol = tcp
```

Make sure these credentials are correct and allow you to log in.

---

## Firewall

Outgoing connections on **port 3306** must be allowed. Ensure this port is open in the Sentinel Firewall (CSF).

Check `/etc/csf/csf.conf` for the `TCP_OUT=` setting and confirm that port `3306` is included. If not, add it and restart CSF:

```bash
csf -r
```

---

If none of these steps resolve the issue, contact us via the forums or open a support ticket, and we will help troubleshoot the problem.

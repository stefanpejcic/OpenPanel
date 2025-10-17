# Error establishing a database connection

The error: *Error establishing a database connection* is visible on OpenAdmin interface if databse conenction fails.


## MySQL service


OpenPanel and OpenAdmin use a MySQL docker service defined in `/root/docker-compose.yml` a

First check if servie is indeed running. 

From OpenAdmin > Services > Services Status check status for msyql service - restart it from tis page if needed.

From termianl: `docker ps -a` and check for openpanel_mysql service

## Error starting MySQL

If mysql service fails to start, docker will keep restaritng it, this sstatus is visibel from `docker ps -a`  command ouput.

if te uptime is a few seconds and status is not runing or healthy, rhen cehck servie logs:

```
docker logs -f openpanel_mysql
```

copy the issue and gogle it, most common issues are: switchin from mysql to mariadb, mysql update, mysql not runing properly on arm cpu, etc.


## Wrong Credentials

Try to establish conenction from the temrinal with jsut `mysql` comamdn.

Both termianl and OpenAdmin will use login credentials stored in `/etc/my.cnf`

```
root@demo:~# cat /etc/my.cnf 
[client]
user = panel
database = panel
password = e3971ac794321d110c
host = 127.0.0.1
protocol = tcp
```

Make sure credentials are valid, ie you can login with those credentials.

## Firewall

Outgoing connection (`TCP_OUT`) on port `3306` is required. make sure port is opened on Sentinel Firewall (CSF).

In `/etc/csf/csf.conf` check the `TCP_OUT=` and make sure it includes `3306` port, if not, add it and restart csf: `csf -r`

----

If none of the above steps work, message us on the forums or via a ticket and we'll troubleshoot it for you.

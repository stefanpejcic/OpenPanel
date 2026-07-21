# After restarting server, Nginx cannot automatically start

### Symptoms
After restarting the server, Nginx cannot automatically start

![nginx failed](https://i.postimg.cc/CL80NY5s/66a8ae7a3d7ff.png)

---

### Description 

Whhen clicking the 'Restart' or 'Start' buttons, it displays error:

![nginx failed](https://i.postimg.cc/tRwyW7qg/66a8aed75d65f.png)

`systemctl status nginx.service` displays:
```
Jul 29 16:32:12 ubuntu-s-2vcpu-4gb-120gb-intel-sgp1-01 nginx[933]: nginx: configuration file /etc/nginx/nginx.conf test failed
Jul 29 16:32:12 ubuntu-s-2vcpu-4gb-120gb-intel-sgp1-01 systemd[1]: nginx.service: Control process exited, code=exited, status=1/FAILURE
Jul 29 16:32:12 ubuntu-s-2vcpu-4gb-120gb-intel-sgp1-01 systemd[1]: nginx.service: Failed with result 'exit-code'.
Jul 29 16:32:12 ubuntu-s-2vcpu-4gb-120gb-intel-sgp1-01 systemd[1]: Failed to start A high performance web server and a reverse proxy server.
Jul 29 16:32:12 ubuntu-s-2vcpu-4gb-120gb-intel-sgp1-01 systemd[1]: nginx.service: Unit cannot be reloaded because it is inactive.
```

---

### Solution

There are errors in configuration files of one or more domains, and now after the reboot, nginx refuses to start.

Run this command to dispaly the exact error, file and even line number that causes problem:
```
nginx -t
```

You need to edit the specified lines in files, example with nano editor:

```
nano /etc/nginx/sites-enabled/DOMAIN_NAME.conf
```

then again validate nginx configuration and restart it:

```
nginx -t && service nginx restart
```


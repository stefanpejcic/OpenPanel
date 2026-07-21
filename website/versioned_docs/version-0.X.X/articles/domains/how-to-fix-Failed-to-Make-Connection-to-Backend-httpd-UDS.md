# Failed to Make Connection to Backend: httpd-UDS


When loading a domain within a browser, you may receive the following error: 

```
503 error
Service Unavailable
```

Check the error log for  the webserver:

- For Nginx:
  ```bash
  tail -f /var/log/nginx/error.log
  ```
- For Apache:
  ```bash
  tail -f /var/log/apache2/error.log
  ```

Upon checking error log, the following error is appended: 

> [Mon Jan 15 11:23:58.525837 2024] [proxy:error] [pid 63085:tid 140529576699456] (2)No such file or directory: AH02454: FCGI: attempt to connect to Unix domain socket /run/php/phpphp8.2-fpm.sock (*) failed


In the error itself we can see that the domain is using invalid path to php-fpm service: **/run/php/phpphp8.2-fpm.sock** instead of  **/run/php/php8.2-fpm.sock**

To fix this, from _OpenPanel > Domains_ click on the three dots for the domain and click on "_Edit VirtualHosts_"

![screenshot](https://i.postimg.cc/fRQqQSSb/2024-01-15-11-48.png)

then edit the _SetHandler_ part and make sure that the php path does not contain phpphp but just php

![screenshot](https://i.postimg.cc/XYcbRbJT/2024-01-15-11-50.png)

click on Save and from _Server Settings > Service Status_ restart Nginx/Apache service to apply changes.

![screenshot](https://i.postimg.cc/D08GN6PM/2024-01-15-11-52.png)

# Domain is loading the SSL of another unrelated domain

### Symptoms
Attempting to load an SSL returns a different domain website.

---

### Description 
When accessing with `https://` a domain that has no SSL installed, the Nginx webserver will automatically serve the SSL of the first domain that it finds on the server. This will result with a SSL warning for the user in browser.
![ssl-warning](https://i.postimg.cc/dV9FTBN2/prva.png)

On 'Advanced' you can see that the SSL and domain name does not match:
![ssl-exception](https://i.postimg.cc/dVFv1X6Q/druga.png)

 If ssl is accepted, it will redirect user to the domain that issued SSL.

---

### Workaround

If you have a domain name set for accessing OpenPanel, then you can set that domain ssl to be used for websites that have no SSL, and if accepted it will die:

```
nano /etc/nginx/sites-enabled/default
```

and add the following block **but replace server.stefan.rs with your domain and 11.22.33.44 with your server IP address**:
```
server {
    listen 11.22.33.44 :443 ssl http2 default_server;
    server_name _;

    ssl_certificate /etc/letsencrypt/live/server.stefan.rs/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/server.stefan.rs/privkey.pem;

    return 444;
}
```

Save and restart.
```
nginx -t && service nginx reload
```

then when user accepts the SSL it will show an error:

![444 error nginx](https://i.postimg.cc/02Yjc7HQ/treca.png)

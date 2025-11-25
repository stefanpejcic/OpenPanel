---
sidebar_position: 4
---

# Edit Domain Templates

This interface allows you to edit the VirtualHosts templates and default landing pages used for new domains.

- Default Page: Template displayed on new domains without any content.
- Suspended Website: Template displayed on websites (domains) that user suspends.
- Suspended User: Template displayed on all websites (domains) that suspended user owns.
- Apache VirtualHost: Template used for creating VirtualHosts of domains added by users with Apache webserver.
- Nginx VirtualHost: Template used for creating VirtualHosts of domains added by users with Nginx webserver.
- OpenResty VirtualHost: Template used for creating VirtualHosts of domains added by users with OpenResty webserver.

## Default Page

This is the HTML page shown on new domains that do **not** have `index.html` or `index.php` in their document root.

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="robots" content="noindex">
    <title>No website yet</title>
    <style>body {display: flex;justify-content: center;align-items: center;min-height: 100vh;margin: 0;font-family: Arial, sans-serif;background-color: #f9f9f9;color: #333;}.container {text-align: center;padding: 20px;border: 1px solid #ddd;border-radius: 8px;background-color: #fff;box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);}h1 {font-size: 24px;margin-bottom: 10px;}p {font-size: 16px;color: #666;}</style>
</head>
<body>
    <div class="container">
        <h1>Reaaady, set, internet 🎉</h1>
        <p>This domain currently has no website. Please check back later.</p>
    </div>
</body>
</html>
```

- Edit the code on the **left side** of the editor.
- Preview live changes on the **right side**.
- When finished, click **Save Files** to apply your changes.

To revert to the original template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.

## Suspended Website

This is the HTML page shown on all domains that users suspend **manually**.

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="robots" content="noindex">
    <title>Website Suspended</title>
    <style>body {display: flex;justify-content: center;align-items: center;min-height: 100vh;margin: 0;font-family: Arial, sans-serif;background-color: #ffecec;color: #333;}.container {text-align: center;padding: 20px;border: 1px solid #f5c2c2;border-radius: 8px;background-color: #fff;box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);}h1 {font-size: 24px;margin-bottom: 10px;color: #d9534f;}p {font-size: 16px;color: #666;}</style>
</head>
<body>
    <div class="container">
        <h1>Website Suspended</h1>
        <p>This website has been suspended. Please contact the administrator for more details.</p>
    </div>
</body>
</html>
```

- Edit the code on the **left side** of the editor.
- Preview live changes on the **right side**.
- When finished, click **Save Files** to apply your changes.

To revert to the original template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.



## Suspended User

This is the HTML page shown on all domains owned by the user when their account is suspended **by the Administrator**.

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="robots" content="noindex">
    <title>Account Suspended</title>
    <style>body {display: flex;justify-content: center;align-items: center;min-height: 100vh;margin: 0;font-family: Arial, sans-serif;background-color: #ffefef;color: #333;}.container {text-align: center;padding: 20px;border: 1px solid #ffcccc;border-radius: 8px;background-color: #fff;box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);}h1 {font-size: 24px;margin-bottom: 10px;color: #d9534f;}p {font-size: 16px;color: #666;}</style>
</head>
<body>
    <div class="container">
        <h1>Account Suspended</h1>
        <p>This account has been suspended. If you believe this is a mistake, please contact support.</p>
    </div>
</body>
</html>
```

- Edit the code on the **left side** of the editor.
- Preview live changes on the **right side**.
- When finished, click **Save Files** to apply your changes.

To revert to the original template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.


## Apache VirtualHost

This is the template used for creating VirtualHosts of domains added by users **when using the Apache webserver**.

```bash
<VirtualHost *:80>
    ServerName <DOMAIN_NAME>
    ServerAlias www.<DOMAIN_NAME>
    DocumentRoot <DOCUMENT_ROOT>

    # <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
    <Directory <DOCUMENT_ROOT>>
        <FilesMatch "\.(git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor)">
            Require all denied
        </FilesMatch>
    </Directory>
    # <!-- END EXPOSED RESOURCES PROTECTION -->

    <Directory <DOCUMENT_ROOT>>
        Options Indexes FollowSymLinks
        AllowOverride All
        Require all granted
    </Directory>

    DirectoryIndex index.php index.html default_page.html

    # Alias for default_page.html
    Alias /default_page.html /etc/apache2/default_page.html
    <Directory "/etc/apache2">
        <FilesMatch "^default_page.html$">
            Require all granted
            Options -Indexes
            SetEnvIf Request_URI ^/default_page.html allow_default_page
        </FilesMatch>
    </Directory>

    <FilesMatch \.php$>
        SetHandler "proxy:fcgi://php-fpm-<PHP>:9000"
    </FilesMatch>

</VirtualHost>


<VirtualHost *:443>
    ServerName <DOMAIN_NAME>
    ServerAlias www.<DOMAIN_NAME>
    DocumentRoot <DOCUMENT_ROOT>
    SSLEngine on
    SSLCertificateFile /etc/apache2/ssl/cert.crt
    SSLCertificateKeyFile /etc/apache2/ssl/cert.key

    <Directory <DOCUMENT_ROOT>>
        Options Indexes FollowSymLinks
        AllowOverride All
        Require all granted
    </Directory>

    DirectoryIndex index.php index.html default_page.html

    # Alias for default_page.html
    Alias /default_page.html /etc/apache2/default_page.html
    <Directory "/etc/apache2">
        <FilesMatch "^default_page.html$">
            Require all granted
            Options -Indexes
            SetEnvIf Request_URI ^/default_page.html allow_default_page
        </FilesMatch>
    </Directory>

    <FilesMatch \.php$>
        SetHandler "proxy:fcgi://php-fpm-<PHP>:9000"
    </FilesMatch>

</VirtualHost>
```

- Edit the code in the editor.
- When finished, click **Save Files** to apply your changes.

To revert to the original Apache template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.

## Nginx VirtualHost

This is the template used for creating VirtualHosts of domains added by users **when using the Nginx webserver**.

```bash
# content
server {
    listen 80;
    server_name <DOMAIN_NAME> www.<DOMAIN_NAME>;
    access_log off;

    # <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
     location ~* ^/(\.git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor) {
       deny all;
       return 403;
     }
    # <!-- END EXPOSED RESOURCES PROTECTION -->

    root <DOCUMENT_ROOT>;

    location / {
        real_ip_header    X-Forwarded-For;
        set_real_ip_from   172.17.0.0/16;
        try_files $uri $uri/ /index.php$is_args?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    location = /default_page.html {
        alias /etc/nginx/default_page.html;
        default_type text/html;
        access_log off;
        log_not_found off;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass php-fpm-<PHP>:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }
    
}


server {
    listen 443 ssl;
    http2 on; #nginx >= 1.25.1
    server_name <DOMAIN_NAME> www.<DOMAIN_NAME>;
    access_log off;
    
    root <DOCUMENT_ROOT>;

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/cert.crt;
    ssl_certificate_key /etc/nginx/ssl/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH";

    location / {
        real_ip_header    X-Forwarded-For;
        set_real_ip_from   172.17.0.0/16;
        try_files $uri $uri/ /index.php$is_args?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    
    location = /default_page.html {
        alias /etc/nginx/default_page.html;
        default_type text/html;
        access_log off;
        log_not_found off;
    }

    
    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass php-fpm-<PHP>:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }

}

```

- Edit the code in the editor.
- When finished, click **Save Files** to apply your changes.

To revert to the original Nginx template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.


## OpenResty VirtualHost

This is the template used for creating VirtualHosts of domains added by users **when using the OpenResty webserver**.

```bash
# content
server {
    listen 80;
    server_name <DOMAIN_NAME> www.<DOMAIN_NAME>;
    access_log off;

    # <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
     location ~* ^/(\.git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor) {
       deny all;
       return 403;
     }
    # <!-- END EXPOSED RESOURCES PROTECTION -->

    root <DOCUMENT_ROOT>;

    location /hello {
        default_type 'text/plain';
        content_by_lua_block {
            ngx.say("Hello from OpenResty!")
        }
    }

    location / {
        real_ip_header    X-Forwarded-For;
        set_real_ip_from   172.17.0.0/16;
        try_files $uri $uri/ /index.php$is_args?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    location = /default_page.html {
        alias /etc/nginx/default_page.html;
        default_type text/html;
        access_log off;
        log_not_found off;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass php-fpm-<PHP>:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }
    
}


server {
    listen 443 ssl;
    http2 on; #nginx >= 1.25.1
    server_name <DOMAIN_NAME> www.<DOMAIN_NAME>;
    access_log off;
    
    root <DOCUMENT_ROOT>;

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/cert.crt;
    ssl_certificate_key /etc/nginx/ssl/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers "EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH";

    location /hello {
        default_type 'text/plain';
        content_by_lua_block {
            ngx.say("Hello from OpenResty!")
        }
    }

    location / {
        real_ip_header    X-Forwarded-For;
        set_real_ip_from   172.17.0.0/16;
        try_files $uri $uri/ /index.php$is_args?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    location = /default_page.html {
        alias /etc/nginx/default_page.html;
        default_type text/html;
        access_log off;
        log_not_found off;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass php-fpm-<PHP>:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }

}

```

- Edit the code in the editor.
- When finished, click **Save Files** to apply your changes.

To revert to the original OpenResty template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.

## Varnish Template

This is the template used for creating `default.vcl` file for users **when using the Varnish Caching**.

The placeholder `VARNISH_BACKEND_HOST` is automatically replaced with the user's actual web server - Nginx, Apache, or OpenResty. 

```bash
vcl 4.1;

backend default {
    .host = "VARNISH_BACKEND_HOST";
    .port = "80";
    .max_connections = 2048;
}

# Define an access control list to restrict cache purging.
acl purge {
    "localhost";
    "172.17.0.0/16";
}

sub vcl_hit {
    if (req.http.X-Forwarded-Proto == "https") {
        set req.http.X-Forwarded-Proto = "https";
    } else {
        set req.http.X-Forwarded-Proto = "http";
    }

    if (obj.ttl > 0s) {
        return (deliver);
    }
    if (obj.ttl + obj.grace > 0s) {
        return (deliver);
    }

    return (miss);  # Use "miss" instead of "fetch"
}

sub vcl_recv {

    # Remove empty query string parameters
    # e.g.: www.example.com/index.html?
    if (req.url ~ "\?$") {
        set req.url = regsub(req.url, "\?$", "");
    }

    # Remove port number from host header
    set req.http.Host = regsub(req.http.Host, ":[0-9]+", "");

    # Remove the proxy header to mitigate the httpoxy vulnerability
    # See https://httpoxy.org/
    unset req.http.proxy;

    # Purge logic to remove objects from the cache.
    # Tailored to the Proxy Cache Purge WordPress plugin
    # See https://wordpress.org/plugins/varnish-http-purge/
    if(req.method == "PURGE") {
        if(!client.ip ~ purge) {
            return(synth(405,"PURGE not allowed for this IP address"));
        }
        if (req.http.X-Purge-Method == "regex") {
            ban("obj.http.x-url ~ " + req.url + " && obj.http.x-host == " + req.http.host);
            return(synth(200, "Purged"));
        }
        ban("obj.http.x-url == " + req.url + " && obj.http.x-host == " + req.http.host);
        return(synth(200, "Purged"));
    }

   # Only handle relevant HTTP request methods
    if (
        req.method != "GET" &&
        req.method != "HEAD" &&
        req.method != "PUT" &&
        req.method != "POST" &&
        req.method != "PATCH" &&
        req.method != "TRACE" &&
        req.method != "OPTIONS" &&
        req.method != "DELETE"
    ) {
        return (pipe);
    }

    # Remove tracking query string parameters used by analytics tools
    if (req.url ~ "(\?|&)(_branch_match_id|_bta_[a-z]+|_bta_c|_bta_tid|_ga|_gl|_ke|_kx|campid|cof|customid|cx|dclid|dm_i|ef_id|epik|fbclid|gad_source|gbraid|gclid|gclsrc|gdffi|gdfms|gdftrk|hsa_acc|hsa_ad|hsa_cam|hsa_grp|hsa_kw|hsa_mt|hsa_net|hsa_src|hsa_tgt|hsa_ver|ie|igshid|irclickid|matomo_campaign|matomo_cid|matomo_content|matomo_group|matomo_keyword|matomo_medium|matomo_placement|matomo_source|mc_[a-z]+|mc_cid|mc_eid|mkcid|mkevt|mkrid|mkwid|msclkid|mtm_campaign|mtm_cid|mtm_content|mtm_group|mtm_keyword|mtm_medium|mtm_placement|mtm_source|nb_klid|ndclid|origin|pcrid|piwik_campaign|piwik_keyword|piwik_kwd|pk_campaign|pk_keyword|pk_kwd|redirect_log_mongo_id|redirect_mongo_id|rtid|s_kwcid|sb_referer_host|sccid|si|siteurl|sms_click|sms_source|sms_uph|srsltid|toolid|trk_contact|trk_module|trk_msg|trk_sid|ttclid|twclid|utm_[a-z]+|utm_campaign|utm_content|utm_creative_format|utm_id|utm_marketing_tactic|utm_medium|utm_source|utm_source_platform|utm_term|vmcid|wbraid|yclid|zanpid)=") {
        set req.url = regsuball(req.url, "(_branch_match_id|_bta_[a-z]+|_bta_c|_bta_tid|_ga|_gl|_ke|_kx|campid|cof|customid|cx|dclid|dm_i|ef_id|epik|fbclid|gad_source|gbraid|gclid|gclsrc|gdffi|gdfms|gdftrk|hsa_acc|hsa_ad|hsa_cam|hsa_grp|hsa_kw|hsa_mt|hsa_net|hsa_src|hsa_tgt|hsa_ver|ie|igshid|irclickid|matomo_campaign|matomo_cid|matomo_content|matomo_group|matomo_keyword|matomo_medium|matomo_placement|matomo_source|mc_[a-z]+|mc_cid|mc_eid|mkcid|mkevt|mkrid|mkwid|msclkid|mtm_campaign|mtm_cid|mtm_content|mtm_group|mtm_keyword|mtm_medium|mtm_placement|mtm_source|nb_klid|ndclid|origin|pcrid|piwik_campaign|piwik_keyword|piwik_kwd|pk_campaign|pk_keyword|pk_kwd|redirect_log_mongo_id|redirect_mongo_id|rtid|s_kwcid|sb_referer_host|sccid|si|siteurl|sms_click|sms_source|sms_uph|srsltid|toolid|trk_contact|trk_module|trk_msg|trk_sid|ttclid|twclid|utm_[a-z]+|utm_campaign|utm_content|utm_creative_format|utm_id|utm_marketing_tactic|utm_medium|utm_source|utm_source_platform|utm_term|vmcid|wbraid|yclid|zanpid)=[-_A-z0-9+(){}%.*]+&?", "");
        set req.url = regsub(req.url, "[?|&]+$", "");
    }

    # Only cache GET and HEAD requests
    if (req.method != "GET" && req.method != "HEAD") {
        set req.http.X-Cacheable = "NO:REQUEST-METHOD";
        return(pass);
    }

    # Mark static files with the X-Static-File header, and remove any cookies
    # X-Static-File is also used in vcl_backend_response to identify static files
    if (req.url ~ "^[^?]*\.(7z|avi|bmp|bz2|css|csv|doc|docx|eot|flac|flv|gif|gz|ico|jpeg|jpg|js|less|mka|mkv|mov|mp3|mp4|mpeg|mpg|odt|ogg|ogm|opus|otf|pdf|png|ppt|pptx|rar|rtf|svg|svgz|swf|tar|tbz|tgz|ttf|txt|txz|wav|webm|webp|woff|woff2|xls|xlsx|xml|xz|zip)(\?.*)?$") {
        set req.http.X-Static-File = "true";
        unset req.http.Cookie;
        return(hash);
    }

    # No caching of special URLs, logged in users and some plugins
    if (
        req.http.Cookie ~ "wordpress_(?!test_)[a-zA-Z0-9_]+|wp-postpass|comment_author_[a-zA-Z0-9_]+|woocommerce_cart_hash|woocommerce_items_in_cart|wp_woocommerce_session_[a-zA-Z0-9]+|wordpress_logged_in_|comment_author|PHPSESSID" ||
        req.http.Authorization ||
        req.url ~ "add_to_cart" ||
        req.url ~ "edd_action" ||
        req.url ~ "nocache" ||
        req.url ~ "^/addons" ||
        req.url ~ "^/bb-admin" ||
        req.url ~ "^/bb-login.php" ||
        req.url ~ "^/bb-reset-password.php" ||
        req.url ~ "^/cart" ||
        req.url ~ "^/checkout" ||
        req.url ~ "^/control.php" ||
        req.url ~ "^/login" ||
        req.url ~ "^/logout" ||
        req.url ~ "^/lost-password" ||
        req.url ~ "^/my-account" ||
        req.url ~ "^/product" ||
        req.url ~ "^/register" ||
        req.url ~ "^/register.php" ||
        req.url ~ "^/server-status" ||
        req.url ~ "^/signin" ||
        req.url ~ "^/signup" ||
        req.url ~ "^/stats" ||
        req.url ~ "^/wc-api" ||
        req.url ~ "^/wp-admin" ||
        req.url ~ "^/wp-comments-post.php" ||
        req.url ~ "^/wp-cron.php" ||
        req.url ~ "^/wp-login.php" ||
        req.url ~ "^/wp-activate.php" ||
        req.url ~ "^/wp-mail.php" ||
        req.url ~ "^/wp-login.php" ||
        req.url ~ "^\?add-to-cart=" ||
        req.url ~ "^\?wc-api=" ||
        req.url ~ "^/preview=" ||
        req.url ~ "^/\.well-known/acme-challenge/"
    ) {
	     set req.http.X-Cacheable = "NO:Logged in/Got Sessions";
	     if(req.http.X-Requested-With == "XMLHttpRequest") {
		     set req.http.X-Cacheable = "NO:Ajax";
	     }
        return(pass);
    }

    # Remove any cookies left
    unset req.http.Cookie;
    return(hash);
}


sub vcl_hash {
    if(req.http.X-Forwarded-Proto) {
        # Create cache variations depending on the request protocol
        hash_data(req.http.X-Forwarded-Proto);
    }
}

sub vcl_pipe {
  if (req.http.upgrade) {
    set bereq.http.upgrade = req.http.upgrade;
  }
  return (pipe);
}

sub vcl_backend_response {
    # Inject URL & Host header into the object for asynchronous banning purposes
    set beresp.http.x-url = bereq.url;
    set beresp.http.x-host = bereq.http.host;

    # If we dont get a Cache-Control header from the backend
    # we default to 1h cache for all objects
    if (!beresp.http.Cache-Control) {
        set beresp.ttl = 1h;
        set beresp.http.X-Cacheable = "YES:Forced";
    }

    # If the file is marked as static we cache it for 1 day
    if (bereq.http.X-Static-File == "true") {
        unset beresp.http.Set-Cookie;
        set beresp.http.X-Cacheable = "YES:Forced";
        set beresp.ttl = 1d;
    }

	# Remove the Set-Cookie header when a specific Wordfence cookie is set
    if (beresp.http.Set-Cookie ~ "wfvt_|wordfence_verifiedHuman") {
	    unset beresp.http.Set-Cookie;
	 }

    if (beresp.http.Set-Cookie) {
        set beresp.http.X-Cacheable = "NO:Got Cookies";
    } elseif(beresp.http.Cache-Control ~ "private") {
        set beresp.http.X-Cacheable = "NO:Cache-Control=private";
    }
}


sub vcl_deliver {
    # Debug header
    if(req.http.X-Cacheable) {
        set resp.http.X-Cacheable = req.http.X-Cacheable;
    } elseif(obj.uncacheable) {
        if(!resp.http.X-Cacheable) {
            set resp.http.X-Cacheable = "NO:UNCACHEABLE";
        }
    } elseif(!resp.http.X-Cacheable) {
        set resp.http.X-Cacheable = "YES";
    }

### uncomment to remove from responses ###
#  unset resp.http.X-Powered-By;
#  unset resp.http.Server;
#  unset resp.http.server;
#  unset resp.http.via;
#  unset resp.http.x-powered-by;
#  unset resp.http.x-runtime;
#  unset resp.http.x-varnish;
    unset resp.http.x-url;
    unset resp.http.x-host;

  return (deliver);
}
```

- Edit the code in the editor.
- When finished, click **Save Files** to apply your changes.

To revert to the original Varnish template:

1. Click **Restore Default** to reload the version from GitHub.
2. Don’t forget to click **Save Files** afterward to confirm the reset.





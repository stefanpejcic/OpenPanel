# Lakk használata a WordPress-szel

A Varnish nem támogatja az SSL-t, ezért le kell állítani a TLS-t, mielőtt kéréseket továbbítana neki, majd miután megkapta a http-t a Varnish-től, használja újra a https-t.

WordPress esetén elegendő a következő kód hozzáadása a wp-config.php fájlhoz:

```
define('FORCE_SSL_ADMIN', true); if ($_SERVER['HTTP_X_FORWARDED_PROTO'] == 'https') $_SERVER['HTTPS']='on';
```


> **SZERKESZTÉS**: Az 1.5.7-es verziótól ez automatikusan hozzáadódik az új WP-telepítésekhez.

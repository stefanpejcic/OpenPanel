# Using Varnish with WordPress

Varnish does not support SSL, so you need to terminate TLS before passing requests to it, then after receiving http from Varnish, use again https.

For WordPress, adding the following code in wp-config.php is enough:

```
define('FORCE_SSL_ADMIN', true); if ($_SERVER['HTTP_X_FORWARDED_PROTO'] == 'https') $_SERVER['HTTPS']='on';
```

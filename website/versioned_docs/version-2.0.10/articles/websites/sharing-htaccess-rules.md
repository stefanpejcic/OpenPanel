# Sharing .htaccess rules

Shared .htaccess Across Multiple Domains/Subdomains

---

An .htaccess file may be shared across multiple domains or subdomains by being located in a common parent directory. Locating an .htaccess under `/var/www/html` will allow any domain or subdomain located under `/var/www/html/` to inherit these rules.

Example Directory Structure:
```
example.com        -> /var/www/html/example.com/
blog.example.com   -> /var/www/html/example.com/blog/
another-site.com   -> /var/www/html/another-site.com/
```

To apply a common `.htaccess` to all three, place the file in:

```
/var/www/html/.htaccess
```

This way, all domains and subdomains under that path will follow the shared rules automatically.

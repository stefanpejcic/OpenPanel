# A .htaccess szabályok megosztása

Megosztott .htaccess több tartományon/aldomainen keresztül

---

Egy .htaccess fájl több tartomány vagy aldomain között is megosztható, ha egy közös szülőkönyvtárban található. Ha a .htaccess fájlt a „/var/www/html” alatt találja meg, akkor a „/var/www/html/” alatt található bármely domain vagy aldomain örökölheti ezeket a szabályokat.

Példa könyvtárszerkezetre:
```
example.com        -> /var/www/html/example.com/
blog.example.com   -> /var/www/html/example.com/blog/
another-site.com   -> /var/www/html/another-site.com/
```

Ha közös `.htaccess-t szeretne alkalmazni mindháromra, helyezze a fájlt a következő helyre:

```
/var/www/html/.htaccess
```

Így az adott útvonalon lévő összes domain és aldomain automatikusan követi a megosztott szabályokat.

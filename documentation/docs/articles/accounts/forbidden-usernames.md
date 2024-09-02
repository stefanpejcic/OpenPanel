# Reserved and Invalid Usernames

## Basic restrictions

OpenPanel applies the following rules when you create or modify a OpenPanel username:

Usernames may **only** use lowercase letters (a–z) and digits (0–9).
- Usernames cannot contain less than 3 characters.
- Usernames cannot contain more than  than 16 characters.

##  Reserved usernames

OpenPanel reserves some usernames for the system’s use, and you cannot use them for OpenPanel accounts. This list of reserved usernames can grow over time, and new versions of OpenPanel may add to this list.

OpenPanel checks the following file to determine whether to reserve or restrict a username:

```bash
/etc/openpanel/openadmin/config/forbidden_usernames.txt
```

Currently reserved usernames:

- test
- restart
- reboot
- shutdown
- exec
- root
- admin
- ftp
- lsws
- litespeed
- 1000
- vsftpd
- httpd
- apache2
- apache
- docker
- podman
- nginx
- php
- mysql
- mysqld
- www-data
- openpanel
- openadmin

To reserver a certail username, simply add it to the forbidden_usernames.txt file.

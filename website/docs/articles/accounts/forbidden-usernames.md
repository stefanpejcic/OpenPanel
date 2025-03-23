# Reserved and Invalid Usernames

## Basic restrictions


### OpenPanel

OpenPanel applies the following rules when you create or modify a OpenPanel username:

Usernames may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Usernames cannot contain less than **3** characters.
- Usernames cannot contain more than **16** characters.

### OpenAdmin
OpenAdmin applies the following rules when you create or modify an admin username:

Usernames may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Usernames cannot contain less than **5** characters.
- Usernames cannot contain more than **30** characters.

Passwords may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Passwords cannot contain less than **5** characters.
- Passwords cannot contain more than **30** characters.

### SSH
OpenAdmin applies the following rules when you modify an SSH password:

Passwords may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Passwords cannot contain less than **8** characters.
- Passwords cannot contain more than **20** characters.

### MySQL / MariaDB
OpenPanel applies the following rules when you create databases and users:

Database names may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Database name cannot contain less than **1** character.
- Database name cannot contain more than **64** characters.


Database users may **only** use lowercase letters (`a–z`) and digits (`0–9`).
- Usernames cannot contain less than **1** characters.
- Usernames cannot contain more than **32** characters.


Database User's Passwords may **only** use lowercase letters (`a–z`), underscores (`_`) and digits (`0–9`).
- Passwords cannot contain less than **8** characters.
- Passwords cannot contain more than **32** characters.


### FTP
OpenPanel applies the following rules when you create FTP sub-accounts:

FTP sub-users **must end** with dot (`.`) followed by the OpenPanel username - example: `ftpuser.openpaneluser`.
- Usernames cannot contain less than **3** characters.
- Usernames cannot contain more than **32** characters.


FTP User's Passwords must contain **at least one** uppercase letters (`A-Z`), lowercase letters (`a–z`), digits (`0–9`) and special symbols (`!`, `.`, `,`, `@`, `#`, `_`, `-`).
- Passwords cannot contain less than **8** characters.
- Passwords cannot contain more than **32** characters.

### Emails

OpenPanel applies the following rules when you create email accounts:

Email accounts may **only** use lowercase letters (`a–z`), dashes (`-`), underscores (`_`) and digits (`0–9`).
Email accounts **must contain** `@` symbol followed by the domain name - example: `account@example.net`.
- Usernames cannot contain less than **1** character.
- Usernames cannot contain more than **32** characters.

Email accounts passwords can contain **only** uppercase letters (`A-Z`), lowercase letters (`a–z`), underscores (`_`) and digits (`0–9`).
- Passwords cannot contain less than **8** characters.
- Passwords cannot contain more than **32** characters.


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

To reserve a certain username, simply add it to the forbidden_usernames.txt file.

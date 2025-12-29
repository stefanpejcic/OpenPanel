# How to install a PHP extension in OpenPanel

OpenPanel uses the [`shinsenter/php`](https://github.com/shinsenter/php) Docker images for PHP services. These images include many commonly used PHP extensions out of the box and provide a simple way to install additional extensions when needed.

## Pre-installed PHP Extensions

Most popular PHP extensions are pre-installed by default, allowing projects to start quickly without extra configuration:

```
apcu
bcmath
calendar
Core
ctype
curl
date
dom
exif
fileinfo
filter
gd
gettext
gmp
hash
iconv
igbinary
intl
json
lexbor
libxml
mbstring
msgpack
mysqli
mysqlnd
openssl
pcntl
pcre
PDO
pdo_mysql
pdo_pgsql
pdo_sqlite
pgsql
Phar
posix
random
readline
redis
Reflection
session
SimpleXML
sodium
SPL
sqlite3
standard
tidy
tokenizer
uri
uuid
xml
xmlreader
xmlwriter
yaml
Zend OPcache
zip
zlib
```

and OpenPanel adds the `Zend OPcache` extension.

----

## Adding Custom PHP Extensions

The PHP images include a helper command called `phpaddmod` to install PHP extensions: it installs the extension and edits the `php.ini` file to enable it.


### Install a PHP Extension for a Single User

To install a PHP extension for a single user, open a terminal for the desired PHP version.

### Example assumptions

* **Username**: `stefan`
* **PHP version**: `8.5`
* **Service name**: `php-fpm-8.5`

Adjust these values to match your environment.

---

### Option 1: OpenPanel Terminal (Enterprise Edition)

If the **Docker** module is enabled:

1. Go to **Containers → Terminal** in OpenPanel
2. Select the desired PHP version
3. Run the following commands

Example: installing the **GMP** extension:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### Option 2: OpenAdmin Terminal

1. Go to **Users → (select user) → Services**
2. Open the terminal for the desired PHP version
3. Run the following commands

Example: installing the **GMP** extension:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### Option 3: SSH Access (root)

1. Connect to the server via SSH as `root`
2. Enter the PHP container for the desired user and version:

```
docker --context=USERNAME_HERE exec -it php-fpm-8.5 bash
```

3. Install the extension

Example: installing the **GMP** extension:

```
apt update
apt install -y libgmp-dev
phpaddmod gmp
```

---

### Verify Installation

You can verify that the extension is installed by running:

```
php -m
```

Or by checking your application or website.

---

## Install PHP Extensions for All New Users

To have a PHP extension automatically available for **all newly created users**, you have two options:

1. **Modify the startup command** in the `docker-compose.yml` template to install the extension on container startup

   * This increases PHP container startup time

2. **Use a custom PHP Docker image** with the required extensions pre-installed

   * Recommended for production environments

> These approaches are intended for advanced users. If you need help, please contact the OpenPanel community forum or Discord channel.

---

## Install PHP Extensions for All Existing Users

To install a PHP extension for **all existing users**, you can:

* Run a custom Bash script that installs the extension for each user and PHP version
* Switch all users to a custom PHP Docker image that already includes the extension

Both approaches require advanced knowledge of Docker and OpenPanel internals.

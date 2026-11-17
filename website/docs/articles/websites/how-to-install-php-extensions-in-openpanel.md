---
sidebar_label: "How to install a PHP extension in OpenPanel"
---

# How to Install a PHP Extension

OpenPanel uses the [`openpanel/php-fpm`](https://github.com/stefanpejcic/php-fpm) Docker images for PHP services, one image per PHP version from 5.6 to 8.5. These images include many commonly used PHP extensions out of the box and provide a simple way to install additional extensions when needed.

Users can also install, enable and disable extensions themselves from **OpenPanel > PHP > Extensions**.

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
ftp
gd
gettext
hash
iconv
igbinary
intl
json
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
uuid
xml
xmlreader
xmlwriter
yaml
Zend OPcache
zip
zlib
```

The exact list depends on the PHP version (for example `random`, `lexbor` and `uri` exist only in newer PHP versions), and the [ionCube Loader](https://www.ioncube.com/loaders.php) is added by OpenPanel for every version that has one.


----
## Enabling installed PHP Extensions

1. **Connect to the server** via SSH as `root`
2. **Enter the PHP container** for the desired user and version:
```
   opencli docker USERNAME_HERE php-fpm-8.5
```
 
3. **Enable the extension** using `phpaddmod` (re-enables an installed but disabled extension, or installs it if it's missing):
```
   phpaddmod mysqli pdo_mysql
```
   To disable one, use `phpdelmod`:
```
   phpdelmod mysqli
```

4. **Exit and restart** the PHP container:
```
   exit
   source /usr/local/opencli/lib/podman.sh
   podman_user USERNAME_HERE restart php-fpm-8.5
```

## Adding Custom PHP Extensions

The PHP images include a helper command called `phpaddmod` to install PHP extensions: it installs any system libraries the extension needs, compiles the extension if needed, and enables it. It uses [install-php-extensions](https://github.com/mlocati/docker-php-extension-installer), so every extension listed there can be installed.


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
phpaddmod gmp
```

---

### Option 2: OpenAdmin Terminal

1. Go to **Users → (select user) → Services**
2. Open the terminal for the desired PHP version
3. Run the following commands

Example: installing the **GMP** extension:

```
phpaddmod gmp
```

---

### Option 3: SSH Access (root)

1. Connect to the server via SSH as `root`
2. Enter the PHP container for the desired user and version:

```
opencli docker USERNAME_HERE php-fpm-8.5
```

3. Install the extension

Example: installing the **GMP** extension:

```
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

1. **Use a custom PHP Docker image** built from `openpanel/php-fpm` with the required extensions pre-installed, and set it as the `image:` of the `php-fpm-X.Y` services in the `docker-compose.yml` template:

   ```dockerfile
   FROM openpanel/php-fpm:8.5
   RUN phpaddmod gmp imagick
   ```

   * Recommended for production environments

2. **Modify the startup command** in the `docker-compose.yml` template to install the extension on container startup

   * This increases PHP container startup time by the time needed to compile the extension

> These approaches are intended for advanced users. If you need help, please contact the OpenPanel community forum or Discord channel.

---

## Install PHP Extensions for All Existing Users

To install a PHP extension for **all existing users**, you can:

* Run a custom Bash script that installs the extension for each user and PHP version
* Switch all users to a custom PHP Docker image that already includes the extension

Both approaches require advanced knowledge of Docker and OpenPanel internals.

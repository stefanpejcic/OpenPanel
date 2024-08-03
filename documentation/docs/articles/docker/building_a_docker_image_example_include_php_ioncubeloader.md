# Building a Custom Docker image: pre-install IonCube Loader

Docker images are a base for every OpenPanel plan, and they define what Technology Stack is pre-installed for the user.

If you need something pre-installed or configured for all users on a plan, like the ionCube loader extension for the installed PHP version, follow these recommended steps:

1. Download example Docker image files.
2. Edit to include the new configuration/service.
3. Create the Docker image.
4. Create a new plan to use the image.
5. Test: create a new user on that plan.

### Download Docker Examples

Currently, we provide and maintain 2 Docker images, one that uses Nginx and the other with Apache. Download the image files that you want, either nginx or apache:

For Nginx:
```
git clone -n --depth=1 --filter=tree:0 \
https://github.com/stefanpejcic/OpenPanel/tree/main/docker/nginx
cd OpenPanel
git sparse-checkout set --no-cone docker/nginx
git checkout
```

For Apache:
```
git clone -n --depth=1 --filter=tree:0 \
https://github.com/stefanpejcic/OpenPanel/tree/main/docker/apache
cd OpenPanel
git sparse-checkout set --no-cone docker/apache
git checkout
```


After downloading the source files, there should be:
- **Dockerfile**: This file defines the software stack of the image and is used only to build the image.
- **entrypoint.sh**: Used by OpenPanel to monitor users' services status and versions, such as tracking the random IP of users' Docker containers, status of Redis/Memcached services, PHP versions installed, etc. This file is essential and should not be modified. It is executed on every restart of the user container.
- **my.cnf**: A MySQL service configuration file for customizing MySQL limits and settings.
- **pma.php**: A configuration file for phpMyAdmin to enable passwordless login from OpenPanel to phpMyAdmin.

### Edit Dockerfile

In this example, we will be adding the ionCube loader PHP extension to the image so that the pre-installed PHP 8.3 version already has ionCube enabled.

Open the **Dockerfile** with a text editor:

``` 
nano Dockerfile
```

Inside, include steps to download the ionCube extension and include it in the php.ini file for both CLI and FPM service.

First, add the part to download the extension. Ensure this is added after PHP install: 
```
#download for PHP 8.3
RUN mkdir -p /usr/local/lib && curl -sSlL  -o /tmp/ioncube.tar.gz https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz && tar -x --strip-components=1 -C /usr/local/lib -f /tmp/ioncube.tar.gz ioncube/ioncube_loader_lin_8.3.so
```

This downloads the extension from the ionCube website and adds it to the folder.

Now, add the part to include it in the php.fpm files:

```bash
# add for phpfpm service
RUN (echo 'zend_extension = /usr/local/lib/ioncube_loader_lin_8.3.so' && cat /etc/php/8.3/fpm/php.ini) > /etc/php/8.3/fpm/php.ini.new && mv /etc/php/8.3/cli/php.ini.new /etc/php/8.3/fpm/php.ini

# add for cli
RUN (echo 'zend_extension = /usr/local/lib/ioncube_loader_lin_8.3.so' && cat /etc/php/8.3/cli/php.ini) > /etc/php/8.3/cli/php.ini.new && mv /etc/php/8.3/cli/php.ini.new /etc/php/8.3/cli/php.ini
```


That Is it. Save the file and exit.

### Create Docker image

After adding the commands to download and install the ionCube loader, we now need to build a Docker image from the files.

Depending on the downloaded example, replace `nginx` or `apache` in this command:

```
docker build . -t my_custom_nginx_image
```

Make sure that the name contains either `nginx` or `apache`, as this is used by OpenPanel to determine the features to display to the user on their interface.

If the build fails, there will be an error message indicating which exact line caused the error. That line needs to be edited and the build repeated.

After successfully building the image, the next step is to create a new OpenPanel hosting plan to use it.

### Create Hosting Plan

Currently, new hosting plans with custom images can only be added from the terminal.
To create a new plan, run the following command:

```
opencli plan-create <NAME> <DESCRIPTION> <DOMAINS_LIMIT> <WEBSITES_LIMIT> <DISK_LIMIT> <INODES_LIMITS> <DATABASES_LIMIT> <CPU_LIMIT> <RAM_LIMIT> <DOCKER_IMAGE> <PORT_SPEED_LIMIT>
```

Example:

```
opencli plan-create cloud_8 "Custom plan with 8GB of RAM&CPU" 0 0 15 500000 0 8 8 my_custom_nxinx_image 200
```

### Create new user
After creating a plan that uses our newly built Docker image, the final step is to create a new user on the plan to test that the ionCube loader is indeed enabled.

Create a new user from the OpenAdmin interface or via terminal:
```
opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_NAME>
```

Example:
```
opencli user-add stefan strongpass123 stefan@pejcic.rs cloud_8
```

After creating the user, log in to their OpenPanel and confirm that the extension is enabled by opening the PHP.INI editor or by creating a php.info file and searching for the ionCube loader.

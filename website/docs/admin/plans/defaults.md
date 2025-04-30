---
sidebar_position: 3
---

# Defaults

From **OpenAdmin > Hosting Plans > Defaults** Administrators can edit `docker-compose.yml` and `.env` files used for new users.

In these files you can configure additional services (docker containers) and change defaults for existing services.

Keep in mind that this is intended for advanced users and misconfiguration can cause exposed system ports, user hogging resources or exciding disk limits.

When adding new services keep in mind the following:

- container name must be same as service name
- cpu and mempory limits for service must be named in format: `SERVICE_`CPU and `SERVICE_`RAM.
- other variables for service should also be prefixed with `SERVICE_`
- processes inside containers must be run as root (`0`) user in order for container files to be counted against user quota and avoid permission issues.
- configuration files should be mounted in read-only mode.


## docker-compose.yml

In `docker-compose.yml` you can view and configure services for new users.

### cron

[Ofelia](https://github.com/mcuadros/ofelia) is used as a job scheduler from OpenPanel 1.2.0

- `crons.ini` file needs to be mounted from user's home directory
- and docker socket must be mounted to execute commands in all containers.
- cpu and ram limits are optional but recommended in multi-user setup.



### mysql

### mariadb

### phpmyadmin


### backup

### memcached

### redis

### opensearch

### elasticserach

### php-fpm



### varnish

### nginx

### apache

### openresty

`openresty` is a webserver based on Nginx core but with support Lua language.

Available variables are:

- `OPENRESTY_VERSION` - version (tag) from [Docker Hub](https://hub.docker.com/r/openresty/openresty/tags)
- `PROXY_HTTP_PORT` - if defined means Varnish is in use, so use `HTTP_PORT` instead.
- `HTTPS_PORT` - for https to https proxy.
- `OPENRESTY_CPU` - cpu limit for service
- `OPENRESTY_RAM` - memory limit for service

It also has a few hard-coded values:
- `pids: 100` - because [docker has a bug](https://github.com/docker/cli/issues/5009) that does not allow setting valiable for pids number.


### examples

Other services are provied as examples and they don't have pages in OpenPanel UI, yet are fully managable from **OpenPanel > Contianers** page.

- filebrowser
- minecraft
- mssql
- postgres
- pgadmin



### deprecated

- `user_service` (DEPRECATED) remains for backwards compatibility only and should not be used or modified.
- `busybox` (DEPRECATED).



---------


## .env

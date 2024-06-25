FROM ubuntu:latest

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"

ENV TZ=UTC
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y software-properties-common && \
    add-apt-repository ppa:ondrej/php && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
        msmtp \
        ttyd \
        screen \
        nginx \
        mysql-server \
        php8.2-fpm \
        php8.2-mysql \
        php8.2-curl \
        php8.2-gd \
        php8.2-mbstring \
        php8.2-xml \
        php8.2-xmlrpc \
        php8.2-soap \
        php8.2-intl \
        php8.2-zip \
        php8.2-bcmath \
        php8.2-calendar \
        php8.2-exif \
        php8.2-ftp \
        php8.2-ldap \
        php8.2-sockets \
        php8.2-sysvmsg \
        php8.2-sysvsem \
        php8.2-sysvshm \
        php8.2-tidy \
        php8.2-uuid \
        php8.2-opcache \
        php8.2-redis \
        curl \
        cron \
        pwgen \
        zip \
        unzip \
        wget \
        nano \
        phpmyadmin \
        openssh-server \
        php-mbstring && \
        apt-get clean && \
        apt-get autoremove -y && \
        rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

########## PHP Composer ##########
COPY --from=composer:latest /usr/bin/composer /usr/local/bin/composer

########## MySQL ##########
COPY mysql/mysqld.cnf /etc/mysql/mysql.conf.d/
RUN chown mysql:mysql -R /var/lib/mysql
RUN usermod -d /var/lib/mysql mysql
RUN chmod a+rwx /run/mysqld

# Get MySQL root password from debian.cnf
RUN MYSQL_PASSWORD=$(awk -F "=" '/password/ {gsub(/[ \t]+/, "", $2); print $2; exit}' /etc/mysql/debian.cnf); \
    \
    if [ -z "$MYSQL_PASSWORD" ]; then \
        echo "Error: Password not found in /etc/mysql/debian.cnf"; \
        exit 1; \
    fi; \
    \
    service mysql start && \
    mysql -u root -e "ALTER USER 'debian-sys-maint'@'localhost' IDENTIFIED BY '$MYSQL_PASSWORD';"


########## EXPOSED PORTS ##########
EXPOSE 22 80 3306 7681 8080


########## NGINX ##########
RUN sed -i \
    -e 's/# server_names_hash_bucket_size 640;/server_names_hash_bucket_size 640;/' \
    -e 's/# worker_connections 768;/worker_connections 10000;/' \
    /etc/nginx/nginx.conf && \
    chmod 700 /etc/nginx/sites-available && \
    chmod 700 /etc/nginx/sites-enabled

COPY nginx/default.conf /etc/nginx/sites-available/default



########## PHPMYADMIN ##########

COPY phpmyadmin/config.inc.php /etc/phpmyadmin/
COPY phpmyadmin/pma.php /usr/share/phpmyadmin/pma.php

RUN new_password=$(openssl rand -base64 12 | tr -d '/+' | head -c 16) \
    && sed -i "s/\(\$dbpass='.*'\)/\$dbpass='$new_password';/" "/etc/phpmyadmin/config-db.php" \
    && pma_file="/usr/share/phpmyadmin/pma.php" \
    && sed -i "s/\(\$_SESSION\['PMA_single_signon_user'] = '\).*\(';.*\)/\1phpmyadmin\2/" "$pma_file" \
    && sed -i "s/\(\$_SESSION\['PMA_single_signon_password'] = '\).*\(';.*\)/\1$new_password\2/" "$pma_file" \
    && sed -i "s/\(\$_SESSION\['PMA_single_signon_host'] = '\).*\(';.*\)/\1localhost\2/" "$pma_file" \
    && service mysql start \
    && mysql -u root -e "CREATE DATABASE IF NOT EXISTS phpmyadmin;" \
    && mysql -u root -e "CREATE USER 'phpmyadmin'@'localhost' IDENTIFIED BY '$new_password';" \
    && mysql -u root -e "GRANT ALL PRIVILEGES ON phpmyadmin.* TO 'phpmyadmin'@'localhost';" \
    && mysql -u root -e "GRANT ALL ON *.* TO 'phpmyadmin'@'localhost';" \
    && mysql -u root -e "REVOKE CREATE USER ON *.* FROM 'phpmyadmin'@'localhost';" \
    && mysql -u root -e "REVOKE CREATE ON *.* FROM 'phpmyadmin'@'localhost';" \
    && mysql -u root -e "FLUSH PRIVILEGES;" \
    && mysql -u root < /usr/share/doc/phpmyadmin/examples/create_tables.sql \
    && service mysql stop

########## PHP-FPM ##########
# 8.2
RUN update-alternatives --set php /usr/bin/php8.2
RUN sed -i \
    -e 's/^upload_max_filesize = .*/upload_max_filesize = 1024M/' \
    -e 's/^max_input_time = .*/max_input_time = 600/' \
    -e 's/^memory_limit = .*/memory_limit = -1/' \
    -e 's/^post_max_size = .*/post_max_size = 1024M/' \
    -e 's/^max_execution_time = .*/max_execution_time = 600/' \
    -e 's/^opcache.enable= .*/opcache.enable=1/' \
    /etc/php/8.2/fpm/php.ini
RUN sed -i 's|;sendmail_path = *|sendmail_path = "/usr/bin/msmtp -t"|g' /etc/php/8.2/fpm/php.ini


RUN sed -i \
    -e 's/^upload_max_filesize = .*/upload_max_filesize = 1024M/' \
    -e 's/^max_input_time = .*/max_input_time = 600/' \
    -e 's/^memory_limit = .*/memory_limit = -1/' \
    -e 's/^post_max_size = .*/post_max_size = 1024M/' \
    -e 's/^max_execution_time = .*/max_execution_time = 600/' \
    -e 's/^opcache.enable= .*/opcache.enable=1/' \
    /etc/php/8.2/cli/php.ini
RUN sed -i 's|;sendmail_path = *|sendmail_path = "/usr/bin/msmtp -t"|g' /etc/php/8.2/cli/php.ini


########## EMAIL ##########
COPY email/msmtprc /etc/msmtprc



########## SSH ##########
ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile


########## SSL #############
RUN mkdir -p /etc/nginx/ssl/ && cd /etc/nginx/ssl/ && openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -subj "/C=GB/ST=London/L=London/O=Global Security/OU=R&D Department/CN=openpanel.co"  -keyout cert.key  -out cert.crt

########## TERMINAL #############
# fix for webterminal: bash: permission denied: /home/user/.bashrc
RUN chmod 755 /root


########## WP-CLI ##########
RUN curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

########## cleanup ##########
RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*



########## docker run entrypoint  ##########
COPY entrypoint.sh /etc/entrypoint.sh
RUN chmod +x /etc/entrypoint.sh
CMD ["/bin/sh", "-c", "/etc/entrypoint.sh ; tail -f /dev/null"]

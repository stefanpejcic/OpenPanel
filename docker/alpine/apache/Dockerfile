FROM alpine:latest

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"

ENV TZ=UTC

# Install required packages
RUN apk update && \
    apk add --no-cache \
        bash \
        msmtp \
        ttyd \
        screen \
        apache2 \
        mariadb \
        php82 \
        php82-fpm \
        php82-mysqli \
        php82-curl \
        php82-gd \
        php82-mbstring \
        php82-xml \
        php82-soap \
        php82-intl \
        php82-zip \
        php82-bcmath \
        php82-calendar \
        php82-exif \
        php82-ftp \
        php82-ldap \
        php82-sockets \
        php82-sysvmsg \
        php82-sysvsem \
        php82-sysvshm \
        php82-tidy \
        php82-uuid \
        php82-opcache \
        php82-redis \
        curl \
        cron \
        pwgen \
        zip \
        unzip \
        wget \
        nano \
        phpmyadmin \
        openssh-server \
        php82-mbstring \
        openrc \
        tzdata && \
    rm -rf /var/cache/apk/*

# Create alias for service
RUN echo "alias service='rc-service'" >> /etc/profile

########## PHP Composer ##########
COPY --from=composer:latest /usr/bin/composer /usr/local/bin/composer

########## MySQL ##########
COPY mysql/mysqld.cnf /etc/mysql/my.cnf
RUN chown mysql:mysql -R /var/lib/mysql

# Configure MariaDB (MySQL) root user and start the service
RUN rc-update add mariadb default && \
    /etc/init.d/mariadb setup && \
    /etc/init.d/mariadb start && \
    MYSQL_PASSWORD=$(awk -F "=" '/password/ {gsub(/[ \t]+/, "", $2); print $2; exit}' /etc/mysql/debian.cnf) && \
    mysql -u root -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '$MYSQL_PASSWORD';"

########## EXPOSED PORTS ##########
EXPOSE 22 3306 7681 8080

########## APACHE ##########
RUN rc-update add apache2 default
COPY apache/apache2.conf /etc/apache2/

# Enable necessary Apache modules
RUN sed -i 's/#LoadModule rewrite_module/LoadModule rewrite_module/' /etc/apache2/httpd.conf && \
    sed -i 's/#LoadModule proxy_module/LoadModule proxy_module/' /etc/apache2/httpd.conf && \
    sed -i 's/#LoadModule proxy_http_module/LoadModule proxy_http_module/' /etc/apache2/httpd.conf && \
    sed -i 's/#LoadModule remoteip_module/LoadModule remoteip_module/' /etc/apache2/httpd.conf && \
    sed -i 's/#LoadModule headers_module/LoadModule headers_module/' /etc/apache2/httpd.conf

RUN mkdir -p /var/log/apache2/domlogs/

########## PHPMYADMIN ##########
COPY phpmyadmin/config.inc.php /etc/phpmyadmin/
COPY phpmyadmin/pma.php /usr/share/phpmyadmin/pma.php

RUN new_password=$(openssl rand -base64 12 | tr -d '/+' | head -c 16) && \
    sed -i "s/\(\$dbpass='.*'\)/\$dbpass='$new_password';/" "/etc/phpmyadmin/config-db.php" && \
    sed -i "s/\(\$_SESSION\['PMA_single_signon_user'] = '\).*\(';.*\)/\1phpmyadmin\2/" /usr/share/phpmyadmin/pma.php && \
    sed -i "s/\(\$_SESSION\['PMA_single_signon_password'] = '\).*\(';.*\)/\1$new_password\2/" /usr/share/phpmyadmin/pma.php && \
    sed -i "s/\(\$_SESSION\['PMA_single_signon_host'] = '\).*\(';.*\)/\1localhost\2/" /usr/share/phpmyadmin/pma.php && \
    /etc/init.d/mariadb start && \
    mysql -u root -e "CREATE DATABASE IF NOT EXISTS phpmyadmin;" && \
    mysql -u root -e "CREATE USER 'phpmyadmin'@'localhost' IDENTIFIED BY '$new_password';" && \
    mysql -u root -e "GRANT ALL PRIVILEGES ON phpmyadmin.* TO 'phpmyadmin'@'localhost';" && \
    mysql -u root -e "GRANT ALL ON *.* TO 'phpmyadmin'@'localhost';" && \
    mysql -u root -e "REVOKE CREATE USER ON *.* FROM 'phpmyadmin'@'localhost';" && \
    mysql -u root -e "REVOKE CREATE ON *.* FROM 'phpmyadmin'@'localhost';" && \
    mysql -u root -e "FLUSH PRIVILEGES;" && \
    mysql -u root < /usr/share/doc/phpmyadmin/examples/create_tables.sql && \
    /etc/init.d/mariadb stop

########## PHP-FPM ##########
RUN update-alternatives --set php /usr/bin/php8.2 && \
    sed -i \
    -e 's/^upload_max_filesize = .*/upload_max_filesize = 1024M/' \
    -e 's/^max_input_time = .*/max_input_time = 600/' \
    -e 's/^memory_limit = .*/memory_limit = -1/' \
    -e 's/^post_max_size = .*/post_max_size = 1024M/' \
    -e 's/^max_execution_time = .*/max_execution_time = 600/' \
    -e 's/^opcache.enable= .*/opcache.enable=1/' \
    -e 's|^;sendmail_path = .*|sendmail_path = "/usr/bin/msmtp -t"|' \
    /etc/php82/php.ini && \
    sed -i 's|;sendmail_path = *|sendmail_path = "/usr/bin/msmtp -t"|g' /etc/php82/php.ini

########## EMAIL ##########
COPY email/msmtprc /etc/msmtprc

########## SSH ##########
ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

########## SSL #############
RUN mkdir -p /etc/apache2/ssl/ && \
    cd /etc/apache2/ssl/ && \
    openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
        -subj "/C=GB/ST=London/L=London/O=Global Security/OU=R&D Department/CN=openpanel.co" \
        -keyout cert.key -out cert.crt

########## TERMINAL #############
# Fix for webterminal: bash: permission denied: /home/user/.bashrc
RUN chmod 755 /root

########## WP-CLI ##########
RUN curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

########## Cleanup ##########
RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

########## Docker run entrypoint ##########
COPY entrypoint.sh /etc/entrypoint.sh
RUN chmod +x /etc/entrypoint.sh
CMD ["/bin/sh", "-c", "/etc/entrypoint.sh"]

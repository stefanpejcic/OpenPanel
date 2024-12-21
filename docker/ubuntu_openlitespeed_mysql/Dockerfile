FROM ubuntu:latest

ARG OLS_VERSION
ARG PHP_VERSION

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"

ENV TZ=UTC
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y software-properties-common && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
        msmtp \
        ttyd \
        screen \
        mysql-server \
        curl \
        cron \
        pwgen \
        zip \
        unzip \
        wget \
        nano \
        less \
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
EXPOSE 22 3306 7681 8080 7080


########## Litespeed ##########
RUN wget https://openlitespeed.org/packages/openlitespeed-$OLS_VERSION.tgz && \
    tar xzf openlitespeed-$OLS_VERSION.tgz && cd openlitespeed && ./install.sh && \
    echo 'cloud-docker' > /usr/local/lsws/PLAT && rm -rf /openlitespeed && rm /openlitespeed-$OLS_VERSION.tgz

RUN wget -O - https://repo.litespeed.sh | bash

RUN install -y $PHP_VERSION $PHP_VERSION-common $PHP_VERSION-mysql $PHP_VERSION-opcache \
    $PHP_VERSION-curl $PHP_VERSION-imagick $PHP_VERSION-redis $PHP_VERSION-memcached $PHP_VERSION-intl
    
RUN ["/bin/bash", "-c", "if [[ $PHP_VERSION == lsphp7* ]]; then apt-get install $PHP_VERSION-json -y; fi"]

RUN wget -O /usr/local/lsws/admin/misc/lsup.sh \
    https://raw.githubusercontent.com/litespeedtech/openlitespeed/master/dist/admin/misc/lsup.sh && \
    chmod +x /usr/local/lsws/admin/misc/lsup.sh && \
    ln -s /usr/local/lsws/$PHP_VERSION/bin/php /usr/bin/php

ENV PATH="/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/bin"

ADD docker.conf /usr/local/lsws/conf/templates/docker.conf
ADD setup_docker.sh /usr/local/lsws/bin/setup_docker.sh
ADD htpasswd /usr/local/lsws/admin/conf/htpasswd

RUN /usr/local/lsws/bin/setup_docker.sh && rm /usr/local/lsws/bin/setup_docker.sh
RUN chown 994:994 /usr/local/lsws/conf -R
RUN cp -RP /usr/local/lsws/conf/ /usr/local/lsws/.conf/
RUN cp -RP /usr/local/lsws/admin/conf /usr/local/lsws/admin/.conf/
#RUN sed -i "s|fcgi-bin/lsphp|/usr/local/lsws/$PHP_VERSION/bin/lsphp|g" /usr/local/lsws/conf/httpd_config.conf
RUN ["/bin/bash", "-c", "if [[ $PHP_VERSION == lsphp8* ]]; then ln -sf /usr/local/lsws/$PHP_VERSION/bin/lsphp /usr/local/lsws/fcgi-bin/lsphp8; fi"]
RUN ["/bin/bash", "-c", "if [[ $PHP_VERSION == lsphp8* ]]; then ln -sf /usr/local/lsws/fcgi-bin/lsphp8 /usr/local/lsws/fcgi-bin/lsphp; fi"]
RUN ["/bin/bash", "-c", "if [[ $PHP_VERSION == lsphp7* ]]; then ln -sf /usr/local/lsws/$PHP_VERSION/bin/lsphp /usr/local/lsws/fcgi-bin/lsphp7; fi"]
RUN ["/bin/bash", "-c", "if [[ $PHP_VERSION == lsphp7* ]]; then ln -sf /usr/local/lsws/fcgi-bin/lsphp7 /usr/local/lsws/fcgi-bin/lsphp; fi"]





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
    -e 's|^;sendmail_path = .*|sendmail_path = "/usr/bin/msmtp -t"|' \
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
RUN mkdir -p /etc/apache2/ssl/ && cd /etc/apache2/ssl/ && openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -subj "/C=GB/ST=London/L=London/O=Global Security/OU=R&D Department/CN=openpanel.co"  -keyout cert.key  -out cert.crt


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

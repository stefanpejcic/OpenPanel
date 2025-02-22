FROM ubuntu:latest

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"
LABEL db="mysql" webserver="nginx" php="php8.2"
ENV TZ=UTC
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
        ttyd \
        curl \
        cron \
        pwgen \
        zip \
        unzip \
        wget \
        nano \
        less \
        openssh-server \
        mysql-client \
        postgresql-client && \        
        apt-get clean && \
        apt-get autoremove -y && \
        rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


########## EXPOSED PORTS ##########
EXPOSE 22 3306 7681 8080


########## EMAIL ##########
COPY email/msmtprc /etc/msmtprc



########## SSH ##########
ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile


########## SSL #############
#  openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -subj "/C=GB/ST=London/L=London/O=Global Security/OU=R&D Department/CN=openpanel.com"  -keyout cert.key  -out cert.crt
RUN mkdir -p /etc/nginx/ssl/ 
COPY  cert.key /etc/nginx/ssl/cert.key
COPY  cert.crt /etc/nginx/ssl/cert.crt

########## TERMINAL #############
# fix for webterminal: bash: permission denied: /home/user/.bashrc
RUN chmod 755 /root

########## cleanup ##########
RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

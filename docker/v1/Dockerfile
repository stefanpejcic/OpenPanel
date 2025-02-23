FROM ubuntu:latest

LABEL maintainer="stefan@pejcic.rs"
LABEL author="Stefan Pejcic"
LABEL db="mysql" webserver="nginx" php="php8.2"
ENV TZ=UTC
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
        ca-certificates \
        ttyd \
        curl \
        cron \
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
RUN mkdir -p /etc/nginx/ssl/ 
COPY certs/ /etc/nginx/ssl/

########## TERMINAL #############
# fix for webterminal: bash: permission denied: /home/user/.bashrc
RUN chmod 755 /root

########## cleanup ##########
RUN rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

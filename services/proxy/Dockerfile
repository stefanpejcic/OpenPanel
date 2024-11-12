# Use a base image for Rocky Linux
FROM rockylinux:9.3

# Environment variables
ENV DOMAIN openpanel.org
ENV PREVIEW_DOMAIN preview.openpanel.org

# Update system and install necessary tools
RUN dnf update -y && \
    dnf install -y 'dnf-command(copr)' && \
    dnf copr enable @caddy/caddy -y && \
    dnf install -y cronie caddy php php-fpm php-cli php-common

# Copy files
COPY html /var/www/html
COPY delete_cron.sh /var/www/delete_cron.sh
COPY caddy/fullchain.pem /etc/caddy/certs/fullchain.pem
COPY caddy/privkey.pem /etc/caddy/certs/privkey.pem
COPY caddy/Caddyfile /etc/caddy/Caddyfile

# order is important!
RUN find /var/www/html -type f -exec sed -i "s/preview.openpanel.org/$PREVIEW_DOMAIN/g" {} + && \
    find /etc/caddy/Caddyfile -type f -exec sed -i "s/preview.openpanel.org/$PREVIEW_DOMAIN/g" {} + && \
    find /var/www/html -type f -exec sed -i "s/openpanel.org/$DOMAIN/g" {} + && \
    find /etc/caddy/Caddyfile -type f -exec sed -i "s/openpanel.org/$DOMAIN/g" {} +

# Set up permissions
RUN chown -R caddy:caddy /var/www/html && \
    chmod -R 755 /var/www/html/domains && \
    chmod +x /var/www/delete_cron.sh && \
    usermod -aG caddy caddy && \
    #####chown caddy:caddy /run/php-fpm/www.sock && \
    sed -i 's/^user = .*/user = caddy/' /etc/php-fpm.d/www.conf && \
    sed -i 's/^group = .*/group = caddy/' /etc/php-fpm.d/www.conf

# Enable and start services
RUN systemctl enable crond && \
    echo "*/5 * * * * bash /var/www/delete_cron.sh" | crontab -
    
# Cron setup
RUN echo "*/5 * * * * bash /var/www/delete_cron.sh" | crontab -

# Expose necessary ports for Caddy
EXPOSE 80 443

# Run services
CMD ["sh", "-c", "php-fpm && caddy run --config /etc/caddy/Caddyfile --adapter caddyfile"]

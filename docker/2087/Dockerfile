# Use the official Python image as a base
FROM python:3.10-slim

# Set the working directory
WORKDIR /usr/local/admin

# need this fixed since WHMCS does not allow custom port to be defined..
EXPOSE 2087

RUN apt-get update && apt-get install -y \
    procps \
    coreutils \
    sudo \
    wget \
    curl \
    default-mysql-client \
    sqlite3 \
    zip \
    unzip \
    geoip-bin \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash openadmin \
    && adduser openadmin sudo \
    && echo 'openadmin ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# here will do wget or git clone - to avoid rebuilding image
COPY . .

# Download and untar opencli
RUN mkdir -p scripts && \
    cd scripts && \
    wget -O compressed.tar.gz https://storage.googleapis.com/openpanel/0.1.6/get.openpanel.co/downloads/0.1.6/opencli/compressed.tar.gz && \
    tar -xzf compressed.tar.gz && \
    rm compressed.tar.gz && \
    cd ..


COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt


# Make all bash scripts in this directory executable for root only
RUN chown root:root /usr/local/admin/scripts/*
RUN chmod +x -R /usr/local/admin/scripts/*

# Add opencli binary to the PATH
RUN cp /usr/local/admin/scripts/opencli /usr/local/bin/opencli
RUN chmod +x /usr/local/bin/opencli

# Generate a list of commands for opencli
RUN opencli commands

# Set autocomplete for all available opencli commands
RUN echo "# opencli aliases\n\
ALIASES_FILE=\"/usr/local/admin/scripts/aliases.txt\"\n\
generate_autocomplete() {\n\
    awk '{print \$NF}' \"\$ALIASES_FILE\"\n\
}\n\
complete -W \"\$(generate_autocomplete)\" opencli" >> /root/.bashrc


# Create admin user and output credentials to Docker logs
RUN admin_password=$(openssl rand -base64 12 | tr -d '=+/') && \
    password_hash=$(python3 /usr/local/admin/core/users/hash $admin_password) && \
    sqlite3 /ets/openpanel/openadmin/users.db "CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', is_active BOOLEAN DEFAULT 1 NOT NULL);" && \
    sqlite3 /ets/openpanel/openadmin/users.db "INSERT INTO user (username, password_hash, role) VALUES ('admin', \"$password_hash\", 'admin');" && \
    echo "Admin username: admin" && \
    echo "Admin password: $admin_password"


# Gunicorn + redirect stderr to stdout
CMD ["gunicorn", "-c", "/etc/openpanel/openadmin/service/service.config.py", "app:app", "--log-file", "-"]

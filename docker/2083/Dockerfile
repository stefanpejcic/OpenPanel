# Use the official Python image as a base
FROM python:3.10-slim

# Set the working directory
WORKDIR /usr/local/panel

# Expose the port
EXPOSE 2083

# Install necessary packages
RUN apt-get update && apt-get install -y \
    procps \
    coreutils \
    sudo \
    wget \
    curl \
    default-mysql-client \
    zip \
    unzip \
    geoip-bin \
    && rm -rf /var/lib/apt/lists/*

# Create the openpanel user and add to sudoers
RUN useradd -ms /bin/bash openpanel \
    && adduser openpanel sudo \
    && echo 'openpanel ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

# Copy the project files
COPY . .
COPY scripts/ /usr/local/admin/scripts/

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


# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt


# Set the CMD to run gunicorn and redirect stderr to stdout
CMD ["gunicorn", "-c", "/etc/openpanel/openpanel/service/service.config.py", "app:app", "--log-file", "-"]

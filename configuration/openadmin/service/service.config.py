# Gunicorn configuration file
# https://docs.gunicorn.org/en/stable/configure.html#configuration-file
# https://docs.gunicorn.org/en/stable/settings.html

import multiprocessing
from gunicorn.config import Config
import configparser
import os
import re
from pathlib import Path


# From version 1.1.4, we no longer restart admin/user services on configuration changes. Instead, 
# we create a flag file (/root/openadmin_restart_needed) and remind the user via the GUI that a restart 
# is needed to apply the changes. 
# Here, on restart, we check and remove that flag to ensure it’s cleared.
RESTART_FILE_PATH = '/root/openadmin_restart_needed'

# From version 1.2.8, we have an option for admin to ompletelly disable admin panel, if done so,
# we create a flag file (/root/openadmin_is_disabled) and exit 
# Here, on startup, we check if that file exists.
DISABLE_FILE_PATH = '/root/openadmin_is_disabled'

# Function to check if the file exists and remove it
def check_and_remove_restart_file():
    if os.path.exists(RESTART_FILE_PATH):
        try:
            os.remove(RESTART_FILE_PATH)
            print(f"Removed the restart-needed flag for OpenAdmin panel.")
        except Exception as e:
            print(f"Error removing {RESTART_FILE_PATH}: {e}")

# Call the function before starting the Gunicorn server
check_and_remove_restart_file()

if os.path.exists(DISABLE_FILE_PATH):
    print(f"OpenAdmin is disabled! enable it from terminal with 'opencli admin on' or by removing the flag file: {DISABLE_FILE_PATH}")
    sys.exit(1)

# File paths
CADDYFILE_PATH = "/etc/openpanel/caddy/Caddyfile"
CADDY_CERT_DIRS = [
    "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/",
    "/etc/openpanel/caddy/ssl/custom/"
]
DOCKER_COMPOSE_PATH = "/root/docker-compose.yml"


##############

# added in 1.5.1 to chmod+x files on startup
import stat
def make_executable_if_exists(path):
    if os.path.exists(path):
        st = os.stat(path)
        # chmod +x
        if not (st.st_mode & stat.S_IXUSR and st.st_mode & stat.S_IXGRP and st.st_mode & stat.S_IXOTH):
            try:
                os.chmod(path, st.st_mode | 0o111)  # chmod a+x
                print(f"Made {path} executable (+x)")
            except Exception as e:
                print(f"Failed to set +x on {path}: {e}")

def symlink_force(target, link_name):
    try:
        if os.path.islink(link_name) or os.path.exists(link_name):
            os.remove(link_name)  # Remove existing file or symlink
        os.symlink(target, link_name)
        print(f"Created symlink: {link_name} -> {target}")
    except Exception as e:
        print(f"Failed to create symlink {link_name} -> {target}: {e}")

make_executable_if_exists("/etc/openpanel/wordpress/wp-cli.phar")     # wpcli for php containers
make_executable_if_exists("/usr/local/admin/modules/security/csf.pl") # csf gui
make_executable_if_exists("/etc/openpanel/services/watcher.sh")       # reload dns zones
make_executable_if_exists("/etc/openpanel/mysql/scripts/dump.sh")     # mysql export script for backups
make_executable_if_exists("/etc/openpanel/openlitespeed/start.sh")    # overwrites ols entrypoint

symlink_force("/etc/csf/ui/images/", "/usr/local/admin/static/configservercsf")

##############
def get_domain_from_caddyfile():
    domain = None
    in_block = False
    
    # Check if the Caddyfile exists first
    if not os.path.exists(CADDYFILE_PATH):
        print(f"Caddyfile does not exist at {CADDYFILE_PATH}. No SSL will be used.")
        return None

    try:
        with open(CADDYFILE_PATH, "r") as file:
            for line in file:
                line = line.strip()

                if "# START HOSTNAME DOMAIN #" in line:
                    in_block = True
                    continue

                if "# END HOSTNAME DOMAIN #" in line:
                    break

                if in_block:
                    match = re.match(r"^([\w.-]+) \{", line)
                    if match:
                        domain = match.group(1)
                        break
    except Exception as e:
        print(f"Error reading Caddyfile: {e}")

    return domain


def check_ssl_exists(domain):
    for base_dir in CADDY_CERT_DIRS:
        cert_path = os.path.join(base_dir, domain)
        if os.path.exists(cert_path) and os.listdir(cert_path):
            return cert_path
    return None



DOMAIN = get_domain_from_caddyfile()

if DOMAIN and check_ssl_exists(DOMAIN):
    import ssl
    certfile = os.path.join(CADDY_CERT_DIR, DOMAIN, f'{DOMAIN}.crt')
    keyfile = os.path.join(CADDY_CERT_DIR, DOMAIN, f'{DOMAIN}.key')    
    #ssl_version = 'TLS'
    #ca_certs = f'/etc/letsencrypt/live/{hostname}/fullchain.pem'
    cert_reqs = ssl.CERT_NONE
    ciphers = 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH'

bind = ["0.0.0.0:2087"]
backlog = 2048
calculated_workers = multiprocessing.cpu_count() * 2 + 1
max_workers = 10
workers = min(calculated_workers, max_workers)

# Use gevent worker class
worker_class = 'gevent'
worker_connections = 1000
timeout = 30
graceful_timeout = 10
keepalive = 2
max_requests = 1000
max_requests_jitter = 50
pidfile = 'openadmin'

# BUG https://github.com/benoitc/gunicorn/issues/2382
#errorlog = "-"   # Log to stdout
#accesslog = "-"
errorlog = "/var/log/openpanel/admin/error.log"
accesslog = "/var/log/openpanel/admin/access.log"

def ensure_directory(file_path):
    directory = Path(file_path).parent
    directory.mkdir(parents=True, exist_ok=True)

ensure_directory(errorlog)
ensure_directory(accesslog)

def post_fork(server, worker):
    server.log.info("Worker spawned (pid: %s)", worker.pid)

def pre_fork(server, worker):
    pass

def pre_exec(server):
    server.log.info("Forked child, re-executing.")

def when_ready(server):
    server.log.info("Server is ready. Spawning workers")

def worker_int(worker):
    worker.log.info("worker received INT or QUIT signal")


# Allow specific IP addresses
#config.allow_ip = ['192.168.1.100']
forwarded_allow_ips = '*'  # workaround for cloudflare

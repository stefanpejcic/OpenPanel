# Gunicorn configuration file
# https://docs.gunicorn.org/en/stable/configure.html#configuration-file
# https://docs.gunicorn.org/en/stable/settings.html

import multiprocessing
from gunicorn.config import Config
import configparser
import os
import re
from pathlib import Path
import logging
import sys

# ======================================================================
# If dev_mode=on then redirect all prints to '/var/log/openpanel/admin/error.log'
CONFIG_FILE = "/etc/openpanel/openpanel/conf/openpanel.config"
def is_dev_mode():
    if not os.path.exists(CONFIG_FILE):
        return False
    try:
        with open(CONFIG_FILE, "r") as f:
            for line in f:
                if line.strip().lower() == "dev_mode=on":
                    return True
    except Exception:
        pass
    return False

DEV_MODE = is_dev_mode()

loglevel = "error"
errorlog = "/var/log/openpanel/admin/error.log"
accesslog = "/var/log/openpanel/admin/access.log"

if DEV_MODE:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s", datefmt="%Y-%m-%d %H:%M:%S", filename=errorlog, filemode='a')
    logger = logging.getLogger("openpanel")

    class StreamToLogger:
        def __init__(self, logger, log_level=logging.INFO):
            self.logger = logger
            self.log_level = log_level
    
        def write(self, buf):
            for line in buf.rstrip().splitlines():
                if " - " in line:
                    word, msg = line.split(" - ", 1)
                    line = f"[{word}] {msg}"
                self.logger.log(self.log_level, line)
    
        def flush(self):
            pass

    sys.stdout = StreamToLogger(logger, logging.INFO)
    sys.stderr = StreamToLogger(logger, logging.ERROR)
else:
    class DevNull:
        def write(self, _): pass
        def flush(self): pass

    sys.stdout = DevNull()


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
    "/etc/openpanel/caddy/ssl/custom/",
    "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/"   
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
def read_from_caddyfile():
    domain = None
    port = 2087
    in_block = False
    
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
                    match_domain = re.match(r"^([\w.-]+) \{", line)
                    if match_domain:
                        domain = match_domain.group(1)
                        continue

                    match_port = re.search(r"reverse_proxy\s+[\w.-]+:(\d+)", line)
                    if match_port:
                        port = int(match_port.group(1))
                        continue

    except Exception as e:
        print(f"Error reading Caddyfile: {e}")

    if domain == "example.net":
        domain = None
    
    return domain, port


def check_ssl_exists(domain):
    for base_dir in CADDY_CERT_DIRS:
        cert_dir = os.path.join(base_dir, domain)
        cert_file = os.path.join(cert_dir, f'{domain}.crt')
        key_file = os.path.join(cert_dir, f'{domain}.key')
        if os.path.exists(cert_file) and os.path.exists(key_file):
            cert_type = "letsencrypt" if "letsencrypt" in base_dir else "custom"
            return cert_file, key_file, cert_type            
    return None, None, None


DOMAIN, PORT = read_from_caddyfile()

certfile, keyfile, type = (None, None, None)
if DOMAIN:
    certfile, keyfile, cert_type = check_ssl_exists(DOMAIN)
    if certfile and keyfile:
        print(f"HTTPS - {cert_type} certificate is configured.")
        import ssl
        #ssl_version = 'TLS'
        cert_reqs = ssl.CERT_NONE
        ciphers = 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH'
    else:
        print("HTTP - Domain is set but no certificate exists, point domain A record to issue LetsEcrypt SSL or add custom cert: https://openpanel.com/docs/articles/server/how-to-set-custom-ssl-openpanel-webmail/")
else:
    print(f"HTTP - Using IP address for panel access, use 'opencli domain <DOMAIN_NAME>' to set a domain.")

if PORT != 2087:
    print(f"Custom port will be used for OpenAdmin service: {PORT}")

bind = [f"0.0.0.0:{PORT}"]
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

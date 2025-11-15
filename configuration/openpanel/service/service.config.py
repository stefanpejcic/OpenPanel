################################### NOTICE ####################################
#                                                                             #
# Manually modifying this file is not recommended!                            #
#                                                                             #
# This gunicorn configuration file is often overwritten on updates            #
#                                                                             #
# https://docs.gunicorn.org/en/stable/configure.html#configuration-file       #
# https://docs.gunicorn.org/en/stable/settings.html                           #
#                                                                             #
###############################################################################

import multiprocessing
import os
import re
import yaml  # pip install pyyaml
from pathlib import Path
import subprocess
import logging
import sys

# ======================================================================
# If dev_mode=on then redirect all prints to 'docker logs openpanel'
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

if DEV_MODE:
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger("openpanel")

    class StreamToLogger:
        def __init__(self, logger, log_level=logging.INFO):
            self.logger = logger
            self.log_level = log_level

        def write(self, buf):
            for line in buf.rstrip().splitlines():
                self.logger.log(self.log_level, line.rstrip())

        def flush(self):
            pass

    sys.stdout = StreamToLogger(logger, logging.INFO)
    sys.stderr = StreamToLogger(logger, logging.ERROR)
else:
    class DevNull:
        def write(self, _): pass
        def flush(self): pass

    sys.stdout = DevNull()
    #sys.stderr = DevNull()



# ====================================================================== #
# From version 1.6.3, we allow executing a custom script on startup
CUSTOM_SCRIPT = "/root/openpanel_run_on_startup"
if os.path.exists(CUSTOM_SCRIPT) and os.path.isfile(CUSTOM_SCRIPT):
    print(f"Executing custom script: {CUSTOM_SCRIPT} with BASH.")
    try:
        subprocess.run(["bash", CUSTOM_SCRIPT], check=True)
        print(f"Executed custom script successfully.")
    except subprocess.CalledProcessError as e:
        print(f"Error executing {CUSTOM_SCRIPT}: {e}")

# ====================================================================== #
# From version 1.1.4, we no longer restart admin/user services on configuration changes. Instead, 
# we create a flag file (/root/openadmin_restart_needed) and remind the user via the GUI that a restart 
# is needed to apply the changes. 
# Here, on restart, we check and empty that flag to ensure notification in GUI is cleared.
RESTART_FILE_PATH = '/root/openpanel_restart_needed'

def empty_flag_file():
    if os.path.exists(RESTART_FILE_PATH):
        try:
            with open(RESTART_FILE_PATH, 'r+'):
                pass
            with open(RESTART_FILE_PATH, 'w') as f:
                f.truncate(0)
        except Exception as e:
            print(f"Error clearing {RESTART_FILE_PATH}: {e}")

empty_flag_file()

# ====================================================================== #
# SSL
CADDYFILE_PATH = "/etc/openpanel/caddy/Caddyfile"
CADDY_CERT_DIRS = [
    "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/",
    "/etc/openpanel/caddy/ssl/custom/"
]
DOCKER_COMPOSE_PATH = "/root/docker-compose.yml"

def get_domain_from_caddyfile():
    domain = None
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
                    match = re.match(r"^([\w.-]+) \{", line)
                    if match:
                        domain = match.group(1)
                        break
    except Exception as e:
        print(f"Error reading Caddyfile: {e}")
    
    print(f"Domain name set for OpenPanel: {domain}")
    return domain


def check_ssl_exists(domain):
    for base_dir in CADDY_CERT_DIRS:
        cert_path = os.path.join(base_dir, domain)
        if os.path.exists(cert_path) and os.listdir(cert_path):
            return cert_path
    return None


DOMAIN = get_domain_from_caddyfile()
PORT = "2083"
ssl_cert_path = None
if DOMAIN:
    ssl_cert_path = check_ssl_exists(DOMAIN)

if DOMAIN and check_ssl_exists(DOMAIN):
    print(f"Custom domain is set and certificate exists, HTTPS will be used.")
    import ssl
    certfile = os.path.join(ssl_cert_path, f"{DOMAIN}.crt")
    keyfile = os.path.join(ssl_cert_path, f"{DOMAIN}.key")
    keyfile = keyfile
    certfile = certfile
    print(f"Certificate file: {certfile}")
    print(f"Certificate key:  {keyfile}")
    ssl_version = ssl.PROTOCOL_TLS
    cert_reqs = ssl.CERT_NONE
    ciphers = 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH'

# ======================================================================
# Performance
bind = [f"0.0.0.0:{PORT}"]
backlog = 2048
calculated_workers = multiprocessing.cpu_count() * 2 + 1
max_workers = 10
workers = min(calculated_workers, max_workers)
worker_class = 'gevent'
worker_connections = 1000
timeout = 10
graceful_timeout = 10
keepalive = 2
max_requests = 1000
max_requests_jitter = 50
pidfile = 'openpanel'

# ======================================================================
# Create Log files
errorlog = "/var/log/openpanel/user/error.log"
accesslog = "/var/log/openpanel/user/access.log"

def ensure_directory(file_path):
    directory = Path(file_path).parent
    directory.mkdir(parents=True, exist_ok=True)

print(f"Creating log files..")
ensure_directory(errorlog)
ensure_directory(accesslog)


# ======================================================================
# SERVER
def post_fork(server, worker):
    server.log.info("Worker spawned (pid: %s)", worker.pid)

def pre_fork(server, worker):
    pass

def pre_exec(server):
    server.log.info("Forked child, re-executing.")

def when_ready(server):
    server.log.info("Server is ready. Spawning workers")
    try:
        cmd = [
            "docker", "--context=default", "exec", "openpanel_redis",
            "bash", "-c",
            "redis-cli --raw KEYS 'flask_cache_*' | xargs -r redis-cli DEL"
        ]
        result = subprocess.run(cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        server.log.info("Redis cache cleared:\n%s", result.stdout.strip())
    except subprocess.CalledProcessError as e:
        server.log.error("Failed to clear Redis cache:\n%s", e.stderr.strip())
    except Exception as e:
        server.log.error("Unexpected error clearing Redis cache: %s", str(e))

def worker_int(worker):
    worker.log.info("worker received INT or QUIT signal")

# ======================================================================
# Allow specific IP addresses
#config.allow_ip = ['192.168.1.100']
forwarded_allow_ips = '*'  # for cloudflare

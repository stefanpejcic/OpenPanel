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
import threading
from pathlib import Path
import subprocess
import logging
import sys

# ======================================================================
# Read config file
CONFIG_FILE = "/etc/openpanel/openpanel/conf/openpanel.config"
def is_config_enabled(setting: str) -> bool:
    if not os.path.exists(CONFIG_FILE):
        return False

    try:
        with open(CONFIG_FILE, "r") as f:
            for line in f:
                if line.strip().lower() == setting.strip().lower():
                    return True
    except Exception:
        pass

    return False

# ======================================================================
# If dev_mode=on then redirect all prints to 'docker logs openpanel'
DEV_MODE = is_config_enabled("dev_mode=on")

if DEV_MODE:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s", datefmt="%Y-%m-%d %H:%M:%S")
    logger = logging.getLogger("openpanel")

    class StreamToLogger:
        def __init__(self, logger, log_level=logging.INFO):
            self.logger = logger
            self.log_level = log_level
    
        def write(self, buf):
            if not buf:
                return
            for line in buf.rstrip().splitlines():
                if " - " in line:
                    word, msg = line.split(" - ", 1)
                    line = f"[{word}] {msg}"
                self.logger.log(self.log_level, line)
    
        def flush(self):
            pass
    
        def isatty(self):
            return False
    
        @property
        def closed(self):
            return False

    sys.stdout = StreamToLogger(logger, logging.INFO)
    sys.stderr = StreamToLogger(logger, logging.ERROR)
else:
    class DevNull:
        def write(self, _): pass
        def flush(self): pass

    sys.stdout = DevNull()


# ====================================================================== #
# If screenshots=local then install playwright
LOCAL_SCREENSHOTS = is_config_enabled("screenshots=local")

def install_playwright():
    print("Screenshots API is set to local, installing Playwright and dependencies..")
    try:
        subprocess.run(["apt-get", "update"], check=True)

        dependencies = [
            "libatspi2.0-0","libx11-6", "libxcomposite1", "libxdamage1"
            # TODO: remove after 1.8.0
        ]
        subprocess.run(["apt-get", "install", "-y"] + dependencies, check=True)
        subprocess.run(["playwright", "install"], check=True)
        print("Playwright installation finished.")
    except subprocess.CalledProcessError as e:
        print(f"Error installing Playwright: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")

if LOCAL_SCREENSHOTS:
    threading.Thread(target=install_playwright, daemon=True).start()

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

def get_domain():
    try:
        result = subprocess.run(["opencli", "domain"], capture_output=True, text=True, check=True)
        output = result.stdout.strip()
        ip_pattern = re.compile(r"^(?:\d{1,3}\.){3}\d{1,3}$")
        if ip_pattern.match(output):
            return None
        return output
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {e}")
        return None

def get_port():
    try:
        result = subprocess.run(["opencli", "port"], capture_output=True, text=True, check=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        return None

def check_ssl_exists(domain):
    for base_dir in CADDY_CERT_DIRS:
        cert_path = os.path.join(base_dir, domain)
        if os.path.exists(cert_path) and os.listdir(cert_path):
            return cert_path
    return None

DOMAIN = get_domain()
ssl_cert_path = None
if DOMAIN:
    ssl_cert_path = check_ssl_exists(DOMAIN)

if DOMAIN and ssl_cert_path:
    PORT = get_port()
    if PORT == '443':
        print(f"Custom domain is set and certificate exists, but due to 443 port, HTTPS will NOT be used, reverse proxy should handle SSL.")
    else:
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
bind = [f"0.0.0.0:2083"]
backlog = 2048
calculated_workers = multiprocessing.cpu_count() * 2 + 1
max_workers = 10
workers = min(calculated_workers, max_workers)
worker_class = 'gthread'
worker_connections = 1000
timeout = 360
graceful_timeout = 10
keepalive = 5
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
        server.log.info("Redis cache cleared: %s", result.stdout.strip())
    except subprocess.CalledProcessError as e:
        server.log.error("Failed to clear Redis cache: %s", e.stderr.strip())
    except Exception as e:
        server.log.error("Unexpected error clearing Redis cache: %s", str(e))

def worker_int(worker):
    worker.log.info("worker received INT or QUIT signal")

# ======================================================================
# Allow specific IP addresses
#config.allow_ip = ['192.168.1.100']
forwarded_allow_ips = '*'  # for cloudflare

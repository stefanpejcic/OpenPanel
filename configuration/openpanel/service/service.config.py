# Gunicorn configuration file
# https://docs.gunicorn.org/en/stable/configure.html#configuration-file
# https://docs.gunicorn.org/en/stable/settings.html

import multiprocessing
from gunicorn.config import Config
import configparser
import os
from pathlib import Path

CONFIG_FILE_PATH = '/etc/openpanel/openpanel/conf/openpanel.config'

def read_config():
    config = configparser.ConfigParser()
    if os.path.exists(CONFIG_FILE_PATH):
        config.read(CONFIG_FILE_PATH)
    return config

def get_custom_port():
    config = read_config()
    return int(config.get('DEFAULT', 'port', fallback=2083))

def get_ssl_status():
    config = read_config()
    return config.getboolean('DEFAULT', 'ssl', fallback=False)

if get_ssl_status():
    import ssl
    import socket
    hostname = socket.gethostname()

    certfile = f'/etc/letsencrypt/live/{hostname}/fullchain.pem'
    keyfile = f'/etc/letsencrypt/live/{hostname}/privkey.pem'
    ssl_version = 'TLS'
    ca_certs = f'/etc/letsencrypt/live/{hostname}/fullchain.pem'
    cert_reqs = ssl.CERT_NONE
    ciphers = 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH'

bind = ["0.0.0.0:" + str(get_custom_port())]
backlog = 2048
calculated_workers = multiprocessing.cpu_count() * 2 + 1
max_workers = 10
workers = min(calculated_workers, max_workers)
worker_class = 'gevent'
worker_connections = 1000
timeout = 30
graceful_timeout = 30
keepalive = 2
max_requests = 1000
max_requests_jitter = 50
pidfile = 'openpanel'
# BUG https://github.com/benoitc/gunicorn/issues/2382
#errorlog = "-"   # Log to stdout
#accesslog = "-"
errorlog = "/var/log/openpanel/user/error.log"
accesslog = "/var/log/openpanel/user/access.log"

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
forwarded_allow_ips = '*'  # for cloudflare

# faq.py
import os
import requests
import click
import re

CONFIG_FILE_PATH = '/etc/openpanel/openpanel/conf/openpanel.config'
IP_SERVERS = ["https://ip.openpanel.com"]

def read_config():
    config = {}
    if os.path.isfile(CONFIG_FILE_PATH):
        with open(CONFIG_FILE_PATH, 'r') as file:
            current_section = None
            for line in file:
                line = line.strip()
                if line.startswith("[") and line.endswith("]"):
                    current_section = line[1:-1]
                elif current_section == "DEFAULT" and "=" in line:
                    key, value = line.split("=", 1)
                    config[key.strip()] = value.strip()
    return config

def get_ssl_status(config):
    return config.get("ssl", "no").lower() == "yes"

def get_force_domain(config):
    force_domain = config.get("force_domain")
    if force_domain:
        return force_domain
    return get_public_ip()

def get_public_ip():
    for server in IP_SERVERS:
        try:
            response = requests.get(server, timeout=2)
            if response.status_code == 200:
                ip = response.text.strip()
                if re.match(r"^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$", ip):
                    return ip
        except requests.RequestException:
            continue
    return os.popen("hostname -I | awk '{print $1}'").read().strip()

def generate_admin_url():
    config = read_config()
    if get_ssl_status(config):
        hostname = get_force_domain(config)
        return f"https://{hostname}:2087/"
    ip = get_public_ip()
    return f"http://{ip}:2087/"

@click.command()
def faq():
    """Display frequently asked questions."""
    admin_url = generate_admin_url()
    click.echo(click.style("Frequently Asked Questions", fg="blue", bold=True))
    
    click.echo(f"{click.style('1.', fg='magenta')} What is the login link for admin panel?")
    click.echo(f"LINK: {click.style(admin_url, fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('2.', fg='magenta')} How to restart OpenAdmin or OpenPanel services?")
    click.echo(f"- OpenPanel: {click.style('docker restart openpanel', fg='red')}")
    click.echo(f"- OpenAdmin: {click.style('service admin restart', fg='red')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('3.', fg='magenta')} How to reset admin password?")
    click.echo(f"execute command {click.style('opencli admin password USERNAME NEW_PASSWORD', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('4.', fg='magenta')} How to create new admin account?")
    click.echo(f"execute command {click.style('opencli admin new USERNAME PASSWORD', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('5.', fg='magenta')} How to list admin accounts?")
    click.echo(f"execute command {click.style('opencli admin list', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('6.', fg='magenta')} How to check OpenPanel version?")
    click.echo(f"execute command {click.style('opencli --version', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('7.', fg='magenta')} How to update OpenPanel?")
    click.echo(f"execute command {click.style('opencli update --force', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('8.', fg='magenta')} How to disable automatic updates?")
    click.echo(f"execute command {click.style('opencli config update autoupdate off', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))

    click.echo(f"{click.style('9.', fg='magenta')} Where are the logs?")
    click.echo(f"- User panel errors:    {click.style('docker logs openpanel', fg='green')}")
    click.echo(f"- Admin panel errors:   {click.style('/var/log/openpanel/admin/error.log', fg='green')}")
    click.echo(f"- Admin panel access:   {click.style('/var/log/openpanel/admin/access.log', fg='green')}")
    click.echo(f"- Admin API accesss:    {click.style('/var/log/openpanel/admin/api.log', fg='green')}")
    click.echo(f"- Admin logins log:     {click.style('/var/log/openpanel/admin/login.log', fg='green')}")
    click.echo(f"- Admin notifications:  {click.style('/var/log/openpanel/admin/notifications.log', fg='green')}")
    click.echo(f"- Admin cron execution: {click.style('/var/log/openpanel/admin/cron.log', fg='green')}")
    click.echo(click.style("-" * 60, fg="blue"))


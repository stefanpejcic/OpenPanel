# user.py
import click
import subprocess
import json

# Configuration for MySQL connection from /etc/my.cnf
CONFIG_FILE_PATH = '/etc/my.cnf'  # Use the provided config file

def ensure_jq_installed():
    """Check if jq is installed; if not, attempt to install it."""
    if subprocess.call(['command', '-v', 'jq'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL) != 0:
        click.echo("jq is not installed. Installing...")
        try:
            subprocess.run(['sudo', 'apt-get', 'install', '-y', 'jq'], check=True)
        except subprocess.CalledProcessError:
            click.echo("Error: Failed to install jq. Please install it manually and try again.")
            raise SystemExit(1)

def fetch_user_count(json_output):
    """Fetch the total user count from the database."""
    user_count_command = [
        'mysql', '--defaults-extra-file={}'.format(CONFIG_FILE_PATH),
        '-se', "SELECT COUNT(*) FROM users"
    ]
    
    user_count = subprocess.check_output(user_count_command).strip()
    
    if json_output:
        click.echo(json.dumps({"total_users": int(user_count)}))
    else:
        click.echo("Total number of users: {}".format(user_count.decode('utf-8')))

def fetch_users_data(json_output):
    """Fetch user information from the database."""
    users_command = [
        'mysql', '--defaults-extra-file={}'.format(CONFIG_FILE_PATH),
        '-e', "SELECT users.id, users.username, users.email, plans.name AS plan_name, users.registered_date FROM users INNER JOIN plans ON users.plan_id = plans.id;"
    ]

    users_data = subprocess.check_output(users_command).decode('utf-8').strip().split('\n')

    if json_output:
        # Create JSON output
        users_list = []
        for line in users_data[1:]:
            if line:  # Skip empty lines
                id, username, email, plan_name, registered_date = line.split('\t')
                users_list.append({
                    "id": id,
                    "username": username,
                    "email": email,
                    "plan_name": plan_name,
                    "registered_date": registered_date
                })
        click.echo(json.dumps(users_list, indent=4))
    else:
        # Print data in tabular format
        if users_data:
            click.echo("\n".join(users_data))
        else:
            click.echo("No users.")

@click.command(name='user-list')
@click.option('--json', is_flag=True, help='Output data in JSON format.')
@click.option('--total', is_flag=True, help='Display total number of users.')
def list_users(json, total):
    """List users or total count of users."""
    if total:
        fetch_user_count(json)
    else:
        fetch_users_data(json)

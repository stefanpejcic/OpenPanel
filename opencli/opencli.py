import click
import logging
import os
from faq import faq 
from user import list_users

# Set up logging
log_file = "opencli_commands.log"
logging.basicConfig(filename=log_file, level=logging.INFO, 
                    format='%(asctime)s - %(message)s')

# Global variable to keep track of command usage
command_usage = {}

@click.group(invoke_without_command=True)
@click.option("--debug", is_flag=True, help="Enable debug mode.")
@click.pass_context
def opencli(ctx, debug):
    """OpenCLI: A command-line interface for managing OpenPanel server."""
    ctx.ensure_object(dict)
    ctx.obj["DEBUG"] = debug
    if debug:
        click.echo("Debug mode is ON")

@click.command()
def commands():
    """List all available commands."""
    click.echo("Available commands:")
    for command_name, command in opencli.commands.items():
        click.echo(f"- {command_name}: {command.help}")
    # Print registered commands for debugging
    click.echo("Registered commands:")
    for command in opencli.commands:
        click.echo(f"  - {command}")

def log_command(command_name, *args):
    """Log the command execution."""
    command_line = f"{command_name} {' '.join(args)}"
    logging.info(command_line)



################ version ################
@opencli.command(name='version', short_help='Show the OpenPanel version')
def version():
    """Check and display the installed OpenPanel version."""
    VERSION_FILE_PATH = "/usr/local/panel/version"
    if os.path.isfile(VERSION_FILE_PATH):
        with open(VERSION_FILE_PATH, "r") as file:
            local_version = file.read().strip()
            click.echo(local_version)
    else:
        click.echo('{"error": "Local version file not found"}', err=True)

############# version ################

# Register the users command
opencli.add_command(faq)
opencli.add_command(list_users)

if __name__ == "__main__":
    opencli(obj={})

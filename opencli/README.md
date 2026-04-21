# OpenCLI  

OpenCLI is the command-line interface for managing [OpenPanel](https://openpanel.com/).  

On installation, the `opencli.sh` script is added to the system path, and `opencli --help` generates a list of all commands to be included in `.bashrc` for autocomplete.  

All scripts from `/usr/local/opencli/` can be accessed using the `opencli` command by replacing `/` with `-`.  

For example: to run user/add.sh script you would type: `opencli user-add`.

## Updates

OpenPanel is proud of the modularity, so you can independently update just the OpenCLI when needed.

To update OpenCLI:

```sh
opencli update --cli
```

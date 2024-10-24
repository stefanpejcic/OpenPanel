# cPanel 2 OpenPanel account import
Free OpenPanel module to import cPanel backup in OpenPanel

Maintained by [CodeWithJuber](https://github.com/CodeWithJuber)

## Features

Currently suported for import:
```
├─ DOMAINS:
│  ├─ Primary domain, Addons, Aliases and Subdomains
│  ├─ SSL certificates
│  ├─ Domains access logs (Apache domlogs)
│  └─ DNS zones
├─ WEBSITES:
│  └─ WordPress instalations from WPToolkit & Softaculous 
├─ DATABASES:
│    ├─ Remote access to MySQL
│    └─ MySQL databases, users and grants
├─ PHP:
│    └─ Installed version from Cloudlinux PHP Selector
├─ FILES
├─ CRONS
├─ SSH
│   ├─ Remote SSH access
│   ├─ SSH password
│   └─ SSH keys
└─ ACCOUNT
    ├─ Notification preferences
    ├─ cPanel account creation date
    └─ locale

***emails, ftp, nodejs/python, postgres are not yet supported***
```


## Usage

Run the script with sudo privileges:

```
git clone https://github.com/stefanpejcic/cPanel-to-OpenPanel
```

```
bash cPanel-to-OpenPanel/cp-import.sh --backup-location /path/to/cpanel_backup.file --plan-name "default_plan_nginx"
```

## Parameters

- `--backup-location`: Path to the cPanel backup file (required)
- `--plan-name`:       Name of the hosting plan in OpenPanel (required)
- `--dry-run`:         extract archive and display data without actually importing account (optional)

## Important Notes

- This script should be run on the OpenPanel server where you want to import the cPanel backup.
- The script requires internet access to install dependencies if they are not already present.
- Large backups may take a considerable amount of time to process.
- Some manual configuration may be required after the migration, depending on the complexity of the cPanel account.

## Troubleshooting

If you encounter any issues:

1. Check the script's output for error messages.
2. Verify that all prerequisites are met.
3. Ensure you have sufficient disk space and system resources.
4. Check the OpenPanel logs for any additional error information.

## Contributing

Contributions to improve the script are welcome. Please feel free to submit issues or pull requests.

## License

[MIT License](LICENSE)

## Disclaimer

This script is provided as-is, without any guarantees. Always test thoroughly in a non-production environment before using in production.

# cPanel 2 OpenPanel account import
Import cPanel full account backup to OpenPanel

Maintained by [CodeWithJuber](https://github.com/CodeWithJuber)

## Features

Currently suported for import:
```
├─ DOMAINS:
│  ├─ Primary domain, Addons, Aliases and Subdomains
│  ├─ DNS zones
│  ├─ SSL certificates
│  ├─ Modsecurity status
│  └─ Access logs (Apache domlogs)
├─ WEBSITES:
│  └─ WordPress instalations from WP Toolkit & Softaculous 
├─ DATABASES:
│    ├─ MySQL databases, users and grants
│    ├─ PostgreSQL databases, users and grants
│    └─ Remote access to MySQL
├─ PHP:
│    └─ Installed version from Cloudlinux PHP Selector
├─ FILES
├─ CRONS
└─ ACCOUNT
    ├─ Account Password
    ├─ Notification preferences
    ├─ Creation date
    └─ Locale (Language)

***ftp accounts and nodejs/python apps are not yet supported***
```


## Usage

Run the script with sudo privileges:

```
git clone https://github.com/stefanpejcic/cPanel-to-OpenPanel
```

```
bash cPanel-to-OpenPanel/cp-import.sh --backup-location=/path/to/cpanel_backup.file --plan-name='Standard plan'
```

## Parameters

- `--backup-location=` Path to the cPanel backup file
- `--plan-name=`      Name of the hosting plan in OpenPanel
- `--dry-run`         extract archive and display data without actually importing account

## Troubleshooting

If you encounter any issues:

1. Check the script's output for error messages.
2. If it is a bug woth the sceipt not properly handling some cpanel files, please open an issue.

## Contributing

Contributions to improve the script are welcome. Please feel free to submit issues or pull requests.

## License

[MIT License](LICENSE)

## Disclaimer

This script is provided as-is, without any guarantees. Always test thoroughly in a non-production environment before using in production.

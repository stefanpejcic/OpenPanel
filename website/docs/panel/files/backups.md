---
sidebar_position: 6
---

# Backups

In OpenPanel, backups are configurable directly **by the panel users**, unlike other panels where backups are managed by the admin.

This empowers users to define their own backup schedules, choose exactly what to back up, and select from a wide range of supported destinations.
As a result, administrators have fewer tasks to manage, and users gain greater control and flexibility.

![backups.png](/img/panel/v2/backups.png)

## Destinations

OpenPanel supports the following destinations:

- [**S3 storage**](#s3) - AWS S3, Filebase, MinIO, etc.
- [**WebDAV**](#webdav) - Remote backups using WebDAV.
- [**SSH**](#azure) - Remote backups via SSH to another server.
- [**Azure**](#azure) - Remote backups to Azure Blob Storage.
- [**Dropbox**](/#dropbox) - Remote backups to Dropbox cloud storage.

User can select destination from **Backups > Destinations** page.

![destinations.png](/img/panel/v2/destinations.png)

Once backup is selected, it's options are available on **Backups > Settings page**.

### SSH/SFTP

| Setting | Description |
|----------|----------|
| **SSH_HOST_NAME**    | The FQDN of the remote SSH server, example: `server.local`  |
| **SSH_PORT**    | The port of the remote SSH server  |
| **SSH_REMOTE_PATH**    | The Directory to place the backups to on the SSH server. Example: `/home/user/backups`  |
| **SSH_USER**    | The username for the SSH server. Example: `user`  |
| **SSH_PASSWORD**    | The password for the SSH server. Example: `password`  |
| **SSH_IDENTITY_FILE**    | The private key path inside `/var/www/html/` for SSH server. Non-RSA keys (e.g. ed25519) will also work.  |
| **SSH_IDENTITY_PASSPHRASE**    | The passphrase for the identity file if applicable. Example: `pass`  |


### S3

| Setting | Description |
|----------|----------|
| **AWS_S3_BUCKET_NAME**    | The name of the remote bucket that should be used for storing backups. If this is not set, no remote backups will be stored. Example: `backup-bucket` |
| **AWS_S3_PATH**    | If you want to store the backup in a non-root location on your bucket  you can provide a path. The path must not contain a leading slash. Example: `my/backup/location`  |
| **AWS_ACCESS_KEY_ID** **AWS_SECRET_ACCESS_KEY**   | Define credentials for authenticating against the backup storage and a bucket name. Although all of these keys are `AWS`-prefixed, the setup can be used with any S3 compatible storage. |
| **AWS_IAM_ROLE_ENDPOINT**    | Instead of providing static credentials, you can also use IAM instance profiles or similar to provide authentication. Some possible configuration options on AWS: `- EC2: http://169.254.169.254` `- ECS: http://169.254.170.2` |
| **AWS_ENDPOINT**    | This is the FQDN of your storage server, e.g. `storage.example.com`. If you need to set a specific (non-https) protocol, you will need to use the `AWS_ENDPOINT_PROTO` option. The default value points to the standard AWS S3 endpoint. Example: `s3.amazonaws.com`  |
| **AWS_ENDPOINT_PROTO**    | The protocol to be used when communicating with your S3 storage server. Defaults to "https". You can set this to "http" when communicating with a different seld-hosted s3 storage. |
| **AWS_ENDPOINT_INSECURE**    | Setting this variable to `true` will disable verification of SSL certificates for `AWS_ENDPOINT`. You shouldn't use this unless you use self-signed certificates for your remote storage backend. This can only be used when `AWS_ENDPOINT_PROTO` is set to `https`.  |
| **AWS_ENDPOINT_CA_CERT**    | If you wish to use self signed certificates your S3 server, you can pass the location of a PEM encoded CA certificate and it will be used for validating your certificates. Alternatively, pass a PEM encoded string containing the certificate. Example: `/var/www/html/cert.pem` |
| **AWS_STORAGE_CLASS**    | Setting a value for this key will change the S3 storage class header. Default behavior is to use the standard class when no value is given. Example: `GLACIER` |
| **AWS_PART_SIZE**    | Setting this variable will change the S3 default part size for the copy step. This value is useful when you want to upload large files. Note: While using Scaleway as S3 provider, be aware that the parts counter is set to 1.000. While Minio uses a hard coded value to 10.000. As a workaround, try to set a higher value. Defaults to `16` (MB) if unset (from minio), you can set this value according to your needs. The unit is in MB and an integer. |


### WebDAV

| Setting | Description |
|----------|----------|
| **WEBDAV_URL**    | The URL of the remote WebDAV server. Example: `https://webdav.example.com` |
| **WEBDAV_PATH**    | The Directory to place the backups to on the WebDAV server. If the path is not present on the server it will be created. Example: `/my/directory/` |
| **WEBDAV_USERNAME**    | The username for the WebDAV server Example: `user` |
| **WEBDAV_PASSWORD**    | The password for the WebDAV server. Example: `password` |
| **WEBDAV_URL_INSECURE**    |  Setting this variable to "true" will disable verification of SSL certificates for WEBDAV_URL. You shouldn't use this unless you use self-signed certificates for your remote storage backend. |

### Azure

| Setting | Description |
|----------|----------|
| **AZURE_STORAGE_ACCOUNT_NAME**    | The credential's account name when using Azure Blob Storage. This has to be set when using Azure Blob Storage.  Example: `account-name` |
| **AZURE_STORAGE_PRIMARY_ACCOUNT_KEY**    | The credential's primary account key when using Azure Blob Storage. If this is not given, the command tries to fall back to using a connection string (if given) or a managed identity (if neither is set). |
| **AZURE_STORAGE_CONNECTION_STRING**    | A connection string for accessing Azure Blob Storage. If this is not given, the command tries to fall back to using a primary account key (if given) or a managed identity (if neither is set). |
| **AZURE_STORAGE_CONTAINER_NAME**    | The container name when using Azure Blob Storage.  Example: `container-name` |
| **AZURE_STORAGE_ENDPOINT**    | The service endpoint when using Azure Blob Storage. This is a template that can be passed the account name.  Example: `https://{{ .AccountName }}.blob.core.windows.net/` |
| **AZURE_STORAGE_ACCESS_TIER**    | The access tier when using Azure Blob Storage. Possible values are [listed here](https://github.com/Azure/azure-sdk-for-go/blob/sdk/storage/azblob/v1.3.2/sdk/storage/azblob/internal/generated/zz_constants.go#L14-L30) Example: `Cold` |


### Dropbox

| Setting | Description |
|----------|----------|
| **DROPBOX_REMOTE_PATH**    | Absolute remote path in your Dropbox where the backups shall be stored. Note: Use your app's subpath in Dropbox, if it doesn't have global access.  Example: `/my/directory` |
| **DROPBOX_APP_KEY**  **DROPBOX_APP_SECRET**   | App key and app secret from your app created at [https://www.dropbox.com/developers/apps](https://www.dropbox.com/developers/apps) |
| **DROPBOX_CONCURRENCY_LEVEL**    | Number of concurrent chunked uploads for Dropbox. Values above 6 usually result in no enhancements. |
| **DROPBOX_REFRESH_TOKEN**    |  Refresh token to request new short-lived tokens (OAuth2) |


Set up Dropbox storage backend:

1. Create a new Dropbox App in the [App Console](https://www.dropbox.com/developers/apps)
2. Open your new Dropbox App and copy the `DROPBOX_APP_KEY` and `DROPBOX_APP_SECRET`
3. Click on `Permissions` in your app and make sure, that the following permissions are cranted (or more):
   - `files.metadata.write`
   - `files.metadata.read`
   - `files.content.write`
   - `files.content.read`
4. Replace APPKEY in `https://www.dropbox.com/oauth2/authorize?client_id=APPKEY&token_access_type=offline&response_type=code `with the app key from step 2
5. Visit the URL and confirm the access of your app. This gives you an `auth code` -> save it somewhere!
6. Replace AUTHCODE, APPKEY, APPSECRET accordingly and perform the request from your terminal:
   ```
   curl https://api.dropbox.com/oauth2/token \
   -d code=AUTHCODE \
   -d grant_type=authorization_code \
   -d client_id=APPKEY \
   -d client_secret=APPSECRET
   ```
7. Execute the request. You will get a JSON formatted reply. Use the value of the `refresh_token` for the last setting `DROPBOX_REFRESH_TOKEN`
8. You should now have `DROPBOX_APP_KEY`, `DROPBOX_APP_SECRET` and `DROPBOX_REFRESH_TOKEN` set. These don’t expire.

:::info
*Note*: Using the “Generated access token” in the app console is not supported, as it is only very short lived and therefore not suitable for an automatic backup solution. The refresh token handles this automatically - the setup procedure above is only needed once.
:::

:::danger
Important: If you chose `App folder` access during the creation of your Dropbox app in step 1 above, `DROPBOX_REMOTE_PATH` will be a relative path under the App folder! (For example, DROPBOX_REMOTE_PATH=/somedir means the backup file will be uploaded to /Apps/myapp/somedir) On the other hand if you chose Full Dropbox access, the value for `DROPBOX_REMOTE_PATH` will represent an absolute path inside your Dropbox storage area. (Still considering the same example above, the backup file will be uploaded to /somedir in your Dropbox root)
:::


## Encryption

Bakcup encryption is by default disbaled and needs to be enabled by the Administrator first.

All of the encryption options are mutually exclusive. Provide a single option for the encryption scheme of your choice.


| Setting | Description |
|----------|----------|
| **GPG_PASSPHRASE**    | Backups can be encrypted symmetrically using gpg in case a passphrase is given. |
| **GPG_PUBLIC_KEY_RING**    |  Backups can be encrypted asymmetrically using gpg in case publickeys are given. You can use pipe syntax to pass a multiline value. |
| **AGE_PASSPHRASE**    | Backups can be encrypted symmetrically using age in case a passphrase is given. |
| **AGE_PUBLIC_KEYS**    | Backups can be encrypted asymmetrically using age in case publickeys are given. Multiple keys need to be provided as a comma separated list. Right now, this supports `age` and `ssh` keys. |

## Rotation

:::danger
**IMPORTANT, PLEASE READ THIS BEFORE USING THIS FEATURE**:
The mechanism used for pruning old backups is not very sophisticated and applies its rules to **all files in the target directory** by default, which means that if you are storing your backups next to other files, these might become subject to deletion too. When using this option make sure the backup files are stored in a directory used exclusively for such files, or to configure `BACKUP_PRUNING_PREFIX` to limit removal to certain files.
:::

| Setting | Description |
|----------|----------|
| **BACKUP_RETENTION_DAYS**    | Pass zero or a positive integer value to enable automatic rotation of old backups. The value declares the number of days for which a backup is kept. Default: `-1` |
| **BACKUP_PRUNING_LEEWAY**    | In case the duration a backup takes fluctuates noticeably in your setup you can adjust this setting to make sure there are no race conditions between the backup sit on the edge of the time window. Set this value to a duration finishing and the rotation not deleting backups that that is expected to be bigger than the maximum difference of backups. Valid values have a suffix of (s)econds, (m)inutes or (h)ours. By default, one minute is used. |
| **BACKUP_PRUNING_PREFIX**    | In case your target bucket or directory contains other files than the ones managed by this container, you can limit the scope of rotation by setting a prefix value. This would usually be the non-parametrized part of your BACKUP_FILENAME. E.g. if BACKUP_FILENAME is `db-backup-%Y-%m-%dT%H-%M-%S.tar.gz`, you can set BACKUP_PRUNING_PREFIX to `db-backup-` and make sure unrelated files are not affected by the rotation mechanism. |

## Source

By default, everything will be backed up. In case you need to backup only a specific data, set `BACKUP_SOURCES`.

Available options are:

| Setting | Description |
|----------|----------|
| **All**    |  Backups content of all docker volumes: website files, mysql/mariadb databases, minecraft, postgre databases, vhosts files. |
| **Files**    |  Backups only website files inside `/var/www/html/` directory. |
| **Emails**    |  Backups only email files inside `/var/mail/` directory. |
| **MySQL Databases**    |  Backups only MySQL/MariaDB databases. |
| **Postgres Databases**    |  Backups only PostgreSQL databases. |
| **MsSQL Databases**    |  Backups only MsSQL databases. |
| **VirtualHosts**    |  Backups only Apache/Nginx VirtualHosts for domains. |
| **Minecraft Data**    |  Backups only `/data` folder from Minecraft container. |


## Schedule

Backups can be run on fixed scheduled that are defined as a cron expression.
`BACKUP_CRON_EXPRESSION` is used to schedule the backup job, a cron expression represents a set of times, using 5 or 6 space-separated fields.

| Field name   | Mandatory? | Allowed values  | Allowed special characters |
| ----------   | ---------- | --------------  | -------------------------- |
| Seconds      | No         | 0-59            | `* / , -` |
| Minutes      | Yes        | 0-59            | `* / , -` |
| Hours        | Yes        | 0-23            | `* / , -` |
| Day of month | Yes        | 1-31            | `* / , - ?` |
| Month        | Yes        | 1-12 or JAN-DEC | `* / , -` |
| Day of week  | Yes        | 0-6 or SUN-SAT  | `* / , - ?` |


Month and Day-of-week field values are case insensitive. "SUN", "Sun", and "sun" are equally accepted. Refer to sites like [crontab.guru](https://crontab.guru) for help.

If no value is set, `@daily` will be used, which runs every day at midnight.

## Compression


| Setting | Description |
|----------|----------|
| `BACKUP_COMPRESSION`    |  The compression algorithm used in conjunction with tar. Valid options are: "gz" (Gzip), "zst" (Zstd) or "none" (tar only).  Default is "gz". Note that the selection affects the file extension. |
| `GZIP_PARALLELISM`    | Parallelism level for "gz" (Gzip) compression. Defines how many blocks of data are concurrently processed. Higher values result in faster compression. No effect on decompression. Default = `1`. Setting this to 0 will use all available threads. |

## Names

Options here are available byt disabled by defuatl and need to be manually enabled by the Administrator first.

| Setting | Description |
|----------|----------|
| `BACKUP_FILENAME`    |  The desired name of the backup file including the extension. Format verbs will be replaced as in `strftime`. Omitting all verbs will result in the same filename for every backup run, which means previous versions will be overwritten on subsequent runs. Extension can be defined literally or via `{{ .Extension }}` template,  in which case it will become either "tar.gz", "tar.zst" or ".tar" (depending on your `BACKUP_COMPRESSION` setting). The default `backup-%Y-%m-%dT%H-%M-%S.{{ .Extension }}` results in filenames like: `backup-2021-08-29T04-00-00.tar.gz`. |


## Exclude

`BACKUP_EXCLUDE_REGEXP` - When a value is given, all files in `BACKUP_SOURCES` whose full path matches the regular expression will be excluded from the archive. Regular Expressions can be used as from [the Go standard library](https://pkg.go.dev/regexp). Example: `\.log$`.

## Notifications

Notifications (email, Slack, etc.) can be sent out when a backup run finishes.

### NOTIFICATION_URLS

Configuration is provided as a comma-separated list of URLs as consumed by [shoutrrr](https://containrrr.dev/shoutrrr/v0.8/services/overview/)

| Service      | URL format                                                                                           |
|--------------|------------------------------------------------------------------------------------------------------|
| [Bark](#bark)         | `bark://devicekey@host`                                                                              |
| [Discord](#discord)      | `discord://token@id`                                                                                 |
| [Email](#email)        | `smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]`               |
| Gotify       | `gotify://gotify-host/token`                                                                         |
| Google Chat  | `googlechat://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz`                         |
| IFTTT        | `ifttt://key/?events=event1[,event2,...]&value1=value1&value2=value2&value3=value3`                  |
| Join         | `join://shoutrrr:api-key@join/?devices=device1[,device2,...][&icon=icon][&title=title]`              |
| Mattermost   | `mattermost://[username@]mattermost-host/token[/channel]`                                            |
| Matrix       | `matrix://username:password@host:port/[?rooms=!roomID1[,roomAlias2]]`                                |
| Ntfy         | `ntfy://username:password@ntfy.sh/topic`                                                             |
| OpsGenie     | `opsgenie://host/token?responders=responder1[,responder2]`                                           |
| Pushbullet   | `pushbullet://api-token[/device/#channel/email]`                                                    |
| Pushover     | `pushover://shoutrrr:apiToken@userKey/?devices=device1[,device2,...]`                                |
| Rocketchat   | `rocketchat://[username@]rocketchat-host/token[/channel|@recipient]`                                 |
| Slack        | `slack://[botname@]token-a/token-b/token-c`                                                          |
| Teams        | `teams://group@tenant/altId/groupOwner?host=organization.webhook.office.com`                         |
| Telegram     | `telegram://token@telegram?chats=@channel-1[,chat-id-1,...]`                                         |
| Zulip Chat   | `zulip://bot-mail:bot-key@zulip-domain/?stream=name-or-id&topic=name`                                |

#### Email

URL Format:

```
smtp://username:password@host:port/?from=fromAddress&to=recipient1[,recipient2,...]
```

URL Fields:

- **Username** – SMTP server username  
  **Default:** _empty_  
  **URL part:** `smtp://username:password@host:port/`

- **Password** – SMTP server password or hash (for OAuth2)  
  **Default:** _empty_  
  **URL part:** `smtp://username:password@host:port/`

- **Host** – SMTP server hostname or IP address (**Required**)  
  **URL part:** `smtp://username:password@host:port/`

- **Port** – SMTP server port  
  **Common values:** 25, 465, 587, 2525  
  **Default:** `25`  
  **URL part:** `smtp://username:password@host:port/`

Query/Param Props:

Props can be supplied via the `params` argument or directly in the URL using `?key=value&key=value`.

- **FromAddress** – Email address the mail is sent from (**Required**)  
  **Aliases:** `from`

- **ToAddresses** – Comma-separated list of recipient emails (**Required**)  
  **Aliases:** `to`

- **Auth** – SMTP authentication method  
  **Default:** `Unknown`  
  **Possible values:** `None`, `Plain`, `CRAMMD5`, `Unknown`, `OAuth2`

- **ClientHost** – Hostname sent to the SMTP server during the HELO phase  
  **Default:** `localhost`  
  Set to `"auto"` to use OS hostname

- **Encryption** – Encryption method  
  **Default:** `Auto`  
  **Possible values:** `None`, `ExplicitTLS`, `ImplicitTLS`, `Auto`

- **FromName** – Name of the sender  
  **Default:** _empty_

- **Subject** – Subject line of the email  
  **Default:** `Shoutrrr Notification`  
  **Aliases:** `title`

- **UseHTML** – Whether the message is in HTML format  
  **Default:** ❌ No

- **UseStartTLS** – Whether to use StartTLS encryption  
  **Default:** ✔ Yes  
  **Aliases:** `starttls`
  

#### Bark

URL Fields:

- **DeviceKey** – The key for each device (**Required**)  
  **URL part:** `bark://:devicekey@host/path`

- **Host** – Server hostname and port (**Required**)  
  **URL part:** `bark://:devicekey@host/path`

- **Path** – Server path  
  **Default:** `/`  
  **URL part:** `bark://:devicekey@host/path`

Query/Param Props:

Props can be either supplied using the `params` argument, or through the URL using  `?key=value&key=value` etc.

- **Badge** – The number displayed next to App icon  
  **Default:** `0`

- **Category** – Reserved field, no use yet  
  **Default:** _empty_

- **Copy** – The value to be copied  
  **Default:** _empty_

- **Group** – The group of the notification  
  **Default:** _empty_

- **Icon** – A URL to the icon, available only on iOS 15 or later  
  **Default:** _empty_

- **Scheme** – Server protocol, `http` or `https`  
  **Default:** `https`

- **Sound** – Value from [Bark Sounds](https://github.com/Finb/Bark/tree/master/Sounds)  
  **Default:** _empty_

- **Title** – Notification title, optionally set by the sender  
  **Default:** _empty_

- **URL** – URL that will open when notification is clicked  
  **Default:** _empty_

#### Discord

URL Format:

```
discord://token@webhookid
```

URL Fields:

- **Token** – (**Required**)  
  **URL part:** `discord://token@webhookid/`

- **WebhookID** – (**Required**)  
  **URL part:** `discord://token@webhookid/`

Query/Param Props:

Props can be either supplied using the `params` argument, or through the URL using  `?key=value&key=value` etc.

- **Avatar** – Override the webhook default avatar with specified URL  
  **Default:** _empty_  
  **Aliases:** `avatarurl`

- **Color** – The color of the left border for plain messages  
  **Default:** `0x50D9ff`

- **ColorDebug** – The color of the left border for debug messages  
  **Default:** `0x7b00ab`

- **ColorError** – The color of the left border for error messages  
  **Default:** `0xd60510`

- **ColorInfo** – The color of the left border for info messages  
  **Default:** `0x2488ff`

- **ColorWarn** – The color of the left border for warning messages  
  **Default:** `0xffc441`

- **JSON** – Whether to send the whole message as the JSON payload instead of using it as the `content` field  
  **Default:** ❌ No

- **SplitLines** – Whether to send each line as a separate embedded item  
  **Default:** ✔ Yes

- **Title**  
  **Default:** _empty_

- **Username** – Override the webhook default username  
  **Default:** _empty_

Creating a Webhook in Discord:

1. Open your channel settings by clicking the gear icon next to the channel name.
   ![screenshot 1](https://i.postimg.cc/FmYrJfhZ/sc-1.png)
2. In the left menu, click **Integrations**.  
   ![screenshot 2](https://i.postimg.cc/NtJGM8vF/sc-2.png)
3. On the right, click **Create Webhook**.  
   ![screenshot 3](https://i.postimg.cc/7Djx1x6y/sc-3.png)
4. Set the name, channel, and icon to your preference, then click **Copy Webhook URL**.  
   ![screenshot 4](https://i.postimg.cc/2rvC2PbL/sc-4.png)
5. Press **Save Changes**.
   ![screenshot 5](https://i.postimg.cc/wqG9GzQL/sc-5.png)
6 Format the service URL:
   ```
   https://discord.com/api/webhooks/693853386302554172/W3dE2OZz4C13_4z_uHfDOoC7BqTW288s-z1ykqI0iJnY_HjRqMGO8Sc7YDqvf_KVKjhJ
                                    └────────────────┘ └──────────────────────────────────────────────────────────────────┘
                                        webhook id                                    token
   
   discord://W3dE2OZz4C13_4z_uHfDOoC7BqTW288s-z1ykqI0iJnY_HjRqMGO8Sc7YDqvf_KVKjhJ@693853386302554172
             └──────────────────────────────────────────────────────────────────┘ └────────────────┘
                                             token                                    webhook id
   ```



### NOTIFICATION_LEVEL

By default, notifications would only be sent out when a backup run fails. To receive notifications for every run, set `NOTIFICATION_LEVEL` to `info` instead of the default `error`.


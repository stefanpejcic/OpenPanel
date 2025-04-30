---
sidebar_position: 6
---

# Backups

![backups.png](/img/panel/v2/backups.png)

## Introduction to Backups

In OpenPanel, backups are configurable directly **by the panel users**, unlike other panels where backups are managed by the admin.

This empowers users to define their own backup schedules, choose exactly what to back up, and select from a wide range of supported destinations.
As a result, administrators have fewer tasks to manage, and users gain greater control and flexibility.

## Destinations

OpenPanel supports the following destinations:

- **S3 storage** - AWS S3, Filebase, MinIO, etc.
- **WebDAV** - Remote backups using WebDAV.
- **SSH** - Remote backups via SSH to another server.
- **Azure** - Remote backups to Azure Blob Storage.
- **Dropbox** - Remote backups to Dropbox cloud storage.

User can select destination from **Backups > Destinations** page.

Once backup is selected, it's options are available on **Backups > Settings page**.

### SSH

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


### Azure


### Dropbox

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

Important: If you chose `App folder` access during the creation of your Dropbox app in step 1 above, `DROPBOX_REMOTE_PATH` will be a relative path under the App folder! (For example, DROPBOX_REMOTE_PATH=/somedir means the backup file will be uploaded to /Apps/myapp/somedir) On the other hand if you chose Full Dropbox access, the value for `DROPBOX_REMOTE_PATH` will represent an absolute path inside your Dropbox storage area. (Still considering the same example above, the backup file will be uploaded to /somedir in your Dropbox root)





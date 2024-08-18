# Disable PHP errors or increase upload limits per website

If you have multiple websites in the same domain folder and only want to enable or disable php settings for a certain website, create a file `.user.ini` inside the website folder.

In the file set the desired php values, for example:

set post and upload max file size:

```bash
upload_max_filesize = 200M
post_max_size = 200M
```

display all php errors:

```bash
error_reporting = E_ALL
display_errors = on
```

After editingq restart the php service. Navigate to PHP.INI editor and select the php version that your domain is using. Simply click on the save button and the service will be restarted, immediately applying settings from the .user.ini file.

# How to Install Older Versions of Ioncube Loader Extensions for PHP

Administrators can configure a custom link to a `.tar.gz` archive containing older Ioncube Loader versions. By default, the file does not exist, and the `opencli php-ioncube` script will download the latest versions from the Ioncube Loader website.

Follow these steps to set up an older version of Ioncube Loader:

1. **Create a directory named `ioncube`.**

2. **Place the loader files in the directory.**  
   Inside the `ioncube` directory, add the loader files for each PHP version in the format: `ioncube_loader_lin_*.so`

3. **Compress the directory into a `.tar.gz` archive.**

4. **Upload the archive online.**  
Upload the `.tar.gz` archive to a web-accessible location and copy the download link.

5. **Set the download link.**  
Add the link to the `/etc/openpanel/php/ioncube.txt` file.

Once the link is set, all subsequent Ioncube Loader installations using the [opencli php-ioncube](https://dev.openpanel.com/cli/php.html#Enable-ioncube-loader) command will download the loader from the specified archive.

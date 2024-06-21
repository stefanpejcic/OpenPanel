---
sidebar_position: 6
---

# Log Viewer

The *OpenAdmin > Services > View Log Files* feature enables users to access and monitor logs for both OpenPanel and system services.

![log viewer page](https://i.postimg.cc/zGmWT8L0/errorlof.png)

## How to View Log Files

Navigate to Services > View Log Files

Select the log file you would like to view and optionally number of lines from the file.

After selecting a log file, two new buttons appear under the log content: 
- *Delete* - will empty the file contents
- *Download* - will download the entire log file to your browser.


## How to add more files to OpenAdmin Log Viewer

This functionality supports modularity by allowing customization of the log files displayed in the viewer.


List of default log files: https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/main/openadmin/config/log_paths.json

To define custom log files for the viewer:

1. Edit Configuration File:
   Modify the file located at `/etc/openpanel/openadmin/config/log_paths.json`. This file should contain entries in JSON format where each key-value pair represents a log file name and its corresponding path.
   
   Here is an example of what the log_paths.json file might look like:
   Simply edit the file `/etc/openpanel/openadmin/config/log_paths.json` and in it set the desired log files and names:
   ```json
   {
     "Nginx Access Log": "/var/log/nginx/access.log",
      "Nginx Error Log": "/var/log/nginx/access.log",
      "OpenAdmin Access Log": "/var/log/openpanel/admin/access.log",
      "OpenAdmin Error Log": "/var/log/openpanel/admin/error.log",
      "OpenAdmin API Log": "/var/log/openpanel/admin/api.log",
      "Custom Service Log": "/path/to/custom/service.log"
      "Syslog": "/var/log/syslog"
   }
   ```
   Replace `/path/to/custom/service.log` with actual path to your custom log files.

2. Verify JSON Validity:
   Ensure that the log_paths.json file is formatted correctly as JSON. Any syntax errors in the JSON file will prevent the custom log files from appearing in the viewer.
   You can check the validity of your JSON file by using a command-line JSON processor like jq:
   ```bash
   cat /etc/openpanel/openadmin/config/log_paths.json | jq
   ```
   If the JSON is valid, `jq` will output the parsed JSON structure. If there are any errors, `jq` will indicate where the problem lies.

3. View Custom Logs in Viewer:
   After saving the changes, navigate to *OpenAdmin > Services > View Log Files* in the interface. The custom log files you specified in log_paths.json should now appear alongside the default logs.
   
By following these steps, you can effectively customize the log files displayed in the OpenAdmin log viewer according to your specific requirements. This flexibility allows you to monitor logs from both standard services and any custom applications or services you integrate with OpenPanel.



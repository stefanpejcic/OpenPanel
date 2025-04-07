---
sidebar_position: 4
---

# Disk Usage

![disk_usage.png](/img/panel/v2/disk_usage-1efbc5d2ebaeba60e4d2e1d8d903a36a.png)

## Introduction to Disk Usage

Disk usage refers to the amount of space occupied by the content of your websites, including databases, files, videos, images, emails, and web pages. Monitoring disk usage is crucial to ensure efficient resource management and to prevent storage-related issues.

## Disk Usage Chart

The Disk Usage Chart provides a visual representation of disk usage per directory. It allows you to quickly assess which directories are consuming the most storage space.

## Browsing Directories

The "Browse Directories" section displays a table listing directories and their respective sizes. You can navigate through directories and view their sizes.

To navigate to a specific directory, click on the directory name in the table.

To return to the parent directory, click on "Up One Level."


## Usage Example
Suppose you want to check the disk usage of a specific directory. Follow these steps:

1. Navigate to the "Browse Directories" section.
2. Click on the directory you want to explore.
3. View the directory's size and content.

## Lowering Disk Usage

Managing disk usage is essential to maintain the performance and stability of your hosting environment. Here are some tips and best practices for reducing disk usage within OpenPanel:

### Regularly Clean Up Unused Files
- Remove Temporary Files: Periodically delete temporary files, such as log files, cache files, and old backups. These files can accumulate and consume valuable disk space over time.
- Empty Trash: Ensure that you regularly empty the trash or recycle bin to permanently remove files that are no longer needed.
- Unused Website Assets: Identify and remove unused images, videos, or other assets from your websites. Hosting unnecessary files can quickly fill up your disk space.

### Optimize Database Usage
- Database Cleanup: Regularly optimize and clean up your databases by removing outdated or unnecessary data. This can help reduce the database's size and improve its performance.
- Use Database Compression: If your hosting service supports it, enable database compression to reduce the storage space required by your databases.

### Monitor and Limit Email Attachments
- Attachments: Encourage users to be mindful of email attachments. Large attachments can quickly consume disk space. Consider setting attachment size limits for email accounts to prevent oversized files from being stored.

### Manage Log Files
- Log Rotation: Implement log rotation to prevent log files from growing indefinitely. Older log files can be archived or deleted to free up space.
- Log Level Adjustment: Adjust the verbosity or log level of your applications to generate fewer log entries, which can help reduce disk usage.

### Offload Large Media
- External File Hosting: Consider offloading large media files, such as videos or high-resolution images, to external hosting services or Content Delivery Networks (CDNs). This reduces the load on your server and conserves disk space.

## Troubleshooting
If you encounter any issues while using the Disk Usage feature, consider the following steps:

1. Ensure that you have appropriate permissions to access the directories.
2. Verify that the disk usage data source is correctly configured.
3. Clear your browser cache if you experience any display issues with the chart or table.

---

By following these guidelines and regularly managing your disk usage, you can maintain a healthy hosting environment and avoid unexpected issues related to storage limitations.

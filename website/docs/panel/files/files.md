---
sidebar_position: 1
---

# Files

The File Manager interface allows you to manage files in your website directory `/var/www/html/`

![filemanager.png](/img/panel/v1/files/filemanager.png)


## Create File

To create a new file, click on the 'New File' button in the menu. In the modal, set the filename, and then click on 'Create File'.

![filemanager_new_file.png](/img/panel/v1/files/filemanager_new_file.png)

## Create Folder

To create a new folder, click on the 'New Folder' button in the menu. In the modal, set the folder name, and then click on 'Create Folder'.

![filemanager_new_folder.png](/img/panel/v1/files/filemanager_new_folder.png)


## Upload Files

To upload a new file, click on the 'Upload' button. In the modal, select the files to be uploaded, and then click on the 'Upload' button.

![filemanager_upload.png](/img/panel/v1/files/filemanager_upload.png)

Except for the account's quota limits, the File Manager doesn't impose any restrictions on upload sizes. This means that you can upload larger files if needed.

Nevertheless, even though it's possible to upload files 10GB+ in size, we advise you to carry out such uploads using a protocol more suitable for larger files, such as FTP or SFTP. These protocols offer enhanced reliability and performance.

File Manager allows you to upload multiple files at once.

![filemanager_upload_multiple.png](/img/panel/v1/files/filemanager_upload_multiple.png)

## Select all

Use your mouse cursor to select multiple files and folders. Click and drag, then click again to release the selection.

To select multiple files, one by one, click and hold `Ctrl` key while clicking on the rows.

To select all files in the directory at once, click on the 'Select All' button. To deselect all files, click on the 'Deselect' button.

![filemanager_selectall.png](/img/panel/v1/files/filemanager_selectall.png)

## Delete

To delete files or folders click on the 'Delete' button. If multiple files or folders are selected, you will see the list in the modal and click 'Delete' to permanently delete the selected files.

![filemanager_delete.png](/img/panel/v1/files/filemanager_delete.png)

:::danger
OpenPanel FileManager does not utilize a separate Trash folder where files are temporarily moved and can later be restored from, similar to the Recycle Bin on your computer system. If you delete files using the delete option, they are permanently deleted immediately and cannot be recovered. Please exercise caution when using the delete function.
:::

## Download File

To download a file, double-click the file name or click on the 'Download' button when the file is selected.

![filemanager_download_file.png](/img/panel/v1/files/filemanager_download_file.png)
:::info
Downloading multiple files at once is not supported. Instead, the 'Compress' option should be used to create an archive of files, and then download that archive.
:::


## View File

To view the content of a file that can be edited in a text editor or opened as an image, click on the file, then on the 'View' button in the menu. A new modal will be displayed with the content of the file in plain text or the base64 encoded image.

![filemanager_view_file.png](/img/panel/v1/files/filemanager_view_file.png)

## Edit File

To edit the content of a file that can be edited in a text editor, click on the file, then on the 'Edit' button in the menu. A new page will open with a text editor where you can edit the file content.

![filemanager_edit_file.png](/img/panel/v1/files/filemanager_edit_file.png)

## Rename File

To rename a file or folder, click on it, then click on the 'Rename' button, and set the new name.

![filemanager_rename.png](/img/panel/v1/files/filemanager_rename.png)

## Copy Files

To copy files from one folder to another, first select the desired files and click on the 'Copy' button. The new modal will display a list of all selected file names and allow you to set the destination name of the folder where the files will be copied.

![filemanager_copy.png](/img/panel/v1/files/filemanager_copy.png)

To initiate the copying process, click on the 'Copy' button in the modal. A progress bar will appear, indicating the progress made, and a 'Copy complete' message will be displayed when the process is finished.

![filemanager_copy_complete.png](/img/panel/v1/files/filemanager_copy_complete.png)


## Move Files

To move files from one folder to another, first select the desired files and click on the 'Move' button. The new modal will display a list of all selected file names and allow you to set the destination name of the folder where the files will be moved to.

![filemanager_move.png](/img/panel/v1/files/filemanager_move.png)

To initiate the move process, click on the 'MOve' button in the modal. A progress bar will appear, indicating the progress made, and a 'Complete' message will be displayed when the process is finished.

![filemanager_move_progress.png](/img/panel/v1/files/filemanager_move_progress.png)


## Extract Archive

To extract an uploaded archive (`.zip`, `.tar`, or `.tar.gz` file), select the file, and then click on the 'Extract' button in the menu.

In the modal, set the name of the folder where files will be extracted, and click on 'Confirm' to start the process.

![filemanager_extract.png](/img/panel/v1/files/filemanager_extract.png)


## Compress to Archive

To create an archive of files, first, select the desired files or folders, and then click on the 'Compress' button. The new modal will display a list of all selected file names and allow you to set the archive name and extension (`.zip`, `.tar`, or `.tar.gz`).

![filemanager_compress.png](/img/panel/v1/files/filemanager_compress.png)


## Empty Folder

If a folder is empty, you will see the 'No items found.' message and the menu with file options will be hidden. Only the options to create a new file, folder, or upload files will be available.

![filemanager_nofiles.png](/img/panel/v1/files/filemanager_nofiles.png)

## Search Files and Folders

The Search button across the OpenPanel interface is used to search for options or websites, but On the FileManager interface, it has a special feature allowing you to also search for your files and folder names.

To activate the search field, click on the Search icon in the top right corner.

![filemanager_search_results.png](/img/panel/v1/files/filemanager_search_results.png)

For performance reasons, search results are limited to a maximum of 10 results for files and 10 results for folders.

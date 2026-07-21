---
sidebar_position: 3
---

# Download from URL

The **Download from URL** feature allows you to fetch files directly from the internet into your selected directory.

To use:

1. Navigate to the desired directory.
2. Click the **Upload** button.
3. Select **Download from URL instead**.
4. Enter the direct download link (URL) to the file.
5. Click **Download** to begin.

The file will be downloaded using `wget` and saved with the same name as on the remote server.

If a file with the same name already exists in the directory, it will not be overwritten—instead, a version with a `+1` suffix will be created.

> ✅ This is especially useful for large files or when direct device uploads are slow or limited.

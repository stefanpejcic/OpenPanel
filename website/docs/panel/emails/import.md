---
sidebar_position: 3
---

# Address Importer

The **Address Importer** page allows you to create multiple email accounts at once by uploading an `.xls` or `.csv` file.

### How to Import Email Addresses

1. Navigate to **OpenPanel > Emails > Address Importer**.
2. Select the `.xls` or `.csv` file from your device and click **Upload**.

### File Format Requirements

* Supported delimiters: `,` (comma) or `;` (semicolon)
* Required columns: `username`, `password`
* Optional column: `quota`

  * If included, the quota should specify a unit: `K`, `M`, `G`, or `T`.
  * If no unit is specified, the default is **MB**.

### Example File

```csv
email,password,quota
alice@example.com,pass123,500M
bob@mydomain.com,s3cr3t,1G
charlie@other.com,charliepwd,
diana@example.com,d1@n@pwd,0
eve@unauthorized.com,evepass,100T
new@example.com,ssaaasa2,10M
```

### Review & Confirmation

After uploading the file:

* You’ll see a preview of the email accounts to be created, along with their passwords and quota (if provided).
* The system will flag entries if they will be skipped during the import:

  * Email addresses that already exist
  * Domains that are not associated with your account

### Final Step

Click **Start Upload** to begin the import process. A log will appear below the table to show the progress and status of the operation.

Once completed, all valid email accounts will be successfully imported.

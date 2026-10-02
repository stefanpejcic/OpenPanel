---
sidebar_position: 6
---

# Fix Permissions
The **Fix Permissions** tool is designed to automatically correct file and folder permissions for your websites, ensuring that they are secure and function properly.

![Fix Permissions page with the directory field and the Fix Permissions button](/img/openpanel-screenshots/files/fix_permissions-form.png#gh-light-mode-only)
![Fix Permissions page with the directory field and the Fix Permissions button](/img/openpanel-screenshots/files/fix_permissions-form_dark.png#gh-dark-mode-only)

Open **Fix Permissions**  page and set the path or leave empty (`/var/www/html/`) to go over all files.

## How It Works
Clicking the **Fix Permissions** button will:

- Set you as the **owner** of all your website files and folders inside `/var/www/html/`.
- Set **files** to `664` so you can edit them and the web server can read them.
- Set **folders** to `775` so you can manage them and the web server can open them.
- Check and fix ownership of **FTP account** folders so FTP users can upload and edit their files.

It will **not** change emails, databases, or any file contents.

## Caution

- Custom permissions or advanced configurations may be overridden.
- If you're running custom scripts that require different permissions (e.g., executable `.sh` files), you'll need to reapply those settings manually after using this tool.

## Recommended Usage

You should use the **Fix Permissions** option only when:

- Your website is showing permission-related errors.
- You've uploaded or modified files manually (e.g., via FTP or File Manager).
- Your CMS (WordPress, Joomla, etc.) cannot write to certain folders.
- You suspect ownership or permission issues after a migration or manual edit.

Use this tool only when you’re experiencing issues. Frequent, unnecessary use is not required and may revert intentional permission changes.

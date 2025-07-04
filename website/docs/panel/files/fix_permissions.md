---
sidebar_position: 8
---

# Fix Permissions
The **Fix Permissions** tool is designed to automatically correct file and folder permissions for your websites, ensuring that they are secure and function properly.

Open **Fix Permissions**  page and set the path or leave empty (`/var/www/html/`) to go over all files.

## How It Works
Clicking the **Fix Permissions** button will:

- Recursively set the correct **ownership** for files and folders.
- Apply standard **permissions**:
  - Files: `644`
  - Folders: `755`
- Fix common permission problems for known CMS directories like `wp-content`, `storage`, `cache`, etc.

## When to use it
You should use the **Fix Permissions** option only when:

- Your website is showing permission-related errors.
- You've uploaded or modified files manually (e.g., via FTP or File Manager).
- Your CMS (WordPress, Joomla, etc.) cannot write to certain folders.
- You suspect ownership or permission issues after a migration or manual edit.

## Caution

- Custom permissions or advanced configurations may be overridden.
- If you're running custom scripts that require different permissions (e.g., executable `.sh` files), you'll need to reapply those settings manually after using this tool.

## Recommended Usage

Use this tool only when you’re experiencing issues. Frequent, unnecessary use is not required and may revert intentional permission changes.

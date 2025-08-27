# 403 Error Troubleshooting Guide

This guide covers the most common causes of **403 Forbidden** errors on a website and how to resolve them.

---

## 1. Coraza WAF

One of the most common causes is the **Coraza WAF** blocking access according to **ModSecurity CoreRuleSet** rules.

1. Navigate to **Advanced > WAF** in your hosting control panel.
   ![WAF disable](https://i.postimg.cc/fZw2Skqv/waf-status.png)

2. Check if WAF is enabled for the domain. Temporarily disable it, then test the website to confirm if WAF is the cause.

3. If WAF is the issue:

   * **Not recommended:** Leave WAF disabled for the domain.
   * **Better option:** Identify the exact rule being triggered and disable only that rule.

4. To find the rule:

   * Click **View Logs** to open ModSecurity logs.
   * Search for your 403 request by IP, time, or request details.
   * Note the **Rule ID**.

5. Go to **Manage Rules** for the domain in the WAF section:

   * Disable by Rule ID: [SecRuleRemoveByID](https://coraza.io/docs/seclang/directives/#secruleremovebyid)
   * Disable by Tag: [SecRuleRemoveByTag](https://coraza.io/docs/seclang/directives/#secruleremovebytag)

   ![WAF edit rules](https://i.postimg.cc/GcSm9Xzm/2025-08-13-11-58.png)

6. Re-enable WAF and retest the website.

---

## 2. WordPress

Content management systems like **WordPress** use `.htaccess` files that may block access.

1. Temporarily rename or remove the `.htaccess` file in your website’s root directory.
2. Create a simple test file (e.g., `index.php`) and open it in the browser.
3. If it loads, the issue is CMS-related.
4. Possible fixes:

   * Disable WordPress plugins temporarily.
   * Restore the default `.htaccess` file.
   * Seek help on your CMS forums or [OpenPanel Community](https://community.openpanel.org/).

---

## 3. File Permissions

Incorrect file ownership or permissions can also trigger 403 errors.

1. Open **File Manager** and navigate to the domain’s directory.

2. Click **Options** and enable **Owner** and **Group** columns.
   ![File manager owner](https://i.postimg.cc/cZ4FdrY4/2025-08-13-11-52.png)

3. Ensure all files have the same owner and group.

4. Check file and folder permissions.

5. If ownership is incorrect, use **Files > Fix Permissions** to restore defaults.

---

## 4. Nginx or Apache Restrictions

By default, **Nginx** and **Apache** block access to sensitive files:

```
.git
composer.json
composer.lock
auth.json
config.php
wp-config.php
vendor
```

If you are sure you want these files accessible (not recommended), edit the domain’s VHost configuration.

### For OpenResty / Nginx:

```nginx
# <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
location ~* ^/(\.git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor) {
    deny all;
    return 403;
}
# <!-- END EXPOSED RESOURCES PROTECTION -->
```

### For Apache:

```apache
# <!-- BEGIN EXPOSED RESOURCES PROTECTION -->
<Directory <DOCUMENT_ROOT>>
    <FilesMatch "\.(git|composer\.(json|lock)|auth\.json|config\.php|wp-config\.php|vendor)">
        Require all denied
    </FilesMatch>
</Directory>
# <!-- END EXPOSED RESOURCES PROTECTION -->
```


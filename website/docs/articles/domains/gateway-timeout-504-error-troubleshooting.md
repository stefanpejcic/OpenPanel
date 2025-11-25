# 504 Error Troubleshooting Guide

This guide explains the most common causes of **504 Gateway Timeout** errors and how to fix them.

---

## 1. Restart PHP-FPM 

A 502 Bad Gateway error is common when using Nginx. It means the web server (Nginx) did not receive a response in time from the backend (such as PHP-FPM, Node.js, or Python).

**Steps to check PHP-FPM:**

1. Verify that the PHP service is running.
   If Docker feature is enabled, go to **Docker > Containers** and check the status of the PHP container.
   If inactive, start it.
   ![screenshot](https://i.postimg.cc/wx8Dm4XP/image.png)
   If active, stop and start it again. This will terminate any background processes that might be stuck.

---

## 2. Wait

If you are not the website owner and have no access to OpenPanel account, the only thing you can do is notify the owner for this issue and wait till website starts responding again.

---

## 3. Update WHMCS

A 504 error can sometimes occur in WHMCS if it hasn’t been updated regularly. The solution is to update WHMCS to the latest version.
[Guide on updating WHMCS](https://docs.whmcs.com/8-13/system/updates/updating-whmcs/)


With these steps, you should be able to identify and fix the most common causes of **504 Gateway Timeout** errors.

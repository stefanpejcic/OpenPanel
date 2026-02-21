# Domain Shows Default Page

When you add a domain to **OpenPanel**, it will initially display the following message:

> **Ready, set, internet 🎉**
> This domain currently has no website. Please check back later.

This default page will appear until an `index.php` or `index.html` file is present in the domain’s document root.

![Default page](https://i.postimg.cc/Zn8gbHm6/2025-08-13-12-20.png)

**Solution:**
Upload your website files (including an `index.php` or `index.html`) to the document root. Once uploaded, your website will replace the default page.

---

## Caching Issue

If **Varnish caching** is enabled and you access the domain before it has content, the default page may get cached.
This means that even after you upload your site, the cached default page could still be shown.

**How to fix:**

* Temporarily disable and then re-enable Varnish caching for the domain, or
* Completely disable and re-enable Varnish to clear all cached content.

**Disable Varnish for a specific domain:**

1. Go to **OpenPanel > Cache > Varnish**
2. Disable Varnish for that domain

![Disable Varnish cache for domain](https://i.postimg.cc/dwSGj2qk/2025-08-13-12-25.png)

**Purge all Varnish cache:**
Simply click **Disable**, then **Enable** on the same page.

---

## Customize the Default Page

Administrators can change the default page shown on domains without content.

To customize:

1. In **OpenAdmin**, go to **Domains > Edit Domain Templates**
2. Modify the HTML code for the default page

![Edit default page in OpenAdmin](https://i.postimg.cc/JRx0Qm3T/2025-08-13-12-28.png)


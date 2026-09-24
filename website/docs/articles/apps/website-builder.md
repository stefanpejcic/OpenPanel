---
sidebar_label: "Website Builder"
description: "How to build a website without coding using the drag-and-drop Website Builder in OpenPanel: create a site, edit blocks, publish static HTML with Tailwind CSS, and back it up."
---

# How to Build a Website Without Coding (Drag-and-Drop Website Builder)

OpenPanel includes a drag-and-drop **Website Builder** based on [GrapesJS](https://grapesjs.com/). You can design a page visually and publish it as a fast, static HTML website styled with [Tailwind CSS](https://tailwindcss.com/) - no WordPress, database or coding needed.

It's a good fit for landing pages, "coming soon" pages, portfolios and simple business sites.

---

## Create a Website

1. Add the domain in **OpenPanel → Domains** ([Add a new domain](/docs/panel/domains/new/)).
2. Go to **OpenPanel → Websites → Install App** and, under **Website Builder**, click **Create website**.
3. Choose the **domain** and, optionally, a **subdirectory**.
4. Click **Open Website Builder**.

![Website Builder page with the domain and folder to create the website in](/img/openpanel-screenshots/applications/builder-form.png#gh-light-mode-only)
![Website Builder page with the domain and folder to create the website in](/img/openpanel-screenshots/applications/builder-form_dark.png#gh-dark-mode-only)

---

## Edit the Page

- Click the **blocks** icon in the top-right corner to browse sections (headers, text, images, buttons, columns) and drag them onto the page.
- Click any element to edit its text, links and styles in the panel on the right.
- Use the device icons to check how the page looks on **desktop, tablet and mobile**.
- **Save** to publish. The site is written as `index.html` and `style.css` in the domain folder and is live immediately.

To edit later, open **OpenPanel → Websites → Sites** and click **Open Editor** next to the site.

For more editing features, see the [GrapesJS documentation](https://grapesjs.com/docs/).

---

## Backups

On the site's page in **Sites**, open the **Backups** tab:

- **Generate Backup** saves the current website files.
- **Display Backups** → pick a date → **Start Restore** brings back an earlier version.

---

## Removing the Site

In the **Remove** tab, **Detach website** stops managing it in the builder but keeps the files; **Delete website** removes the files.

More: [Website Builder in OpenPanel](/docs/panel/applications/builder/).

---

## Related

- [Host a static HTML website](/docs/articles/websites/hosting-a-static-website-with-openpanel/)
- [Install WordPress](/docs/articles/websites/how-to-install-wordpress-with-openpanel/)

---
sidebar_position: 1
---

# Dashboard

The OpenPanel interface is the central hub of your user account, providing access to all available tools, usage information, and notifications.

The dashboard has two views, switched with the **Overview** and **Tools** buttons at the top of the page:

- **Overview** (the default): plan usage, your websites, resource usage, quick actions and the 2FA, Information and How-to widgets.
- **Tools**: every feature available on your plan, grouped in sections.

The panel remembers the view you used last.

When your hosting provider offers an upgrade plan, the dashboard also has an [**Upgrade now**](/docs/panel/dashboard/upgrade/) tab that compares your plan with the upgrade plan.

![OpenPanel dashboard Overview with plan usage, websites, usage and quick actions](/img/openpanel-screenshots/dashboard/dashboard-window.png#gh-light-mode-only)
![OpenPanel dashboard Overview with plan usage, websites, usage and quick actions](/img/openpanel-screenshots/dashboard/dashboard-window_dark.png#gh-dark-mode-only)

## Overview

### Plan usage

A row of tiles shows how many websites, domains, databases, email accounts and FTP accounts you use out of your plan's limit, or **Unlimited** when there is no limit. Only items enabled on your plan are shown.

Next to the tiles, **New website** and **Add domain** open the forms to add them. When you have used all websites or domains on your plan, the button is disabled and hovering over it tells you so.

### Websites

Lists your newest websites with what they run on (for example WordPress or a Node.js app) and their status: **Active**, **Running**, **Stopped** or **Suspended**. Click a site's name to manage it, or open the **Actions** menu on its row for **Manage**, **Open site**, **File manager**, **DNS zone** and **SSL**. **View all** opens the [Sites](/docs/panel/applications/sites/) page.

If you have no websites yet, the section has a **New website** button instead.

### Usage

Shows your storage, inodes, CPU and memory usage against your plan's limits. CPU and memory are read live from the server. Click any row, or **View details**, for the [Resource usage](/docs/panel/statistics/resource_usage/) page.

### Quick actions

Shortcuts to the tools you are most likely to need, such as File manager, phpMyAdmin, DNS zone editor, SSL certificates, Cron jobs, Backups and Webmail. Only tools enabled on your plan are shown. **All tools** switches to the Tools view.

## Tools

All features available on your plan, grouped in sections like Websites, Domains, Email, Files and Databases. Each tool is an icon with its name below it.

Type in **Find a tool** at the top to filter the icons, for example *DNS*, *PHP version* or *backups*. Press **/** anywhere on the Tools view to jump to it.

### Customize the layout

You can rearrange the sections to suit how you work:

- **Reorder**: drag a section by its title to move it.
- **Collapse**: click the arrow on the right of a section's title to hide or show its icons.

Below the sections, next to **Layout**:

- **Lock layout**: stops sections from being dragged, so you can't move them by accident, for example while scrolling on a phone. Click **Unlock layout** to allow it again.
- **Reset layout**: puts every section back in its original order and expands any collapsed sections.

![Tools view with the Layout row, showing the lock and reset buttons](/img/openpanel-screenshots/dashboard/dashboard-layout.png#gh-light-mode-only)
![Tools view with the Layout row, showing the lock and reset buttons](/img/openpanel-screenshots/dashboard/dashboard-layout_dark.png#gh-dark-mode-only)

The section order, collapsed sections and lock are saved in your browser, so they apply only on that browser and are cleared if you clear its site data.

## Two-Factor Authentication

This widget shows whether [Two-Factor Authentication](/docs/panel/security/2fa/) is enabled for your account. If it isn't, click **Enable 2FA** to set it up.

![Two-Factor Authentication widget showing 2FA as disabled with the Enable 2FA button](/img/openpanel-screenshots/dashboard/dashboard-twofa.png#gh-light-mode-only)
![Two-Factor Authentication widget showing 2FA as disabled with the Enable 2FA button](/img/openpanel-screenshots/dashboard/dashboard-twofa_dark.png#gh-dark-mode-only)

## Information

This section displays general server information:

- Username
- Plan name
- Shared or Dedicated IP address
- Last login IP address
- Nameservers (if configured)

![Information widget with the username, plan, IP address and last login IP address](/img/openpanel-screenshots/dashboard/dashboard-information.png#gh-light-mode-only)
![Information widget with the username, plan, IP address and last login IP address](/img/openpanel-screenshots/dashboard/dashboard-information_dark.png#gh-dark-mode-only)

## How-to Guides

If enabled by the Administrator, the How-to guides section will display links to Knowledge Base articles configured by your hosting provider, or to the official OpenPanel documentation by default.

![General How-to widget with links to knowledge base articles](/img/openpanel-screenshots/dashboard/dashboard-howto.png#gh-light-mode-only)
![General How-to widget with links to knowledge base articles](/img/openpanel-screenshots/dashboard/dashboard-howto_dark.png#gh-dark-mode-only)

## Favorites

When [Favorites feature](/docs/admin/settings/modules/#favorites) is enabled by Administrator, users can bookmark up to 10 pages from the OpenPanel interface. These pages will appear in the sidebar menu. To add a page to favorites, simply left-click the star icon in the top-right corner of any page. To remove a page, right-click the same icon.

![Star icon in the top-right corner of the page header, used to add the current page to favorites](/img/openpanel-screenshots/dashboard/dashboard-favorites.png#gh-light-mode-only)
![Star icon in the top-right corner of the page header, used to add the current page to favorites](/img/openpanel-screenshots/dashboard/dashboard-favorites_dark.png#gh-dark-mode-only)

Saved pages are listed under **Favorites** at the top of the sidebar:

![Favorites list at the top of the sidebar with a saved page](/img/openpanel-screenshots/dashboard/dashboard-favorites-sidebar.png#gh-light-mode-only)
![Favorites list at the top of the sidebar with a saved page](/img/openpanel-screenshots/dashboard/dashboard-favorites-sidebar_dark.png#gh-dark-mode-only)

## Services

A service badge with the service name and its status is shown in the top-right corner of the header, but only on pages that manage a container, such as MySQL, Redis or Cron Jobs. Other pages don't show it.

Hover over the service name to see live information about its container:

- Name
- CPU Usage
- RAM Usage
- Memory %
- Network I/O
- Block I/O
- PIDs

If the Services feature is allowed on your plan, a **Manage** link next to the badge opens the service's page. If only the Docker feature is allowed, **Manage** opens the service on the [Containers](/docs/panel/containers/) page instead.

![Service badge in the page header with its tooltip showing the container name and resource usage](/img/openpanel-screenshots/dashboard/dashboard-service.png#gh-light-mode-only)
![Service badge in the page header with its tooltip showing the container name and resource usage](/img/openpanel-screenshots/dashboard/dashboard-service_dark.png#gh-dark-mode-only)

## Search

Click the magnifying glass icon in the top-right corner of the header to open the search box. It searches across pages/features, files and folders, websites, domains, email accounts, FTP accounts, containers, services, cron jobs, and MySQL/PostgreSQL databases and users. Results are limited to what's enabled on your plan — some entity types (databases, domains, emails, FTP, containers, services, websites, cron jobs) additionally require an Enterprise license.

![Search box opened from the magnifying glass icon in the header, with results for mysql](/img/openpanel-screenshots/dashboard/dashboard-search.png#gh-light-mode-only)
![Search box opened from the magnifying glass icon in the header, with results for mysql](/img/openpanel-screenshots/dashboard/dashboard-search_dark.png#gh-dark-mode-only)

## AI Assistant

When the AI Assistant is enabled by your hosting provider, an **AI** button is shown in the top-right corner of the header, next to the search icon. Click it to open the assistant on the right side of the page. It's available on every page, so you can keep it open while you follow its steps.

![AI button in the page header, next to the search icon and the favorites star](/img/openpanel-screenshots/dashboard/dashboard-ai-button.png#gh-light-mode-only)
![AI button in the page header, next to the search icon and the favorites star](/img/openpanel-screenshots/dashboard/dashboard-ai-button_dark.png#gh-dark-mode-only)

Ask a question in your own words, for example *How do I add a new domain?* or *Why does my website show a 502 error?*, and press **Enter** to send it. Use **Shift + Enter** for a new line. When the chat is empty, a few example questions are shown that you can click.

The assistant answers using the panel documentation and knows which features are available on your plan and which page you are on, so you can also ask *What can I do on this page?*.

![AI Assistant open on the right side of the dashboard, answering how to point a domain to the server and listing the account domains](/img/openpanel-screenshots/dashboard/dashboard-ai-window.png#gh-light-mode-only)
![AI Assistant open on the right side of the dashboard, answering how to point a domain to the server and listing the account domains](/img/openpanel-screenshots/dashboard/dashboard-ai-window_dark.png#gh-dark-mode-only)

Depending on what your hosting provider allows, the assistant can also look things up in your account, like your domains, websites, email accounts or disk usage, and make changes for you, like adding a domain or creating an email account. Ask for it in your own words, for example *Which domains do I have?* or *Create info@example.com with a 2 GB quota*. Click **Info** to see what it can do on your account.

Looking things up happens right away. Before the assistant changes anything, it shows a card with the action and the exact values it will use:

- **Approve**: runs the action once.
- **Always allow**: runs the action and doesn't ask again for the same kind of action, for example adding a domain.
- **Decline**: skips it, and the assistant won't try it again unless you ask.

![AI Assistant asking for approval to create an email account, with the request details and Approve, Always allow and Decline buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-confirm.png#gh-light-mode-only)
![AI Assistant asking for approval to create an email account, with the request details and Approve, Always allow and Decline buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-confirm_dark.png#gh-dark-mode-only)

Actions that delete or overwrite data, like removing a domain, deleting files or changing a password, are marked **Destructive** and always ask, even if you allowed a similar action before. The assistant also asks again for every change after it has read files or logs while answering the same question, because those can contain text written by others.

![Destructive action card in the AI Assistant for removing a domain, with only Approve and Decline buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-destructive.png#gh-light-mode-only)
![Destructive action card in the AI Assistant for removing a domain, with only Approve and Decline buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-destructive_dark.png#gh-dark-mode-only)

Everything the assistant looks up or changes is recorded in your [Activity log](/docs/panel/security/account_activity/), together with whether you approved it.

AI can make mistakes, so check the details on each card before you approve it.

Buttons at the top of the assistant:

- **Info**: shows what the assistant can do, how many questions you've asked this month, and the actions you've set to **Always allow**. Click **Remove** next to an action to have the assistant ask again. Your hosting provider can set a monthly limit.
- **New chat**: clears the conversation and starts over.
- **Close**: hides the assistant. You can also press **Esc**.

![AI Assistant info panel with what the assistant can do, questions used this month and the always allowed actions with Remove buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-info.png#gh-light-mode-only)
![AI Assistant info panel with what the assistant can do, questions used this month and the always allowed actions with Remove buttons](/img/openpanel-screenshots/dashboard/dashboard-ai-info_dark.png#gh-dark-mode-only)

The conversation is kept in your browser while the tab is open, so it stays when you move between pages and is cleared when you close the tab.

## Getting Started Tour

On first login, a short guided tour highlights the main areas of the interface. You can skip it at any time; it won't reappear once dismissed.

The tour can be started from the last step of the [Onboarding](/docs/panel/dashboard/onboarding/) wizard.

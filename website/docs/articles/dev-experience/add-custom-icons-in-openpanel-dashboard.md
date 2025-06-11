# Add custom icons in OpenPanel Dashboard

To add custom section in *OpenPanel > Dashboard* page, navigate to **OpenAdmin > Settings > Custom Code** and edit the 'Custom Section'.

Example code:

```json
{
  "section_title": "my.openpanel.com",
  "section_position": "before_domains",
  "items": [
    {
      "label": "Manage Profile",
      "icon": "bi bi-person-fill-gear",
      "url": "https://panel.hostio.rs/clientarea.php?action=details"
    },
    {
      "label": "Manage Billing Information",
      "icon": "bi bi-credit-card",
      "url": "https://panel.hostio.rs/clientarea.php?action=details"
    },
    {
      "label": "View Email History",
      "icon": "bi bi-envelope-open",
      "url": "https://panel.hostio.rs/clientarea.php?action=emails"
    },
    {
      "label": "News & Announcements",
      "icon": "bi bi-megaphone-fill",
      "url": "https://panel.hostio.rs/index.php?rp=/announcements"
    },
    {
      "label": "Knowledgebase",
      "icon": "bi bi-book-half",
      "url": "https://panel.hostio.rs/index.php?rp=/knowledgebase"
    },
    {
      "label": "Server Status",
      "icon": "bi bi-hdd-network",
      "url": "https://panel.hostio.rs/serverstatus.php"
    },
    {
      "label": "Invoices",
      "icon": "bi bi-receipt",
      "url": "https://panel.hostio.rs/clientarea.php?action=invoices"
    },
    {
      "label": "Support Tickets",
      "icon": "bi bi-life-preserver",
      "url": "https://panel.hostio.rs/supporttickets.php"
    },
    {
      "label": "Open Ticket",
      "icon": "bi bi-journal-plus",
      "url": "https://panel.hostio.rs/submitticket.php"
    },
    {
      "label": "Register New Domain",
      "icon": "bi bi-globe",
      "url": "https://panel.hostio.rs/cart.php?a=add&domain=register"
    },
    {
      "label": "Transfer Domain",
      "icon": "bi bi-arrow-repeat",
      "url": "https://panel.hostio.rs/cart.php?a=add&domain=transfer"
    }
  ]
}
```

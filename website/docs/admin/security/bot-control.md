---
sidebar_position: 2.5
---

# AI & Bot Control

The **Security > AI & Bot Control** page lets you allow, rate-limit or block known AI crawlers on every website on the server, and shows who is hitting the server: which bots, from which networks and on which URLs.

Search engines like Googlebot and Bingbot and normal visitors are never affected, the rules only match the user agents of known AI crawlers: GPTBot, ChatGPT-User, OAI-SearchBot, ClaudeBot, Claude-Web, anthropic-ai, PerplexityBot, Perplexity-User, Bytespider, CCBot, Amazonbot, Meta-ExternalAgent, Meta-ExternalFetcher, FacebookBot, Diffbot, Omgilibot, cohere-ai, YouBot, DuckAssistBot, Timpibot and Scrapy.

The page needs the `bot_control` module. While it's off, a box at the top of the page says so, with an **Enable module** button. Turning the module on also shows the [AI & bot control page](/docs/panel/security/bot-control/) to users whose plan has the `bot_control` feature.

![AI & Bot Control page header with the box saying the module is disabled and the Enable module button](/img/openadmin-screenshots/security/bot-control-page.png#gh-light-mode-only)
![AI & Bot Control page header with the box saying the module is disabled and the Enable module button](/img/openadmin-screenshots/security/bot-control-page_dark.png#gh-dark-mode-only)

## Policies

Each domain has one of three policies:

| Policy | What AI crawlers get |
|---|---|
| **Allow** | Normal access, the default for every domain. |
| **Rate-limit** | The [rate limit](#rate-limit) of requests per minute per crawler on that domain, over that HTTP `429`. |
| **Block (403)** | HTTP `403` on every request. |

The policy is one line in the domain's Caddy config, `/etc/openpanel/caddy/domains/<domain>.conf`: `# botcontrol: allow`, `import botcontrol_ratelimit <domain>` or `import botcontrol_block`. The snippets they import are in `/etc/openpanel/caddy/botcontrol.conf`.

Change the policy of one domain from the **Change policy** column of the **Per-site AI crawler policies** table. To set every domain at once, pick a policy next to **Apply to all domains** at the top of the page.

![Per-site AI crawler policies table with the hits, AI hits and blocked requests of each domain, its policy and the Change policy dropdown](/img/openadmin-screenshots/security/bot-control-policies.png#gh-light-mode-only)
![Per-site AI crawler policies table with the hits, AI hits and blocked requests of each domain, its policy and the Change policy dropdown](/img/openadmin-screenshots/security/bot-control-policies_dark.png#gh-dark-mode-only)

Caddy is reloaded after every change. If Caddy rejects the new config, every changed file is put back and the error message shows Caddy's reason. Domains whose config has no place for the rule (suspended domains or empty configs) are skipped and listed in the message instead of failing the whole change.

## Rate limit

The **Rate limit for AI crawlers** box sets how many requests per minute each AI crawler gets on a domain set to **Rate-limit**. The default is **10**, the allowed range is 1 to 100000.

![Rate limit for AI crawlers box with the requests per minute field and the Save button](/img/openadmin-screenshots/security/bot-control-rate-limit.png#gh-light-mode-only)
![Rate limit for AI crawlers box with the requests per minute field and the Save button](/img/openadmin-screenshots/security/bot-control-rate-limit_dark.png#gh-dark-mode-only)

The limit is counted separately for every crawler on every domain, so GPTBot hitting its limit on one site doesn't slow down ClaudeBot or GPTBot on other sites. It's keyed on the User-Agent instead of the IP address, because crawlers come from hundreds of addresses.

The value is saved as `events` in the `botcontrol_ratelimit` snippet in `/etc/openpanel/caddy/botcontrol.conf` and Caddy is reloaded. If Caddy rejects it, the previous value is put back. Only administrators can change it: resellers and users see the current value but can't edit it.

## Who's hitting the server

The **Who's hitting the server** section reads the Caddy access logs of all domains, or of the domain picked in the dropdown, for the last hour, 24 hours or 7 days:

- **Domains**, **Total requests**, **AI crawler hits** and **Blocked / throttled** (HTTP `403` and `429` responses) at the top.
- **Top bots & user-agents** with the operator, category (AI Training, AI Search, Scraper, Search Engine, Browser / Other), hits and blocked hits.
- **Top networks & ASNs** with the number of IPs and hits from each network. ASN names come from `/etc/openpanel/asn_prefixes.tsv` when that file exists.
- **Top hit URLs** with total, bot and blocked hits.

![Who's hitting the server section with the totals, top bots and user-agents, top networks and top hit URLs](/img/openadmin-screenshots/security/bot-control-traffic.png#gh-light-mode-only)
![Who's hitting the server section with the totals, top bots and user-agents, top networks and top hit URLs](/img/openadmin-screenshots/security/bot-control-traffic_dark.png#gh-dark-mode-only)

The **Per-site AI crawler policies** table shows the same counts for each domain, click a domain to see the report for just that domain. Click a column header to sort by domain, owner, hits, AI hits, blocked hits or policy (from Allow to Block), the time window and the picked domain stay. On a [multi-server](/docs/admin/system/multi-server/) master the totals, the domain picker and the tables cover every server, the tables have a **Server** column and the server picker is on the right of the page title.

## Resellers

Resellers can open the page from the menu and from search while the `bot_control` module is on. They see only the domains of their own users and those domains' traffic, and can change the policy of those domains, also with **Apply to all domains**.

Resellers can't turn the module on or off and can't change the [rate limit](#rate-limit), the value the administrator sets applies to their users too.

## Disable

**Disable module** turns off the `bot_control` module, puts every domain back on **Allow** and restarts OpenPanel so users no longer see the page. The module can also be turned on or off on the [Modules page](/docs/admin/settings/modules/#bot-control).

## API

The same actions are available through the [OpenAdmin API](/docs/admin/api/#ai--bot-control): `GET /api/security/bot-control` and `POST /api/security/bot-control`.

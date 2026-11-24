---
sidebar_position: 1.5
---

# AI & bot control

The **AI & bot control** page lets you decide what known AI crawlers can do on each of your websites, and shows who is hitting a website. Normal visitors and search engines like Googlebot and Bingbot are never affected.

The crawlers it recognizes are GPTBot, ChatGPT-User, OAI-SearchBot, ClaudeBot, Claude-Web, anthropic-ai, PerplexityBot, Perplexity-User, Bytespider, CCBot, Amazonbot, Meta-ExternalAgent, Meta-ExternalFetcher, FacebookBot, Diffbot, Omgilibot, cohere-ai, YouBot, DuckAssistBot, Timpibot and Scrapy.

The page is available when the `bot_control` feature is enabled on your hosting plan.

## Policies

The page lists your domains with their current policy. Pick one of three for each domain:

| Policy | What AI crawlers get |
|---|---|
| **Allow** | Normal access, the default. |
| **Rate-limit** | A limited number of requests per minute per crawler, over that HTTP `429`. The number is set by the server administrator, the page shows it in the **Rate-limit** label (10 by default). |
| **Block** | HTTP `403` on every request. |

The change is applied right away. If the web server rejects it, nothing is changed and you see the reason.

## Who's hitting me

Click a domain to open its report, for the last hour, 24 hours or 7 days. It's built from the domain's access log and shows:

- **Total requests**, **AI crawler & scraper hits** and **AI crawlers blocked / throttled** (403 or 429 sent to an AI crawler). If other visitors also got 403 or 429, a note shows how many: those come from the site's per-IP rate limit or the WAF, not from bot control.
- **Top bots & user-agents** with the operator, category, hits and blocked hits.
- **Top networks & ASNs** with the number of IPs and hits from each network.
- **Top hit URLs** with total, bot and blocked hits.

You can change the domain's policy on the same page. If the domain had no visits yet, the page says no access log was found, the stats appear once the web server logs requests.

## API

The same is available through the [OpenPanel API](/docs/panel/api/#ai--bot-control): `GET /api/bot-control`, `POST /api/bot-control/<domain>` and `GET /api/bot-control/<domain>/stats`.

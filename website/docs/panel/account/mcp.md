---
sidebar_position: 6
---

# MCP

OpenPanel exposes an [MCP](https://modelcontextprotocol.io) server at `/mcp`. Once connected, Claude (or any other MCP client) can manage your account directly: list and create databases, manage domains, DNS records, cron jobs, email accounts, and everything else available under `/api/`.

On **Account > AI Assistant (MCP)** page in the UI, you can generate and manage tokens to be used with Claude (MCP).

![MCP page with the token generator and connection snippets for Claude Code and Claude Desktop](/img/openpanel-screenshots/account/mcp-page.png#gh-light-mode-only)
![MCP page with the token generator and connection snippets for Claude Code and Claude Desktop](/img/openpanel-screenshots/account/mcp-page_dark.png#gh-dark-mode-only)

[OpenPanel API documentation](/docs/panel/api/)


## Requirements

- Server must be running the [Enterprise edition](/enterprise/).
- Your plan must have both the **API** and **MCP** features enabled. If `/account/mcp` isn't  visible in your account sidebar, ask your host to enable them.

## 1. Generate a token

1. Go to **Account > AI Assistant (MCP)** (`/account/mcp`).
2. Click **Generate token**, give the token a name (e.g. `Laptop - Claude Desktop`) and click **Generate new token**.
3. Copy the token immediately: it's only shown once, right after creation.

When generating a token you can also set:

- **Expiry**: never, 30 days, 90 days, or 1 year. By default tokens don't expire, so set one of the time limits if you'd rather it stop working on its own after a while.
- **Read-only**: restricts the token to GET/HEAD requests only, so Claude can look things up (list sites, check disk usage, read DNS records, etc.) but can't create, change, or delete anything. Recommended for anything other than your own trusted assistant.
- **Allow deletes and other destructive actions**: on by default. Untick it to let the assistant create and change things but not remove them. It also blocks restores, password changes, suspends and overwriting config files. The token list shows these tokens as **No deletes**.
- **Features**: **All features your account has today** is ticked by default. Untick it to pick only some, for example **DNS Zone Editor** and **Domains**.
- **Sensitive features**: **Account** (password and email changes, sessions), **2FA**, **Passkeys** and **phpMyAdmin** are never included by default, not even with **All features**. Tick them only if the assistant really needs them: together they're close to full control of your account.

A token keeps the feature list it was made with. Features added to your account later aren't added to existing tokens, so make a new token when you need one of them. Tokens made before OpenPanel 2.0.30 had no list. They were pinned once to the features the account had when the panel was updated, sensitive ones included, so nothing they did before stops working.

No token can create or revoke tokens, not even a full-access one, so a token can never hand itself more access. Manage tokens on this page, or through the API with your login.

phpMyAdmin login links and WordPress database passwords (`GET /api/sites/<domain>/wp-info`) need a full-access token that includes the feature. Read-only and no-deletes tokens never get them, because a database login would get around those limits.

These settings are shown next to each token afterwards, and revoking a token any time from the same page takes effect immediately if it's no longer needed or you think it leaked.

The limits are checked on every API request, so they also apply when the token is used against `/api/` directly instead of through `/mcp`. Every change a token makes is written to the account's activity log with the token's name.

Through the API, `POST /api/account/mcp` takes `"mode": "readonly" | "modify" | "full"` and `"features": ["dns", "domains"]`. The account must have every feature in the list, otherwise it answers 400 with the missing ones. Without `features` the token gets every feature the account has except the sensitive ones above. The old `"read_only": true` still works.

## 2. Connect

The panel's MCP endpoint is `https://<your-panel-domain>:<port>/mcp` (also shown on `/account/mcp`).

### Claude Code (CLI)

```
claude mcp add --transport http openpanel https://<your-panel-domain>/mcp \
  --header "Authorization: Bearer <your-token>"
```

### Claude Desktop

Claude Desktop starts local MCP servers only, so it reaches OpenPanel through [`mcp-remote`](https://www.npmjs.com/package/mcp-remote), which needs Node.js. Add this to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "openpanel": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "https://<your-panel-domain>:<port>/mcp", "--header", "Authorization:${OPENPANEL_AUTH}"],
      "env": { "OPENPANEL_AUTH": "Bearer <your-token>" }
    }
  }
}
```

Restart Claude Desktop after saving. OpenPanel's tools should now show up in the tool picker.

### Codex

Add this to `~/.codex/config.toml`:

```toml
[mcp_servers.openpanel]
url = "https://<your-panel-domain>:<port>/mcp"
http_headers = { Authorization = "Bearer <your-token>" }
```

### Cursor

Add this to `~/.cursor/mcp.json`, or to `.cursor/mcp.json` in a project:

```json
{
  "mcpServers": {
    "openpanel": {
      "url": "https://<your-panel-domain>:<port>/mcp",
      "headers": { "Authorization": "Bearer <your-token>" }
    }
  }
}
```

### VS Code / GitHub Copilot

Add this to `.vscode/mcp.json` in your project:

```json
{
  "servers": {
    "openpanel": {
      "type": "http",
      "url": "https://<your-panel-domain>:<port>/mcp",
      "headers": { "Authorization": "Bearer <your-token>" }
    }
  }
}
```

### Gemini CLI

```
gemini mcp add --transport http --header "Authorization: Bearer <your-token>" openpanel https://<your-panel-domain>:<port>/mcp
```

The **AI Assistant (MCP)** page shows all of these with your panel's address filled in, and with the token too right after you generate one. Server admins can create a token for a user and print the same setup from the terminal with [`opencli user-mcp`](/docs/opencli/user/#ai-assistant-mcp-tokens).

## 3. Using it

Once connected, just ask Claude in plain language, e.g.:

- "List all my websites and their status"
- "Create a new MySQL database called `shop` and a user for it"
- "Add an A record for `shop.example.com` pointing to 1.2.3.4"
- "Show me disk usage for example.com"
- "Create a daily cron job that runs `/var/www/html/backup.sh` at 2am"

Every action Claude takes runs through the exact same `/api/` endpoints and permission checks as the panel UI and REST API: Claude only ever operates within your account and your plan's enabled features. If a request fails with a permission error, it usually means that feature isn't enabled on your plan. With a read-only token, anything beyond looking things up will be blocked instead.

## Deploying a git repository

The assistant can put a public git repository online in one step:

- `deploy_detect` looks at the repository and says how it would run: Node.js, Python, Ruby or Java, the framework, the start command and the port. It changes nothing.
- `deploy` takes `repo_url` and `domain` (and optionally `subdirectory`), installs the app the same way **Node.js / Python** apps are installed in the panel, adds the reverse proxy, and waits until the app answers. Anything detection gets wrong can be overridden: `app_type`, `start_cmd`, `port`, `version`, `mem_limit`.
- `deploy_status` shows whether the app is running, whether it answers on its port, and its newest log lines.
- To redeploy after a push, restart the app (`pm2_domain_restart`, with `no_pull` set to skip downloading the runtime image again). It fetches the repository again, and `npm run build` only runs when the commit changed since the last build, so a plain restart takes seconds.
- To change the start command, port or memory limit, delete the app and deploy again.

Detection supports Next.js, Nuxt, SvelteKit, Remix, Astro, NestJS, Express, Fastify, Vite and Create React App builds, plain Node.js, Django, FastAPI, Flask, Rails, Rack/Sinatra, Spring Boot and plain HTML. Only public `https://` repositories work, and a repository with only a `Dockerfile` isn't supported yet; add it from **Containers** instead.

Example prompt: "Deploy https://github.com/heroku/node-js-getting-started to example.com/demo and tell me when it's up."

## How tools are listed

OpenPanel has a few hundred API endpoints, too many to hand to an assistant at once. So the MCP server lists a short set of read-only tools (websites, domains, databases, containers, files, usage) plus two helpers:

- `search_tools` finds a tool by what you want to do, for example "add dns record" or "install wordpress".
- `call_tool` runs a tool that `search_tools` returned.

Only tools your plan and your token allow are listed or returned by search.

Server admins who want every tool listed directly can set `mcp_tool_search=off` in the panel configuration and restart OpenPanel.

## Bulk Actions

Tick the checkbox of one or more tokens, or the checkbox in the table header to select every token shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two MCP tokens selected with the bulk actions bar offering Revoke](/img/openpanel-screenshots/account/mcp-bulk.png#gh-light-mode-only)
![Two MCP tokens selected with the bulk actions bar offering Revoke](/img/openpanel-screenshots/account/mcp-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Revoke** | Revokes the selected tokens. MCP clients using them stop working immediately. |

Click an action, confirm it, and it runs on the selected tokens one after another. When it's done the page reloads with a notice listing the tokens it worked for, or which ones failed and why.

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| `/account/mcp` not visible | The `mcp` (or `api`) feature isn't enabled on your plan. |
| Client can't connect / 401 | Token is wrong, was revoked, expired, or the header isn't set correctly. |
| Tool call returns "Access denied" | The specific feature it needs isn't enabled on your plan, even though MCP itself is. The `hint` says which one, e.g. "This account doesn't have the 'mysql' feature". |
| Tool call returns "This token is read-only" | The token was created with the read-only option and the action needs to create, change, or delete something. Generate a new, non-read-only token instead. |
| Tool call returns "This token can't delete or overwrite things" | The token was created with **Allow deletes** unticked. Do that step yourself in the panel, or use a full-access token. |
| Tool call returns "This token isn't allowed to use the '…' feature" | The feature isn't in the token's list, either because it was left out or because it was added to your account after the token was made. Make a new token that includes it. |
| Tool call returns "This token can't read database passwords or log in to phpMyAdmin" | phpMyAdmin and `wp-info` need a full-access token. |
| Tool call returns "Tokens can't create or revoke tokens" | Tokens never manage tokens. Create or revoke it on **Account > AI Assistant (MCP)**. |
| "Could not create token: the panel database is missing the columns tokens need" | The panel couldn't add the `mode` and `allowed_features` columns to its `mcp_tokens` table. Ask your host to restart OpenPanel. |
| Tool list is empty or very short | The panel admin hasn't enabled the `api` module panel-wide, so most `/api/*` routes don't exist yet. |

## Security notes

- Treat an MCP token like a password: anyone with it can act on your account through every
  enabled `/api/` endpoint (unless it's read-only).
- Prefer a read-only or no-deletes token, limited to the features the assistant needs, and a short expiry for anything other than your own trusted assistant: it meaningfully limits what a leaked token could do.
- Tokens are stored hashed; support cannot recover a lost token, only issue a new one.
- Revoking a token takes effect immediately.

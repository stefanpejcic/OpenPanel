---
title: "Use AI to Manage Your Server: OpenPanel Docs Now Available via llms.txt"
description: "Feed the complete OpenPanel documentation into ChatGPT, Claude, or Cursor and get accurate answers about installation, opencli commands, and server management."
slug: openpanel-llms-usage-guide
authors: stefanpejcic
image: https://openpanel.com/img/blog/openpanel_llms.png
hide_table_of_contents: true
---

OpenPanel documentation is now available in an AI-friendly format at [openpanel.com/llms.txt](https://openpanel.com/llms.txt). This means you can give ChatGPT, Claude, Cursor, or any other AI assistant instant access to our complete, up-to-date docs — and get accurate answers instead of guesses.

Here's how to use it.

<!-- truncate -->

## What is this file?

**[openpanel.com/llms.txt](https://openpanel.com/llms.txt)** contains the complete OpenPanel documentation in a single markdown file, regenerated automatically with every documentation update.

AI models often have outdated or incomplete knowledge about OpenPanel. By pointing them at this file, you're giving them the current docs — including the latest `opencli` commands, API endpoints, and features.

## ChatGPT

Paste the URL directly into your prompt:

```
Read https://openpanel.com/llms.txt and then answer:
how do I set up a DNS cluster with two OpenPanel servers?
```

ChatGPT will fetch the file and answer using the actual documentation instead of stale training data.

## Claude

Same approach — paste the link in your message:

```
Use the docs at https://openpanel.com/llms.txt as your reference.
How do I limit CPU and RAM for a specific user?
```

If you use **Claude Projects**, add the file once as project knowledge: download `llms.txt` and upload it to your project. Every conversation in that project then has full OpenPanel context without re-fetching.

## Cursor

In Cursor, add our docs as a documentation source:

1. Open **Settings → Features → Docs**
2. Click **Add new doc**
3. Enter `https://openpanel.com/llms.txt`

Then reference it in any chat with `@Docs` and select OpenPanel. Cursor will use the documentation when writing scripts against the OpenPanel API, `opencli`, or building integrations.

## Claude Code and other CLI agents

Working on automation scripts for your OpenPanel server? Tell your coding agent to load the docs first:

```
Fetch https://openpanel.com/llms.txt for OpenPanel documentation,
then write a bash script that creates 10 users with the nginx-mariadb plan.
```

## OpenPanel MCP: go beyond docs

Reading docs is one thing — since **OpenPanel 1.7.65**, AI assistants can also *manage your server directly* through our [MCP (Model Context Protocol) support](https://openpanel.com/docs/panel/account/mcp/), covering 160+ API endpoints. Combine the two: llms.txt gives your AI the knowledge, MCP gives it the hands.

## Example prompts to try

Once your AI has the docs loaded:

- "What's the difference between the Community and Enterprise editions?"
- "Walk me through migrating accounts from cPanel to OpenPanel"
- "Which opencli command changes a user's PHP version?"
- "Write a monitoring script that alerts me when a user's container is down"

## Why we did this

AI assistants are how a growing number of admins troubleshoot and learn. If the model has stale training data, you get wrong answers and wasted time. Publishing our docs in llms.txt format costs us nothing to maintain (it's generated on every docs build) and gives you a reliable way to bring any AI up to speed on OpenPanel in one prompt.

Try it: [openpanel.com/llms.txt](https://openpanel.com/llms.txt)

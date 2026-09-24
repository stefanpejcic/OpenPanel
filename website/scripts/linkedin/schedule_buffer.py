#!/usr/bin/env python3
"""Schedule LinkedIn posts for the docs articles listed in pages.txt through the Buffer API.

Dry run (default) - builds posts.json and prints the schedule, sends nothing:
    python3 schedule_buffer.py
Send to Buffer:
    BUFFER_API_KEY=xxx python3 schedule_buffer.py --send
Options (env): START=2026-09-28  HOUR_UTC=7  EVERY_DAYS=2  CHANNEL_ID=... (skip auto-detect)
"""
import datetime, json, os, re, sys, urllib.request
from pathlib import Path

DOCS = Path(__file__).resolve().parents[2] / "docs"
SITE = "https://openpanel.com"

# category -> (paths, hashtags, emoji)
CATEGORIES = {
    "install": (["articles/install-update/install-on-*.md", "articles/install-update/system-requirements.md"],
                "#OpenPanel #WebHosting #ControlPanel #SelfHosted #Linux", "🚀"),
    "apps": (["articles/apps/*.md"], "#OpenPanel #SelfHosted #OpenSource #WebHosting", "🧩"),
    "deploy": (["articles/websites/deploy-*.md"], "#OpenPanel #DevOps #WebDevelopment #Hosting", "⚙️"),
    "websites": (["articles/websites/wordpress-*.md", "articles/websites/password-protect-directory.md",
                  "articles/websites/change-php-version-per-domain.md", "articles/containers/restart-service-with-cron.md"],
                 "#OpenPanel #WordPress #WebHosting #PHP", "🛠️"),
    "domains": (["articles/domains/redirect-http-https-www-domain.md", "articles/domains/subdomain-addon-parked-domain.md",
                 "articles/domains/*ssl*.md", "articles/domains/lets-encrypt-*.md", "articles/domains/cloudflare-with-openpanel.md"],
                "#OpenPanel #SSL #DNS #Cloudflare #WebHosting", "🔒"),
    "email": (["articles/email/why-emails-go-to-spam.md", "articles/email/reverse-dns-ptr-record.md", "articles/email/how-to-set-up-dmarc.md",
               "articles/email/port-25-blocked-smtp-relay.md", "articles/email/test-email-with-mail-tester.md"],
              "#OpenPanel #Email #Deliverability #DMARC #SysAdmin", "📧"),
    "server": (["articles/server/change-server-hostname-or-ip.md", "articles/server/add-ip-addresses.md",
                "articles/server/how-to-add-swap.md", "articles/server/high-cpu-ram-usage.md"],
               "#OpenPanel #SysAdmin #Linux #DevOps", "🖥️"),
    "business": (["articles/hosting-business/*.md"], "#OpenPanel #WebHosting #HostingBusiness #SaaS", "💼"),
}

def page(path):
    s = path.read_text()
    desc = (re.search(r'^description:\s*"(.*)"\s*$', s, re.M) or re.search(r"^description:\s*(.*)$", s, re.M))
    title = re.search(r"^# (.+)$", s, re.M).group(1).strip()
    url = SITE + "/docs/" + re.sub(r"\.mdx?$", "", str(path.relative_to(DOCS))) + "/"
    return {"title": title, "description": desc.group(1).strip() if desc else "", "url": url}

def collect():
    # only the articles listed in pages.txt, sorted into categories by the first matching pattern
    listed = [l.strip() for l in Path(__file__).with_name("pages.txt").read_text().splitlines() if l.strip()]
    cats = {c: [] for c in CATEGORIES}
    for rel in listed:
        p = DOCS / rel
        for cat, (globs, tags, emoji) in CATEGORIES.items():
            if any(p.match(str(DOCS / g)) for g in globs):
                cats[cat].append({**page(p), "category": cat, "hashtags": tags, "emoji": emoji})
                break
        else:
            sys.exit(f"no category for {rel}")
    return {c: v for c, v in cats.items() if v}

def interleave(cats):
    # spread each category evenly over the timeline: item i of n gets position (i + 0.5) / n
    ranked = [((i + 0.5) / len(items), ci, it) for ci, items in enumerate(cats.values()) for i, it in enumerate(items)]
    return [it for _, _, it in sorted(ranked, key=lambda r: (r[0], r[1]))]

def text(p):
    return f"{p['emoji']} {p['title']}\n\n{p['description']}\n\n👉 {p['url']}\n\n{p['hashtags']}"

def buffer(query, variables=None):
    req = urllib.request.Request("https://api.buffer.com", method="POST",
        data=json.dumps({"query": query, "variables": variables or {}}).encode(),
        headers={"Content-Type": "application/json", "Authorization": "Bearer " + os.environ["BUFFER_API_KEY"]})
    with urllib.request.urlopen(req, timeout=30) as r:
        out = json.load(r)
    if out.get("errors"):
        sys.exit(f"Buffer API error: {out['errors']}")
    return out["data"]

def linkedin_channel():
    if os.environ.get("CHANNEL_ID"):
        return os.environ["CHANNEL_ID"]
    orgs = buffer("query { account { organizations { id name } } }")["account"]["organizations"]
    found = []
    for o in orgs:
        chans = buffer("query { channels(input: {organizationId: %s}) { id name displayName service } }"
                       % json.dumps(o["id"]))["channels"]
        found += [c for c in chans if c["service"].lower().startswith("linkedin")]
    for c in found:
        print(f"  LinkedIn channel: {c['id']}  {c.get('displayName') or c['name']}")
    if len(found) != 1:
        sys.exit("Found %d LinkedIn channels - set CHANNEL_ID to the company page's id and run again." % len(found))
    return found[0]["id"]

def main():
    send = "--send" in sys.argv
    start = datetime.date.fromisoformat(os.environ.get("START", "2026-09-28"))
    hour, every = int(os.environ.get("HOUR_UTC", "7")), int(os.environ.get("EVERY_DAYS", "2"))
    posts = interleave(collect())
    for i, p in enumerate(posts):
        day = start + datetime.timedelta(days=i * every)
        p["dueAt"] = f"{day.isoformat()}T{hour:02d}:00:00.000Z"
        p["text"] = text(p)
    Path(__file__).with_name("posts.json").write_text(json.dumps(posts, indent=2, ensure_ascii=False))
    for p in posts:
        print(f"{p['dueAt'][:10]}  {p['category']:<9} {p['title'][:80]}")
    print(f"\n{len(posts)} posts, {posts[0]['dueAt'][:10]} -> {posts[-1]['dueAt'][:10]}. Saved to posts.json.")
    if not send:
        print("Dry run - nothing sent. Review posts.json, then run with --send and BUFFER_API_KEY set.")
        return
    channel = linkedin_channel()
    for p in posts:
        # inline arguments exactly like Buffer's examples (JSON string escaping is valid GraphQL string escaping)
        mutation = """mutation { createPost(input: {
            text: %s, channelId: %s, schedulingType: automatic, mode: customScheduled, dueAt: %s
        }) { ... on PostActionSuccess { post { id dueAt } } ... on MutationError { message } } }""" % (
            json.dumps(p["text"]), json.dumps(channel), json.dumps(p["dueAt"]))
        res = buffer(mutation)["createPost"]
        print(("ok   " if "post" in res else "FAIL ") + p["dueAt"][:10] + "  " + (res.get("message") or p["title"][:70]))

if __name__ == "__main__":
    main()

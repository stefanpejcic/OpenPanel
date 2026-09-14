#!/usr/bin/env python3
"""Builds a week's worth of LinkedIn post drafts from the docs under
website/docs/{admin,panel,articles} (changelog is excluded - too dry for
social).

Run weekly by .github/workflows/linkedin_queue.yml. Each run:
  1. Rebuilds/extends a deterministic rotation order of every eligible doc
     page, stored in website/scripts/linkedin_state.json, so pages don't
     repeat until the whole set has been cycled through once.
  2. Takes the next 7 pages from that rotation.
  3. Writes ready-to-paste post drafts (hook + blurb + url + hashtags) to
     website/linkedin-queue/<date>.md, one per day of the coming week.
  4. Prints the same content to stdout so the workflow can drop it into a
     GitHub issue body.

Nothing here calls the LinkedIn API - posting is manual, via LinkedIn's own
native post scheduler. This just prepares the week's content so that's a
five-minute copy/paste/schedule job instead of a from-scratch one.

Usage: python3 website/scripts/generate_linkedin_queue.py
"""
import json
import os
import re
from datetime import date, timedelta

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
WEBSITE_DIR = os.path.dirname(SCRIPT_DIR)
DOCS_DIR = os.path.join(WEBSITE_DIR, "docs")
STATE_PATH = os.path.join(SCRIPT_DIR, "linkedin_state.json")
QUEUE_DIR = os.path.join(WEBSITE_DIR, "linkedin-queue")

SECTIONS = ["admin", "panel", "articles"]
SITE_URL = "https://openpanel.com"
POSTS_PER_WEEK = 7

NUM_PREFIX_RE = re.compile(r"^\d+[-_.]")
HEADING_RE = re.compile(r"^#{1,6}\s+(.*)")
INLINE_MD_RE = re.compile(r"`([^`]*)`|\*\*([^*]*)\*\*|\*([^*]*)\*|\[([^\]]*)\]\([^)]*\)")

GENERIC_TAGS = ["#OpenPanel", "#WebHosting"]
ROTATING_TAGS = ["#cPanelAlternative", "#ControlPanel", "#SelfHosted", "#DevOps", "#Sysadmin", "#Linux"]
SECTION_TAGS = {"admin": "#ServerAdmin", "panel": "#WebHosting", "articles": "#HowTo"}

HOOKS = [
    "Quick tip:",
    "Did you know?",
    "From the OpenPanel docs:",
    "New to OpenPanel? Start here:",
    "Under the hood:",
    "How-to of the week:",
    "For sysadmins:",
    "Straight from the docs:",
]


def strip_num_prefix(stem):
    return NUM_PREFIX_RE.sub("", stem)


def doc_url(relpath):
    parts = relpath.split(os.sep)
    parts[-1] = strip_num_prefix(os.path.splitext(parts[-1])[0])
    return f"{SITE_URL}/docs/{'/'.join(parts)}"


def clean_inline(line):
    def repl(m):
        return next(g for g in m.groups() if g is not None)
    line = INLINE_MD_RE.sub(repl, line)
    return line.strip()


def extract_title_and_blurb(filepath):
    with open(filepath, encoding="utf-8") as f:
        lines = f.readlines()

    # Skip frontmatter block if present.
    i = 0
    if lines and lines[0].strip() == "---":
        i = 1
        while i < len(lines) and lines[i].strip() != "---":
            i += 1
        i += 1

    title = None
    blurb_lines = []
    in_code_fence = False

    for line in lines[i:]:
        stripped = line.strip()

        if stripped.startswith("```"):
            in_code_fence = not in_code_fence
            continue
        if in_code_fence:
            continue

        heading_match = HEADING_RE.match(stripped)
        if heading_match:
            if title is None:
                title = clean_inline(heading_match.group(1))
            elif blurb_lines:
                break  # next heading ends the first prose block
            continue

        if not stripped:
            if blurb_lines:
                break  # blank line after we've collected some prose
            continue

        if stripped.startswith(("import ", "<", "![", "- ", "* ", "|", "<details>", "<summary>")):
            continue

        blurb_lines.append(clean_inline(stripped))

    blurb = " ".join(blurb_lines).strip()
    if not title:
        title = strip_num_prefix(os.path.splitext(os.path.basename(filepath))[0]).replace("-", " ").title()
    if not blurb:
        blurb = f"A quick look at {title.lower()} in OpenPanel."
    if len(blurb) > 260:
        blurb = blurb[:257].rsplit(" ", 1)[0] + "..."

    return title, blurb


def hashtags_for(relpath, index):
    parts = relpath.split(os.sep)
    section = parts[0]
    if len(parts) >= 3:
        topic = parts[1]
    else:
        topic = strip_num_prefix(os.path.splitext(parts[-1])[0])
    sub_tag = "#" + "".join(w.capitalize() for w in re.split(r"[-_]", topic)) if topic else SECTION_TAGS.get(section, "#OpenPanel")
    tags = [GENERIC_TAGS[0], SECTION_TAGS.get(section, "#OpenPanel"), sub_tag, ROTATING_TAGS[index % len(ROTATING_TAGS)]]
    # de-dupe while preserving order
    seen = []
    for t in tags:
        if t not in seen:
            seen.append(t)
    return seen[:5]


def discover_docs():
    found = []
    for section in SECTIONS:
        section_dir = os.path.join(DOCS_DIR, section)
        for root, _dirs, files in os.walk(section_dir):
            for fname in sorted(files):
                if not fname.endswith(".md") and not fname.endswith(".mdx"):
                    continue
                fpath = os.path.join(root, fname)
                relpath = os.path.relpath(fpath, DOCS_DIR)
                found.append(relpath)
    return sorted(found)


def load_state(eligible):
    if os.path.exists(STATE_PATH):
        with open(STATE_PATH, encoding="utf-8") as f:
            state = json.load(f)
    else:
        state = {"order": [], "cursor": 0}

    eligible_set = set(eligible)
    kept = [p for p in state["order"] if p in eligible_set]
    new_items = sorted(eligible_set - set(kept))
    order = kept + new_items

    cursor = state.get("cursor", 0)
    if cursor >= len(order):
        cursor = 0

    return {"order": order, "cursor": cursor}


def save_state(state):
    with open(STATE_PATH, "w", encoding="utf-8") as f:
        json.dump(state, f, indent=2)
        f.write("\n")


def next_batch(state, count):
    order = state["order"]
    n = len(order)
    if n == 0:
        return []
    picks = [order[(state["cursor"] + i) % n] for i in range(min(count, n))]
    state["cursor"] = (state["cursor"] + len(picks)) % n
    return picks


def build_week_markdown(picks, start_date):
    lines = [f"# LinkedIn queue — week of {start_date.isoformat()}", ""]
    lines.append(
        "Paste each block into LinkedIn's post composer and use **Schedule** "
        "(clock icon) to queue it for the date shown. Takes ~5 minutes for the week."
    )
    lines.append("")
    for idx, relpath in enumerate(picks):
        fpath = os.path.join(DOCS_DIR, relpath)
        title, blurb = extract_title_and_blurb(fpath)
        url = doc_url(relpath)
        hook = HOOKS[idx % len(HOOKS)]
        tags = hashtags_for(relpath, idx)
        post_date = start_date + timedelta(days=idx)

        post_text = f"{hook} {title}\n\n{blurb}\n\n{url}\n\n{' '.join(tags)}"

        lines.append(f"## {post_date.strftime('%A')}, {post_date.isoformat()}")
        lines.append(f"Source: `{relpath}`")
        lines.append("")
        lines.append("```")
        lines.append(post_text)
        lines.append("```")
        lines.append("")

    return "\n".join(lines)


def main():
    eligible = discover_docs()
    state = load_state(eligible)
    picks = next_batch(state, POSTS_PER_WEEK)
    save_state(state)

    start_date = date.today() + timedelta(days=(7 - date.today().weekday()) % 7 or 7)
    markdown = build_week_markdown(picks, start_date)

    os.makedirs(QUEUE_DIR, exist_ok=True)
    out_path = os.path.join(QUEUE_DIR, f"{date.today().isoformat()}.md")
    with open(out_path, "w", encoding="utf-8") as f:
        f.write(markdown + "\n")

    print(markdown)
    print(f"\nWrote {out_path}", flush=True)


if __name__ == "__main__":
    main()

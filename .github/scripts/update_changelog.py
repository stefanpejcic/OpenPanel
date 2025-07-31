import os
import re
from github import Github
from datetime import datetime

REPO = os.environ["GITHUB_REPOSITORY"]
g = Github(os.environ["GH_TOKEN"])
repo = g.get_repo(REPO)

def fmt_date(dt):
    if not dt:
        return ''
    return dt.strftime('%B %d, %Y')

# Convert offset-aware datetime to naive
def to_naive(dt):
    return dt.replace(tzinfo=None) if dt else None

milestones = list(repo.get_milestones(state="all"))

def parse_version(title):
    m = re.match(r'^(\d+)\.(\d+)\.(\d+)(.*)?$', title)
    if m:
        return tuple(map(int, m.group(1, 2, 3))) + (m.group(4) or '',)
    return (0, 0, 0, title)

def valid_version(title):
    return re.match(r'^\d+\.\d+\.\d+.*$', title)

open_milestones = [m for m in milestones if m.state == 'open' and valid_version(m.title)]
closed_milestones = [m for m in milestones if m.state == 'closed' and valid_version(m.title)]

# Safe version for sorting with naive datetime
open_milestones = sorted(
    open_milestones,
    key=lambda m: (to_naive(m.due_on) or datetime.max, parse_version(m.title))
)

closed_milestones = sorted(
    closed_milestones,
    key=lambda m: (to_naive(m.due_on) or datetime.min, parse_version(m.title)),
    reverse=True
)

def row(m):
    link = f"/docs/changelog/{m.title}"
    date = m.closed_at if m.state == "closed" else m.due_on
    date_fmt = fmt_date(date)
    return f"|__[{'%s' % m.title}]({link})__| {date_fmt} |"

upcoming_rows = [row(m) for m in open_milestones]
latest_row = row(closed_milestones[0]) if closed_milestones else ''
previous_rows = [row(m) for m in closed_milestones[1:]] if len(closed_milestones) > 1 else []

def block(title, rows):
    if not rows:
        return f"### {title}\n\n_No releases_\n"
    hdr = "| Version| Release date | \n|---|---|"
    return f"### {title}\n\n{hdr}\n" + "\n".join(rows) + "\n"

md = "# Changelog\n"
md += "\n" + block("Upcoming version", upcoming_rows)
md += "\n" + block("Latest", [latest_row] if latest_row else [])
md += "\n" + block("Previous versions", previous_rows)

os.makedirs("website/docs/changelog", exist_ok=True)
with open("website/docs/changelog/intro.md", "w") as f:
    f.write(md.strip() + "\n")

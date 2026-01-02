# Daily Obsidian Jira Notes Generator

This CLI tool generates a daily Markdown note in your Obsidian vault, pre-filled with your active Jira issues as checkboxes. It is deterministic, safe to re-run, and read-only with respect to Jira.

## What it does

Creates a daily note at:
<vault_path>/<daily_notes_subdir>/YYYY-MM-DD.md

Uses a fixed template with sections for focus, Jira tasks, and notes

Queries Jira for issues:

assignee = currentUser() AND statusCategory != Done


Renders each issue as an Obsidian-compatible task checkbox

Exits without modifying anything if the daily note already exists

## How to run

1. Build or install the binary:

```bash
git clone https://github.com/SlaterL/daily-notes.git
cd daily-notes
go build ./daily-notes
```
OR
```bash
go install github.com/SlaterL/daily-notes
```

2. Run the tool:

Create your daily note.
```bash
daily-notes --cmd start
```

Manually add a commit message (this can normally be done via a post-commit hook using the script below).
```bash
daily-notes --cmd commit --repo <repo> --msg <commit message>
```

Create a new summary note based on today's note.
```bash
daily-notes --cmd sum
```

## Post-commit script

Here is a working post-commit script to append all commits to make to the bottom of your daily note (located at `~/.githooks/post-commit`):
```bash
#!/usr/bin/env bash

set -euo pipefail

if git rev-parse -q --verify MERGE_HEAD >/dev/null; then
exit 0
fi

# --- config ---
NOTES_DIR="/ABSOLUTE/PATH/TO/YOUR/OBSIDIAN/VAULT"
DATE="$(date +%Y-%m-%d)"
NOTE_FILE="$NOTES_DIR/$DATE.md"

# --- git info ---
REPO_ROOT="$(git rev-parse --show-toplevel)"
REPO_NAME="$(basename "$REPO_ROOT")"

COMMIT_MSG="$(git log -1 --pretty=%B | tr '\n' ' ')"
[[ "$COMMIT_MSG" == fixup!* ]] && exit 0
[[ "$COMMIT_MSG" == squash!* ]] && exit 0

# --- run daily-notes ---
daily-notes --cmd commit --repo "$REPO_NAME" --msg "$COMMIT_MSG"
```

Make sure to enable it:
```bash
git config --global core.hooksPath ~/.githooks
```

## Configuration

Create the following config file:
`~/.config/daily-notes/config.yaml`

Example:
```yaml
vault_path: "/ABSOLUTE/PATH/TO/YOUR/OBSIDIAN/VAULT"
daily_notes_subdir: "Daily"

jira:
    base_url: "https://yourcompany.atlassian.net"
    email: "you@company.com"
    token: "<INSERT TOKEN>"
    project_filter: ["CORE", "PROJ"]
readme: false
exclude_commits: ["fixup!"]
model: "nemotron-3-nano:30b"
```

Notes:
* vault_path must be an absolute path
* project_filter is optional (omit or leave empty to include all projects)

## Output behavior

* If the daily note does not exist → it is created
* If the daily note already exists → no changes are made
* If there are no active Jira issues → the note is still created with an empty Jira section
* Any error (config, Jira API, disk write) causes a clear, fatal failure

A successfully created note should look something like this:
```md
# 2026-01-13
## 🏆 Major Accomplishments
- 

## 📋 Jira Tasks
[**CORE-1571**](https://yourcompany.atlassian.net/browse/CORE-1571) (In Progress) — Update auth flow to V2
- [ ] ...
[**CORE-1573**](https://yourcompany.atlassian.net/browse/CORE-1573) (Ready to Merge) — Rewrite payments service in Rust
- [ ] ...


## 📋 Other Tasks
- [ ] Review MRs

## 📝 Notes

```

## README config value

The `readme` config value refers to an option that can automatically embed a link to relevant readmes, if they exist in your vault. This can be done using symlinks to each of the readme's from your cloned repos. Here's a decent command for creating the symlinks:

```bash
#!/usr/bin/env bash
set -euo pipefail

SRC="$HOME/Your/Code/Path"
DEST="$HOME/Your/Obsidian/vault/docs"

mkdir -p "$DEST"

find "$SRC" -type d -name ".git" | while read -r gitdir; do
  repo_root="$(dirname "$gitdir")"
  repo_name="$(basename "$repo_root")"
  repo_dest="$DEST/$repo_name"

  mkdir -p "$repo_dest"

  find "$repo_root" \( -type f -o -type l \) \
    \( -name "*.md" -o -name "*.swagger.json" \) \
    -not -path "*/vendor/*" |
  while read -r file; do
    rel="${file#$repo_root/}"

    flat_name="$(echo "$rel" | sed 's|/|-|g')"
    target="$repo_dest/$flat_name"

    if [[ -e "$target" ]]; then
      base="$(basename "$rel")"
      dir="$(dirname "$rel" | sed 's|/|-|g')"
      flat_name="${dir}-${base}"
      target="$repo_dest/$flat_name"
    fi

    ln -sfn "$file" "$target"
  done
done
```

The program assumes all docs are in a `/docs` dir in your vault and will attempt to use the list of Components on your jira Issue when linking.
There's no magic to this, it will assume a readme exists for any component listed and may result in broken links.

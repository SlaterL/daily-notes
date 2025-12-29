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

```shell
git clone https://github.com/SlaterL/daily-notes.git
cd daily-notes
go build ./daily-notes
```
OR
```shell
go install github.com/SlaterL/daily-notes
```

2. Run the tool:

```shell
daily-notes
```

On success, a new daily note is created. If the note already exists, the tool exits cleanly without overwriting it.

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
```

Notes:
* vault_path must be an absolute path
* project_filter is optional (omit or leave empty to include all projects)

## Output behavior

* If the daily note does not exist → it is created
* If the daily note already exists → no changes are made
* If there are no active Jira issues → the note is still created with an empty Jira section
* Any error (config, Jira API, disk write) causes a clear, fatal failure

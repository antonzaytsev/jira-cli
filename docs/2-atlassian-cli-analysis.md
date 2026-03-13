# `catapultcx/atlassian-cli` (PyPI) — Comprehensive Analysis

**Package**: [atlassian-cli v0.5.9](https://pypi.org/project/atlassian-cli/)
**Source**: [catapultcx/atlassian-cli](https://github.com/catapultcx/atlassian-cli)
**Language**: Python 3.10+
**License**: MIT
**Dependencies**: `requests` (single dependency)
**Last release**: March 5, 2026
**Monthly downloads**: ~1,364

---

## What It Is

Two CLI binaries — `confluence` and `jira` — that wrap Atlassian Cloud REST APIs.
Explicitly designed for AI agents. Uses Confluence REST API v2 and Jira REST API v3.

Architecture: ~770 lines total across 7 source files.

```
src/atlassian_cli/
  config.py       Auth + session factory (.env or env vars)
  http.py         GET/POST/PUT/DELETE with 429 retry + backoff
  output.py       Text and JSON output formatting
  confluence.py   Confluence CLI (v2 API, ADF)
  jira.py         Jira CLI entry point (subparsers)
  jira_issues.py  Issue CRUD, search, transitions, comments
  jira_assets.py  JSM Assets CRUD
```

---

## Confluence CLI — Feature Coverage

| Feature | Supported | Notes |
|---------|-----------|-------|
| Get page (ADF body) | Yes | Downloads to local JSON file |
| Create page | Yes | From ADF file or plain text |
| Update page | Yes | `put` with version conflict detection |
| Delete page | Yes | |
| Diff local vs remote | Yes | Unified diff on ADF JSON |
| Bulk sync (space) | Yes | Parallel workers, version-cached |
| Search (local index) | Yes | Title/ID search on pre-built index |
| Build page index | Yes | From API, supports multiple spaces |
| List comments (inline + footer) | Yes | With author resolution, replies |
| Reply to comment | Yes | Inline and footer |
| Resolve/reopen comment | Yes | Via inline comment API |
| ADF hints for agents | Yes | Built-in `hints` command |
| Attachments | **No** | Not implemented |
| CQL search (remote) | **No** | Only local index search |
| Space CRUD | **No** | Only read (for page listing) |
| Labels | **No** | Not implemented |
| Page tree / hierarchy | **No** | Flat page listing only |

### Confluence Strengths

1. **ADF-native**: Reads and writes Atlassian Document Format (v2 API). No lossy conversion.
2. **Local file model**: Pages stored as `{id}.json` + `{id}.meta.json`. Enables `diff`,
   offline editing, and version-cached `sync`.
3. **Parallel sync**: ThreadPoolExecutor-based bulk download with configurable worker count.
4. **Comment management**: Full inline/footer comment lifecycle (list, reply, resolve, reopen)
   with author name resolution and reply threading.
5. **Rate limit handling**: Exponential backoff with jitter on 429 responses.

### Confluence Gaps

1. **No attachments**: Cannot upload/download/list attachments. Must fall back to curl.
2. **No remote CQL search**: `search` only queries the local index (title/ID match).
   Cannot run CQL like `type = page AND space = DEV AND text ~ "migration"`.
3. **No labels**: Cannot list, add, or remove page labels.
4. **Page addressing by ID only**: Must know the numeric page ID. No `Space:Title` lookup.
5. **Hardcoded default spaces**: `index` command defaults to `['POL', 'COMPLY']` when
   no `--space` is provided — specific to the author's use case.
6. **No page tree navigation**: Flat page list only. No parent/child traversal, no `--tree`.

---

## Jira CLI — Feature Coverage

| Feature | Supported | Notes |
|---------|-----------|-------|
| Get issue | Yes | Key, status, summary |
| Create issue | Yes | Project, type, summary, description, labels, assignee, parent |
| Update issue | Yes | Summary, description, labels (set/add/remove), assignee, raw JSON fields |
| Delete issue | Yes | With subtask option |
| JQL search | Yes | Paginated via POST `/search/jql`, field selection, `--dump` to file |
| Transitions | Yes | With **discovery** — lists available transitions on mismatch |
| Add comment | Yes | ADF body |
| List comments | Yes | With author and date |
| Custom fields via --fields | Yes | Raw JSON passthrough |
| Attachments | **No** | Not implemented |
| View full issue (all fields) | **No** | Only returns key/status/summary |
| Comment CRUD (edit/delete) | **No** | Add and list only |
| Watchers | **No** | |
| Worklogs | **No** | |
| Sprint/Board | **No** | |
| Webhooks | **No** | |

### Jira Strengths — Scored Against jira-cli Investigation

Checking each critical/high issue from
[1-investigation.md](1-investigation.md):

| # | jira-cli Problem | atlassian-cli Status |
|---|------------------|---------------------|
| 1 | Destructive ADF writes | **Fixed** — uses `_text_adf()` to write proper ADF via v3 API |
| 2 | API v2 writes / v3 reads | **Fixed** — all operations use REST API v3 |
| 3 | Interactive prompts / hangs | **Fixed** — pure argparse, no interactive prompts, no `$EDITOR` |
| 4 | Comment management (add only) | **Partial** — add + list, but no edit/delete |
| 5 | No attachments | **Not fixed** — still no attachment support |
| 6 | Custom fields unreliable | **Improved** — `--fields` accepts raw JSON; still no named custom fields |
| 7 | No transition discovery | **Fixed** — `cmd_transition` fetches available transitions, shows them on mismatch |
| 8 | JQL mangling (ORDER BY) | **Fixed** — JQL passed verbatim to POST `/search/jql` |
| 9 | Output not programmatic | **Fixed** — `--json` flag on all commands, structured emit |
| 10 | No field selection | **Fixed** — `--fields` param on search (defaults to `summary,status,assignee,issuetype`) |
| 11 | Debug leaks auth token | **Fixed** — no debug mode that prints headers |
| 12 | Project context forced | **Fixed** — no config-file project binding; JQL is free-form |
| 13 | No bulk operations | **Not fixed** — one issue per call |
| 14 | No webhooks | **Not fixed** |

**Score: 9/14 issues fixed, 2 partially fixed, 3 not addressed.**

### Jira Gaps

1. **Issue `get` returns minimal data**: Only prints `KEY [Status] Summary`. No description,
   no custom fields, no reporter, no timestamps. No `--raw` or `--verbose` to get full fields.
2. **No attachment support**: Same gap as Confluence side.
3. **Comment add only**: Cannot edit or delete comments (issue #4 from investigation).
4. **Description writes are plain-text only**: `_text_adf()` wraps input as a single paragraph.
   Cannot write ADF with headings, tables, code blocks, links. To write rich content, an agent
   must construct ADF JSON and pass it via `--fields '{"description": {...}}'`.
5. **No `view` command**: `get` is terse. No way to see the full issue with all fields
   (description, custom fields, comments, history) in a single call.

---

## Auth & Configuration

- **Env vars**: `ATLASSIAN_URL`, `ATLASSIAN_EMAIL`, `ATLASSIAN_TOKEN`
- **Fallback**: `CONFLUENCE_URL`, `CONFLUENCE_EMAIL`, `CONFLUENCE_TOKEN`
- **.env file**: Auto-loaded from CWD or package directory
- **No profile support**: Single instance only. Cannot switch between staging/prod.
- **No keychain/encrypted storage**: Plaintext in `.env` or env vars.

For your setup, auth would look like:

```bash
export ATLASSIAN_URL=https://dekeo.atlassian.net
export ATLASSIAN_EMAIL=anton.zaytsev@jiffyshirts.com
export ATLASSIAN_TOKEN=$JIRA_API_TOKEN
```

This reuses your existing `JIRA_API_TOKEN` since Atlassian uses the same token for both
Jira and Confluence.

---

## Code Quality Assessment

### Positives

- **Small and readable**: ~770 lines total. Easy to audit, fork, or extend.
- **Single dependency**: Only `requests`. No heavy framework.
- **Correct API versions**: Confluence v2, Jira v3. No version mismatch.
- **Rate limiting**: Proper 429 handling with exponential backoff and jitter.
- **No interactive prompts**: Pure CLI, safe for automation.
- **Trusted publishing**: PyPI package uses Sigstore attestation from GitHub Actions.

### Concerns

- **0 GitHub stars**: Despite 1,364 monthly PyPI downloads. Very early-stage project.
- **No tests visible in PyPI package**: `dev` extra includes `pytest` + `responses`,
  but test coverage is unknown.
- **Error handling is basic**: `APIError` truncates body to 200 chars. No auth hint on 401.
- **No retry on 5xx**: Only retries on 429. Server errors fail immediately.
- **Thread safety**: `_space_cache` and `_user_cache` are module-level dicts shared across
  threads in `sync` — potential race condition (benign in practice).
- **`.env` parsing is naive**: `line.split('=', 1)` doesn't handle quoted values,
  comments after values, or multiline values.
- **`confluence search` is local-only**: Misleading command name — it searches a local JSON
  file, not the Confluence API. An agent expecting CQL search will be confused.
- **Binary name collision**: The `jira` command will conflict with the existing
  `ankitpokhrel/jira-cli` if both are installed.

---

## Comparison: atlassian-cli vs Current Setup

| Capability | Current (jira CLI + curl + MCP) | atlassian-cli |
|------------|--------------------------------|---------------|
| Jira reads | jira CLI (fast, good) | Minimal (key/status/summary only) |
| Jira writes (ADF) | curl only | CLI + raw JSON fallback |
| Jira transitions | curl only | Built-in with discovery |
| Jira JQL | Broken ORDER BY in CLI | Clean passthrough |
| Jira attachments | curl only | Not supported |
| Jira custom fields | curl only | Partial (raw JSON) |
| Confluence reads | MCP server (Cursor only) | CLI (ADF to local files) |
| Confluence writes | MCP server (Cursor only) | CLI (local edit + put) |
| Confluence comments | MCP server | Built-in (full lifecycle) |
| Confluence attachments | MCP server | Not supported |
| Confluence CQL search | MCP server | Not supported (local index only) |
| Works outside Cursor | jira CLI + curl | Yes |
| Works in Claude Code | curl only | Yes |

---

## Verdict

### As a Confluence CLI: Promising but incomplete

The local-file model (get → edit → diff → put) is a good workflow for agents. Comment
management is solid. But missing attachments, CQL search, labels, and page-by-title
addressing means you'd still fall back to the MCP server or curl for ~30% of operations.

### As a Jira CLI replacement: Not ready

It fixes the critical write issues (ADF, v3 API, no prompts, clean JQL), but the `get`
command is too minimal for agent use — only shows key/status/summary. The existing `jira`
CLI with `--plain` or `--raw` returns far more useful data for reads. You'd need both
tools installed, which creates a binary name collision on `jira`.

### Recommendation

**Install and evaluate for Confluence only.** It fills a gap you currently cover only
via the MCP server. For Jira, keep the current `jira` CLI for reads and curl for writes
until a more complete replacement emerges — or until you build your own.

If you decide to adopt it:

```bash
pip install atlassian-cli
```

Then alias the Jira binary to avoid collision:

```bash
# In ~/.zshrc or similar
alias jira-new='/path/to/atlassian-cli/jira'
```

Or only use the `confluence` binary and ignore the `jira` binary entirely.

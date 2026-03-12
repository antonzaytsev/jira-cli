# Jira CLI (`ankitpokhrel/jira-cli` v1.7.0) — Disadvantages for AI Agent Use

## Summary

The [jira-cli](https://github.com/ankitpokhrel/jira-cli) is designed for human interactive use.
It works reasonably well for quick reads but falls apart as a programmatic API for AI agents
that need reliable, fast, structured access to Jira.

Issues are ranked by **impact on agent workflows** — combining how often the limitation is hit,
how severe the consequence is, and whether a workaround exists.

---

## Critical — Blocks core agent operations

### 1. Destructive Description Editing — Corrupts ADF Content

**Impact**: Any agent write to a description silently destroys it.

The CLI's `-b` / `--body` flag and stdin pipe treat input as **plain text**, not Atlassian
Document Format (ADF). When an ADF JSON body is piped in:

```bash
echo '{"type":"doc","version":1,...}' | jira issue edit PO-2701 --no-input
```

The CLI wraps the entire JSON string as a single text paragraph, **destroying the original
rich description**. Verified: this corrupted PO-2701's description during testing (restored
via direct API call). There is no workaround — the CLI cannot write ADF at all.

### 2. API v2 for Writes, v3 for Reads — No Rich Content Writes

**Impact**: Cannot write formatted descriptions, comments with links/tables/code blocks.

Debug output reveals:
- **Reads** (view, list): `GET /rest/api/3/...`
- **Writes** (edit, comment add): `PUT /rest/api/2/...`

API v2 uses wiki markup; v3 uses ADF. The CLI reads ADF but writes plain text via v2.
Agents writing ticket analyses, implementation plans, or structured comments with headings,
tables, and links must fall back to curl every time.

### 3. Interactive Prompts / Hangs Break Automation

**Impact**: Commands hang indefinitely without `--no-input`; some hang even with it.

Without `--no-input`, commands open `$EDITOR` or prompt interactively — instant hang for an
agent. Even with `--no-input`, `edit --custom` was observed hanging for 30+ seconds before
timeout. An agent cannot reliably predict which commands will block.

---

## High — Forces curl fallback for common operations

### 4. No Comment Management — Add Only

**Impact**: Cannot update or delete comments; agents leave stale/duplicate notes.

| Operation | CLI | REST API |
|-----------|-----|----------|
| Add comment | Yes | Yes |
| Edit comment | **No** | Yes |
| Delete comment | **No** | Yes |
| List comments | Only via `view --comments N` | Yes (paginated) |

Agents frequently need to update status comments, clean up investigation notes, or replace
outdated analyses. The CLI forces add-only workflows, cluttering tickets.

### 5. No Attachment Support

**Impact**: Every file upload requires raw curl with multipart/form-data.

The CLI has no `attach` or `attachment` command. Agents uploading analysis docs, screenshots,
or verification results must construct multipart curl calls manually every time.

### 6. Cannot Set Custom Fields Reliably

**Impact**: Fields like Acceptance Criteria, Sprint, and option-type fields cannot be written.

The `--custom` flag uses simple `key=value` strings that cannot express:
- Array fields (multi-select, labels with multiple values)
- Complex objects (ADF rich text in Acceptance Criteria — `customfield_10072`)
- Option fields requiring IDs (`{"id": "10150"}`)

### 7. No Transition Discovery

**Impact**: Agent must guess transition names; wrong guess = error + retry.

`jira issue move` accepts a state name but cannot **list available transitions**. The REST API
exposes `/issue/{key}/transitions`. Without it, agents hard-code state names and break when
workflows change.

---

## Medium — Degrades convenience and reliability

### 8. JQL Mangling — Silently Appends `ORDER BY`

**Impact**: Raw JQL queries with custom sort order fail silently.

The `list -q` flag **always appends** `ORDER BY created DESC`:

```
Input:  -q"... ORDER BY updated DESC"
Sent:   ... ORDER BY updated DESC ORDER BY created DESC
Result: 400 Bad Request
```

Workaround: omit ORDER BY and accept the CLI's default sort, or use curl.

### 9. Output Not Designed for Programmatic Parsing

**Impact**: Agent must handle ANSI codes, inconsistent columns, and 40KB+ responses.

- `--plain` leaks ANSI escape codes in footer
- Tab-delimited columns with no JSON option for `list`
- `--raw` always fetches all fields (`fields=*all`), returning 40KB+ per ticket
- No field selection — the API supports `?fields=summary,status` but the CLI doesn't expose it

### 10. No Field Selection — Always Fetches Everything

**Impact**: Wasted bandwidth and token budget; agent context fills with irrelevant data.

Every call uses `fields=*all`. When an agent only needs to check a ticket's status or assignee,
it still receives every custom field, changelog entry, and attachment metadata. With curl,
`?fields=status` returns a few hundred bytes.

### 11. Debug Mode Leaks Authentication Token

**Impact**: Security risk in logged agent sessions.

`--debug` prints the full Base64-encoded `Authorization: Basic` header to stdout. In agent
contexts where output is persisted in conversation transcripts, this is a credential leak.
No way to get request diagnostics without exposing the token.

### 12. Project Context Forced on All Queries

**Impact**: Cross-project queries require extra flags or JQL workarounds.

The CLI wraps all queries in the config file's project context. When working across PO, MN,
JED, the agent must pass `-p PROJECT` on every call, maintain multiple configs, or use `-q`
(which has the ORDER BY bug).

---

## Low — Nice to have but rarely blocking

### 13. No Bulk Operations

No bulk editing, bulk transitions, or batch comment operations. Each ticket requires a separate
CLI invocation. The REST API supports `/bulk` endpoints.

### 14. No Webhook / Event Support

Cannot create or manage webhooks or poll for changes efficiently. Agents monitoring ticket
state must issue repeated `view` calls.

---

## Read Performance — Not a Significant Issue

Re-tested timings (initial 222s measurement was a sandbox network artifact):

| Command | Time |
|---------|------|
| `jira issue view PO-2701 --plain` | **0.67s** |
| `jira issue view PO-2701 --raw` | **0.65s** |
| `jira issue list --plain -q"..."` | **0.58s** |
| `curl …?fields=summary,status` | **0.44s** |
| `curl …` (all fields) | **0.61s** |

Read performance is comparable. The CLI overhead is ~0.1s — negligible.

---

## Comparison Table

| Capability | CLI | curl + REST API v3 |
|------------|-----|-------------------|
| Read performance | 0.6–0.7s | 0.4–0.6s |
| ADF description read/write | Read only (lossy writes) | Full fidelity |
| Field selection | No (`*all` always) | Yes (`?fields=...`) |
| Comment CRUD | Add only | Full CRUD |
| Attachment upload | No | Yes |
| Transition discovery | No | Yes |
| Custom field types | Simple strings only | Full JSON types |
| Bulk operations | No | Yes |
| Output format | Text + ANSI | Clean JSON |
| Auth token safety | Leaks in debug | Under agent control |
| JQL passthrough | Mangled (appends ORDER BY) | Exact |
| Automation safety | Hangs without --no-input | Always non-interactive |

---

## Conclusion

The CLI is adequate for **reads** — `view --plain` and `view --raw` are fast and useful.
The critical gap is **writes**: descriptions, comments, custom fields, attachments, and
transitions all require falling back to curl with manual ADF construction. A custom CLI
wrapping the REST API v3 would unify reads and writes behind a single, agent-friendly interface.

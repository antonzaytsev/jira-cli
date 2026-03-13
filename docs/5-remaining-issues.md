# Jira CLI — Remaining Issues (as of v1.8.7)

## Bug: Summary-only edit sends empty description

`jira issue edit KEY -s"New title"` (without `-b` or `--body-adf`) fails with
"Operation value must be an Atlassian Document". The CLI includes an empty description
value in the PUT payload when no body flag is provided.

### Reproduction

```bash
# FAILS — summary-only, no body flag
jira issue edit PO-1234 -s"New title"
# Error: description: Operation value must be an Atlassian Document

# FAILS — same with echo pipe
echo "" | jira issue edit PO-1234 -s"New title"
# Error: description: Operation value must be an Atlassian Document

# WORKS — providing an explicit body alongside summary
jira issue edit PO-1234 -s"New title" -b "Some body"
# ✓ Issue updated
```

### Expected behavior

When no body flag is given, omit the `description` field from the PUT payload entirely.
Only include `description` when `-b`, `--body-adf`, or `--template` is explicitly provided.

### Workaround

Always include a body flag when editing. To change only the summary without touching
the description, there is currently no clean workaround via the CLI — use the REST API:

```bash
curl -s -u "$ATLASSIAN_EMAIL:$JIRA_API_TOKEN" \
  -X PUT "https://$ATLASSIAN_URL/rest/api/3/issue/PO-1234" \
  -H "Content-Type: application/json" \
  -d '{"fields":{"summary":"New title"}}'
```

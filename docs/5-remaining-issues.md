# Jira CLI — Remaining Issues (as of v1.8.5)

## Bug: `create` uses API v2 instead of v3

`issue create` sends `POST /rest/api/2/issue`. API v2 expects description as a plain string,
so `--body-adf` fails with "Operation value must be a string". The `edit` command correctly
uses v3. The `create` endpoint was not switched.

```
$ jira issue create -p PO -t Task -s "Test" --body-adf '{"type":"doc",...}' --raw
# POST /rest/api/2/issue HTTP/1.1
# Error: description: Operation value must be a string
```

**Workaround**: Create without body, then edit to add description:
```bash
jira issue create -p PO -t Task -s "Summary" --raw   # returns {"key":"PO-1234"}
jira issue edit PO-1234 --body-adf '{"type":"doc",...}'
```

## Bug: `bulk edit` sends empty description

`bulk edit` sends the description field even when not provided in input JSON. Since the edit
endpoint uses v3, the empty/null description value is rejected as invalid ADF.

```
$ echo '[{"key":"PO-1234","summary":"New title"}]' | jira bulk edit
# Error: description: Operation value must be an Atlassian Document
```

`bulk create` and `bulk transition` work correctly.

## Not applicable: Webhook management on Cloud

`webhook list/create/delete` commands exist but Jira Cloud webhooks require OAuth 2.0 / Connect
app authentication — they cannot be managed via API token. The CLI correctly documents this
limitation. Not a bug, just an architectural constraint of Jira Cloud.

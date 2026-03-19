package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetIssueWatchers(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1/watchers", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{
				"isWatching": true,
				"watchCount": 2,
				"watchers": [
					{"accountId": "abc123", "displayName": "User A", "active": true},
					{"accountId": "def456", "displayName": "User B", "active": true}
				]
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetIssueWatchers("TEST-1")
	assert.NoError(t, err)
	assert.True(t, actual.IsWatching)
	assert.Equal(t, 2, actual.WatchCount)
	assert.Len(t, actual.Watchers, 2)
	assert.Equal(t, "User A", actual.Watchers[0].DisplayName)
	assert.Equal(t, "abc123", actual.Watchers[0].AccountID)

	unexpectedStatusCode = true
	_, err = client.GetIssueWatchers("TEST-1")
	assert.Error(t, err)
}

func TestRemoveIssueWatcher(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/rest/api/3/issue/TEST-1/watchers", r.URL.Path)
		assert.Equal(t, "abc123", r.URL.Query().Get("accountId"))

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(204)
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.RemoveIssueWatcher("TEST-1", "abc123")
	assert.NoError(t, err)

	unexpectedStatusCode = true
	err = client.RemoveIssueWatcher("TEST-1", "abc123")
	assert.Error(t, err)
}

func TestGetIssueWorklog(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1/worklog", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{
				"startAt": 0, "maxResults": 50, "total": 1,
				"worklogs": [{
					"id": "10001",
					"author": {"displayName": "User A"},
					"timeSpent": "2h",
					"timeSpentSeconds": 7200,
					"started": "2024-01-01T09:00:00.000+0000",
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000"
				}]
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetIssueWorklog("TEST-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, actual.Total)
	assert.Len(t, actual.Worklogs, 1)
	assert.Equal(t, "2h", actual.Worklogs[0].TimeSpent)
	assert.Equal(t, 7200, actual.Worklogs[0].TimeSpentSeconds)
	assert.Equal(t, "10001", actual.Worklogs[0].ID)

	unexpectedStatusCode = true
	_, err = client.GetIssueWorklog("TEST-1")
	assert.Error(t, err)
}

func TestGetIssueChangelog(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1/changelog", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{
				"startAt": 0, "maxResults": 50, "total": 1,
				"values": [{
					"id": "100",
					"author": {"displayName": "User A"},
					"created": "2024-01-01T10:00:00.000+0000",
					"items": [{
						"field": "status",
						"fieldtype": "jira",
						"fromString": "To Do",
						"toString": "In Progress"
					}]
				}]
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetIssueChangelog("TEST-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, actual.Total)
	assert.Len(t, actual.Histories, 1)
	assert.Equal(t, "100", actual.Histories[0].ID)
	assert.Equal(t, "status", actual.Histories[0].Items[0].Field)
	assert.Equal(t, "To Do", actual.Histories[0].Items[0].FromString)
	assert.Equal(t, "In Progress", actual.Histories[0].Items[0].ToString)

	unexpectedStatusCode = true
	_, err = client.GetIssueChangelog("TEST-1")
	assert.Error(t, err)
}

func TestGetAttachment(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/attachment/12345", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{
				"id": "12345",
				"filename": "test.png",
				"size": 1024,
				"mimeType": "image/png",
				"content": "https://example.com/test.png",
				"created": "2024-01-01T10:00:00.000+0000",
				"author": {"displayName": "User A"}
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetAttachment("12345")
	assert.NoError(t, err)
	assert.Equal(t, "12345", actual.ID)
	assert.Equal(t, "test.png", actual.Filename)
	assert.Equal(t, int64(1024), actual.Size)
	assert.Equal(t, "image/png", actual.MimeType)

	unexpectedStatusCode = true
	_, err = client.GetAttachment("12345")
	assert.Error(t, err)
}

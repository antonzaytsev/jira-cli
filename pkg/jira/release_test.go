package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelease(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/project/TEST/versions", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`[{"id": "10000", "name": "v1.0", "released": true, "archived": false}]`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.Release("TEST")
	assert.NoError(t, err)
	assert.Len(t, actual, 1)
	assert.Equal(t, "v1.0", actual[0].Name)
	assert.True(t, actual[0].Released)

	unexpectedStatusCode = true
	_, err = client.Release("TEST")
	assert.Error(t, err)
}

func TestCreateVersion(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/rest/api/3/version", r.URL.Path)

		body := new(strings.Builder)
		_, _ = io.Copy(body, r.Body)
		assert.Contains(t, body.String(), `"name":"v2.0"`)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			_, _ = w.Write([]byte(`{"id": "10001", "name": "v2.0", "released": false, "archived": false}`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.CreateVersion(&CreateVersionRequest{
		Name:      "v2.0",
		ProjectID: 10000,
	})
	assert.NoError(t, err)
	assert.Equal(t, "v2.0", actual.Name)
	assert.Equal(t, "10001", actual.ID)

	unexpectedStatusCode = true
	_, err = client.CreateVersion(&CreateVersionRequest{Name: "v2.0", ProjectID: 10000})
	assert.Error(t, err)
}

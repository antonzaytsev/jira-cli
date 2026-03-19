package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectComponents(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/project/TEST/components", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`[
				{"id": "10000", "name": "Backend", "description": "Backend services"},
				{"id": "10001", "name": "Frontend", "description": "Frontend app"}
			]`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetProjectComponents("TEST")
	assert.NoError(t, err)
	assert.Len(t, actual, 2)
	assert.Equal(t, "Backend", actual[0].Name)
	assert.Equal(t, "10000", actual[0].ID)
	assert.Equal(t, "Frontend", actual[1].Name)

	unexpectedStatusCode = true
	_, err = client.GetProjectComponents("TEST")
	assert.Error(t, err)
}

package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Release fetches response from /project/{projectIdOrKey}/version endpoint.
func (c *Client) Release(project string) ([]*ProjectVersion, error) {
	path := fmt.Sprintf("/project/%s/versions", project)
	res, err := c.Get(context.Background(), path, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out []*ProjectVersion

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// CreateVersionRequest holds the request data for creating a version.
type CreateVersionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProjectID   int    `json:"projectId"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	StartDate   string `json:"startDate,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// CreateVersion creates a new version using POST /version endpoint.
func (c *Client) CreateVersion(req *CreateVersionRequest) (*ProjectVersion, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := c.Post(context.Background(), "/version", body, Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		return nil, formatUnexpectedResponse(res)
	}

	var out ProjectVersion
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

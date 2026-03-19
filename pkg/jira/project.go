package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// ProjectTypeClassic is a classic project type.
	ProjectTypeClassic = "classic"
	// ProjectTypeNextGen is a next gen project type.
	ProjectTypeNextGen = "next-gen"
)

// Project fetches response from /project endpoint.
func (c *Client) Project() ([]*Project, error) {
	res, err := c.GetV2(context.Background(), "/project?expand=lead", nil)
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

	var out []*Project

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}

// Component holds project component info.
type Component struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Lead         *User  `json:"lead,omitempty"`
	AssigneeType string `json:"assigneeType"`
}

// GetProjectComponents fetches components for a project using GET /project/{key}/components endpoint.
func (c *Client) GetProjectComponents(project string) ([]*Component, error) {
	path := fmt.Sprintf("/project/%s/components", project)
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

	var out []*Component
	err = json.NewDecoder(res.Body).Decode(&out)
	return out, err
}

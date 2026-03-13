package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const webhookBasePath = "/rest/webhooks/1.0/webhook"

// Webhook holds webhook info returned by Jira Server/DC.
type Webhook struct {
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Events          []string `json:"events"`
	JqlFilter       string   `json:"jqlFilter,omitempty"`
	ExcludeBody     bool     `json:"excludeBody,omitempty"`
	Enabled         bool     `json:"enabled,omitempty"`
	Self            string   `json:"self,omitempty"`
	RegisteredByApp string   `json:"registeredByApp,omitempty"`
}

// WebhookCreateRequest holds data for creating a webhook.
type WebhookCreateRequest struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	JqlFilter   string   `json:"jqlFilter,omitempty"`
	ExcludeBody bool     `json:"excludeBody,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// ListWebhooks fetches all webhooks (Jira Server/DC only: GET /rest/webhooks/1.0/webhook).
func (c *Client) ListWebhooks() ([]*Webhook, error) {
	res, err := c.request(ctx(), http.MethodGet, c.server+webhookBasePath, nil, Header{
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

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out []*Webhook
	err = json.NewDecoder(res.Body).Decode(&out)
	return out, err
}

// CreateWebhook registers a new webhook (Jira Server/DC only: POST /rest/webhooks/1.0/webhook).
func (c *Client) CreateWebhook(req *WebhookCreateRequest) (*Webhook, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := c.request(ctx(), http.MethodPost, c.server+webhookBasePath, body, Header{
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

	var out Webhook
	err = json.NewDecoder(res.Body).Decode(&out)
	return &out, err
}

// DeleteWebhook removes a webhook by ID (Jira Server/DC only: DELETE /rest/webhooks/1.0/webhook/{id}).
func (c *Client) DeleteWebhook(id string) error {
	path := fmt.Sprintf("%s/%s", webhookBasePath, id)
	res, err := c.request(ctx(), http.MethodDelete, c.server+path, nil, Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	})
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return formatUnexpectedResponse(res)
	}
	return nil
}

func ctx() context.Context {
	return context.Background()
}

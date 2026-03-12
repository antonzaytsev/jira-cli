package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Attachment holds attachment metadata.
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
	Created  string `json:"created"`
	Author   User   `json:"author"`
}

// AddAttachment uploads a file to an issue using POST /issue/{key}/attachments endpoint.
func (c *Client) AddAttachment(key, filePath string) ([]*Attachment, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	path := fmt.Sprintf("/issue/%s/attachments", key)
	res, err := c.PostMultipart(context.Background(), path, buf.Bytes(), Header{
		"Accept":            "application/json",
		"Content-Type":      writer.FormDataContentType(),
		"X-Atlassian-Token": "no-check",
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

	var out []*Attachment
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteAttachment deletes an attachment using DELETE /attachment/{id} endpoint.
func (c *Client) DeleteAttachment(attachmentID string) error {
	path := fmt.Sprintf("/attachment/%s", attachmentID)
	res, err := c.Delete(context.Background(), path, Header{
		"Accept": "application/json",
	})
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNoContent {
		return formatUnexpectedResponse(res)
	}
	return nil
}

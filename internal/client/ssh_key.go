package client

import (
	"context"
	"fmt"
)

// SshKey represents an SSH key from the API
type SshKey struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	UserID      int    `json:"user_id"`
}

// CreateSshKeyRequest represents a request to create an SSH key
type CreateSshKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type createSshKeyResponse struct {
	Message string `json:"message"`
	Key     SshKey `json:"key"`
}

type showSshKeyResponse struct {
	Key SshKey `json:"key"`
}

type listSshKeysResponse struct {
	Data       []SshKey   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// CreateSshKey creates a new SSH key
func (c *Client) CreateSshKey(ctx context.Context, req CreateSshKeyRequest) (*SshKey, error) {
	var resp createSshKeyResponse
	if err := c.doRequest(ctx, "POST", "/ssh-keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

// GetSshKey retrieves an SSH key by ID
func (c *Client) GetSshKey(ctx context.Context, id string) (*SshKey, error) {
	var resp showSshKeyResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/ssh-keys/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Key, nil
}

// ListSshKeys retrieves all SSH keys
func (c *Client) ListSshKeys(ctx context.Context) ([]SshKey, error) {
	var resp listSshKeysResponse
	if err := c.doRequest(ctx, "GET", "/ssh-keys", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// DeleteSshKey deletes an SSH key
func (c *Client) DeleteSshKey(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/ssh-keys/%s", id), nil, nil)
}

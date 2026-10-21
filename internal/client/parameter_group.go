package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ParameterValues map[string]interface{}

func (p *ParameterValues) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("[]")) || bytes.Equal(data, []byte("null")) {
		*p = ParameterValues{}
		return nil
	}
	return json.Unmarshal(data, (*map[string]interface{})(p))
}

// ParameterGroup represents a cache/database/queue parameter group.
type ParameterGroup struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	ProviderType     string          `json:"provider_type"`
	Family           *string         `json:"family"`
	Description      *string         `json:"description"`
	Parameters       ParameterValues `json:"parameters"`
	LockedParameters []string        `json:"locked_parameters"`
	TeamID           *int            `json:"team_id"`
	IsDefault        bool            `json:"is_default"`
	IsActive         bool            `json:"is_active"`
	IsSystem         bool            `json:"is_system"`
	CreatedAt        *string         `json:"created_at"`
	UpdatedAt        *string         `json:"updated_at"`
}

// CreateParameterGroupRequest is the payload for creating a parameter group.
type CreateParameterGroupRequest struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ProviderType     string                 `json:"provider_type"`
	Family           *string                `json:"family,omitempty"`
	Description      *string                `json:"description,omitempty"`
	Parameters       map[string]interface{} `json:"parameters"`
	LockedParameters []string               `json:"locked_parameters,omitempty"`
	IsDefault        *bool                  `json:"is_default,omitempty"`
}

// UpdateParameterGroupRequest is the payload for updating a parameter group.
// Nil pointers mean "don't update"; non-nil means "set to this value".
type UpdateParameterGroupRequest struct {
	Name             *string                 `json:"name,omitempty"`
	Description      *string                 `json:"description,omitempty"`
	Parameters       *map[string]interface{} `json:"parameters,omitempty"`
	LockedParameters *[]string               `json:"locked_parameters,omitempty"`
	IsDefault        *bool                   `json:"is_default,omitempty"`
	IsActive         *bool                   `json:"is_active,omitempty"`
}

// ListParameterGroupsOptions filters parameter group listings.
type ListParameterGroupsOptions struct {
	Type         string
	ProviderType string
}

type parameterGroupResponse struct {
	Message        string         `json:"message"`
	ParameterGroup ParameterGroup `json:"parameter_group"`
}

type listParameterGroupsResponse struct {
	Data       []ParameterGroup `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// CreateParameterGroup creates a new parameter group.
func (c *Client) CreateParameterGroup(ctx context.Context, req CreateParameterGroupRequest) (*ParameterGroup, error) {
	var resp parameterGroupResponse
	if err := c.doRequest(ctx, "POST", "/parameter-groups", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ParameterGroup, nil
}

// GetParameterGroup retrieves a parameter group by ID.
func (c *Client) GetParameterGroup(ctx context.Context, id string) (*ParameterGroup, error) {
	var resp parameterGroupResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/parameter-groups/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.ParameterGroup, nil
}

// ListParameterGroups returns all parameter groups matching the filters, handling pagination.
func (c *Client) ListParameterGroups(ctx context.Context, opts ListParameterGroupsOptions) ([]ParameterGroup, error) {
	var all []ParameterGroup
	page := 1

	for {
		q := url.Values{}
		q.Set("page", fmt.Sprintf("%d", page))
		if opts.Type != "" {
			q.Set("type", opts.Type)
		}
		if opts.ProviderType != "" {
			q.Set("provider_type", opts.ProviderType)
		}

		var resp listParameterGroupsResponse
		path := "/parameter-groups"
		if encoded := q.Encode(); encoded != "" {
			path = path + "?" + encoded
		}
		if err := c.doRequest(ctx, "GET", path, nil, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return all, nil
}

// UpdateParameterGroup updates a parameter group in place.
func (c *Client) UpdateParameterGroup(ctx context.Context, id string, req UpdateParameterGroupRequest) (*ParameterGroup, error) {
	var resp parameterGroupResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/parameter-groups/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.ParameterGroup, nil
}

// DeleteParameterGroup deletes a parameter group.
func (c *Client) DeleteParameterGroup(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/parameter-groups/%s", id), nil, nil)
}

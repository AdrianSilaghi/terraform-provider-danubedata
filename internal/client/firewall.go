package client

import (
	"context"
	"fmt"
)

// Firewall represents a firewall from the API
type Firewall struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Status        string         `json:"status"`
	IsDefault     bool           `json:"is_default"`
	DefaultAction string         `json:"default_action"`
	Rules         []FirewallRule `json:"rules"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	TeamID        int            `json:"team_id"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Action         string   `json:"action"`
	Direction      string   `json:"direction"`
	Protocol       string   `json:"protocol"`
	PortRangeStart *int     `json:"port_range_start"`
	PortRangeEnd   *int     `json:"port_range_end"`
	SourceIPs      []string `json:"source_ips"`
	Priority       int      `json:"priority"`
}

// CreateFirewallRequest represents a request to create a firewall
type CreateFirewallRequest struct {
	Name          string                     `json:"name"`
	Description   string                     `json:"description,omitempty"`
	IsDefault     bool                       `json:"is_default,omitempty"`
	DefaultAction string                     `json:"default_action,omitempty"`
	Rules         []CreateFirewallRuleRequest `json:"rules,omitempty"`
}

// CreateFirewallRuleRequest represents a rule in a create request
type CreateFirewallRuleRequest struct {
	Name           string   `json:"name,omitempty"`
	Action         string   `json:"action"`
	Direction      string   `json:"direction"`
	Protocol       string   `json:"protocol"`
	PortRangeStart *int     `json:"port_range_start,omitempty"`
	PortRangeEnd   *int     `json:"port_range_end,omitempty"`
	SourceIPs      []string `json:"source_ips,omitempty"`
	Priority       int      `json:"priority,omitempty"`
}

// UpdateFirewallRequest represents a request to update a firewall
type UpdateFirewallRequest struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	IsDefault     *bool  `json:"is_default,omitempty"`
	DefaultAction string `json:"default_action,omitempty"`
}

// AttachFirewallRequest represents a request to attach a firewall to an instance
type AttachFirewallRequest struct {
	InstanceType string `json:"instance_type"`
	InstanceID   string `json:"instance_id"`
}

type createFirewallResponse struct {
	Message  string   `json:"message"`
	Firewall Firewall `json:"firewall"`
}

type showFirewallResponse struct {
	Firewall Firewall `json:"firewall"`
}

type listFirewallsResponse struct {
	Data       []Firewall `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// CreateFirewall creates a new firewall
func (c *Client) CreateFirewall(ctx context.Context, req CreateFirewallRequest) (*Firewall, error) {
	var resp createFirewallResponse
	if err := c.doRequest(ctx, "POST", "/firewalls", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Firewall, nil
}

// GetFirewall retrieves a firewall by ID
func (c *Client) GetFirewall(ctx context.Context, id string) (*Firewall, error) {
	var resp showFirewallResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/firewalls/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Firewall, nil
}

// ListFirewalls retrieves all firewalls (handles pagination automatically)
func (c *Client) ListFirewalls(ctx context.Context) ([]Firewall, error) {
	var allFirewalls []Firewall
	page := 1

	for {
		var resp listFirewallsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/firewalls?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allFirewalls = append(allFirewalls, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allFirewalls, nil
}

// UpdateFirewall updates a firewall
func (c *Client) UpdateFirewall(ctx context.Context, id string, req UpdateFirewallRequest) (*Firewall, error) {
	var resp showFirewallResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/firewalls/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Firewall, nil
}

// DeleteFirewall deletes a firewall
func (c *Client) DeleteFirewall(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/firewalls/%s", id), nil, nil)
}

// AttachFirewall attaches a firewall to an instance
func (c *Client) AttachFirewall(ctx context.Context, firewallID string, req AttachFirewallRequest) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/firewalls/%s/attach", firewallID), req, nil)
}

// DetachFirewall detaches a firewall from an instance
func (c *Client) DetachFirewall(ctx context.Context, firewallID string, req AttachFirewallRequest) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/firewalls/%s/detach", firewallID), req, nil)
}

// DeployFirewall triggers deployment of a firewall
func (c *Client) DeployFirewall(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/firewalls/%s/deploy", id), nil, nil)
}

package client

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// VpsInstance represents a VPS instance from the API
type VpsInstance struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Status            string   `json:"status"`
	StatusLabel       string   `json:"status_label"`
	ResourceProfile   string   `json:"resource_profile"`
	CPUAllocationType string   `json:"cpu_allocation_type"`
	CPUCores          int      `json:"cpu_cores"`
	MemorySizeGB      int      `json:"memory_size_gb"`
	StorageSizeGB     int      `json:"storage_size_gb"`
	Image             string   `json:"image"`
	Datacenter        string   `json:"datacenter"`
	Node              *string  `json:"node"`
	PublicIP          *string  `json:"public_ip"`
	PrivateIP         *string  `json:"private_ip"`
	IPv6Address       *string  `json:"ipv6_address"`
	VNCAccessURL      *string  `json:"vnc_access_url"`
	MonthlyCostCents  int      `json:"monthly_cost_cents"`
	MonthlyCost       float64  `json:"monthly_cost_dollars"`
	DeployedAt        *string  `json:"deployed_at"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	TeamID            int      `json:"team_id"`
	UserID            int      `json:"user_id"`
	SSHKeyID          *string  `json:"ssh_key_id"`
	CanBeStarted      bool     `json:"can_be_started"`
	CanBeStopped      bool     `json:"can_be_stopped"`
	CanBeRebooted     bool     `json:"can_be_rebooted"`
	CanBeDestroyed    bool     `json:"can_be_destroyed"`
}

// CreateVpsRequest represents a request to create a VPS
type CreateVpsRequest struct {
	Name              string  `json:"name"`
	ResourceProfile   string  `json:"resource_profile,omitempty"`
	CPUAllocationType string  `json:"cpu_allocation_type,omitempty"`
	Image             string  `json:"image"`
	Datacenter        string  `json:"datacenter"`
	NetworkStack      string  `json:"network_stack,omitempty"`
	AuthMethod        string  `json:"auth_method"`
	SSHKeyID          *string `json:"ssh_key_id,omitempty"`
	Password          *string `json:"password,omitempty"`
	PasswordConfirm   *string `json:"password_confirmation,omitempty"`
	CustomCloudInit   *string `json:"custom_cloud_init,omitempty"`
	CPUCores          *int    `json:"cpu_cores,omitempty"`
	MemorySizeGB      *int    `json:"memory_size_gb,omitempty"`
	StorageSizeGB     *int    `json:"storage_size_gb,omitempty"`
}

// UpdateVpsRequest represents a request to update a VPS
type UpdateVpsRequest struct {
	ResourceProfile   string  `json:"resource_profile,omitempty"`
	CPUAllocationType string  `json:"cpu_allocation_type,omitempty"`
	CPUCores          *int    `json:"cpu_cores,omitempty"`
	MemorySizeGB      *int    `json:"memory_size_gb,omitempty"`
	StorageSizeGB     *int    `json:"storage_size_gb,omitempty"`
	Password          *string `json:"password,omitempty"`
	PasswordConfirm   *string `json:"password_confirmation,omitempty"`
}

// VpsImage represents an available VPS image
type VpsImage struct {
	ID          string      `json:"id"`
	Image       string      `json:"image"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Distro      string      `json:"distro"`
	Version     interface{} `json:"version"` // Can be string or number from PHP
	Family      *string     `json:"family"`
	DefaultUser string      `json:"default_user"`
}

// GetVersion returns the version as a string regardless of JSON type
func (v *VpsImage) GetVersion() string {
	switch val := v.Version.(type) {
	case string:
		return val
	case float64:
		// JSON numbers are decoded as float64
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

type createVpsResponse struct {
	Message         string      `json:"message"`
	Instance        VpsInstance `json:"instance"`
	FirewallCreated bool        `json:"firewall_created"`
}

type showVpsResponse struct {
	Instance VpsInstance `json:"instance"`
}

type statusVpsResponse struct {
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
}

type listImagesResponse struct {
	Images []VpsImage `json:"images"`
}

type listVpsResponse struct {
	Data       []VpsInstance `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

// CreateVps creates a new VPS instance
func (c *Client) CreateVps(ctx context.Context, req CreateVpsRequest) (*VpsInstance, error) {
	var resp createVpsResponse
	if err := c.doRequest(ctx, "POST", "/vps", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// GetVps retrieves a VPS instance by ID
func (c *Client) GetVps(ctx context.Context, id string) (*VpsInstance, error) {
	var resp showVpsResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/vps/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// UpdateVps updates a VPS instance
func (c *Client) UpdateVps(ctx context.Context, id string, req UpdateVpsRequest) (*VpsInstance, error) {
	var resp showVpsResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/vps/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// DeleteVps deletes a VPS instance
func (c *Client) DeleteVps(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/vps/%s", id), nil, nil)
}

// StartVps starts a VPS instance
func (c *Client) StartVps(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/vps/%s/start", id), nil, nil)
}

// StopVps stops a VPS instance
func (c *Client) StopVps(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/vps/%s/stop", id), nil, nil)
}

// RebootVps reboots a VPS instance
func (c *Client) RebootVps(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/vps/%s/reboot", id), nil, nil)
}

// ReinstallVpsRequest represents a request to reinstall a VPS
type ReinstallVpsRequest struct {
	Image           string  `json:"image"`
	CustomCloudInit *string `json:"custom_cloud_init,omitempty"`
}

// ReinstallVps reinstalls a VPS with a new OS image
func (c *Client) ReinstallVps(ctx context.Context, id string, req ReinstallVpsRequest) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/vps/%s/reinstall", id), req, nil)
}

// GetVpsStatus gets the current status of a VPS instance
func (c *Client) GetVpsStatus(ctx context.Context, id string) (string, error) {
	var resp statusVpsResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/vps/%s/status", id), nil, &resp); err != nil {
		return "", err
	}
	return resp.Status, nil
}

// ListVpsImages lists all available VPS images
func (c *Client) ListVpsImages(ctx context.Context) ([]VpsImage, error) {
	var resp listImagesResponse
	if err := c.doRequest(ctx, "GET", "/vps/images", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Images, nil
}

// ListVps lists all VPS instances (handles pagination automatically)
func (c *Client) ListVps(ctx context.Context) ([]VpsInstance, error) {
	var allInstances []VpsInstance
	page := 1

	for {
		var resp listVpsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/vps?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allInstances = append(allInstances, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allInstances, nil
}

// WaitForVpsStatus waits for a VPS to reach a target status
func (c *Client) WaitForVpsStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately before waiting
	status, err := c.GetVpsStatus(ctx, id)
	if err != nil {
		if IsNotFound(err) && targetStatus == "deleted" {
			return nil
		}
		return fmt.Errorf("error checking VPS status: %w", err)
	}
	status = strings.ToLower(status)
	if status == targetStatus {
		return nil
	}
	if status == "error" {
		return fmt.Errorf("VPS %s entered error state", id)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for VPS %s to reach status %s", id, targetStatus)
			}

			status, err := c.GetVpsStatus(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking VPS status: %w", err)
			}

			status = strings.ToLower(status)
			if status == targetStatus {
				return nil
			}

			if status == "error" {
				return fmt.Errorf("VPS %s entered error state", id)
			}
		}
	}
}

// WaitForVpsDeletion waits for a VPS to be deleted
func (c *Client) WaitForVpsDeletion(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for VPS %s to be deleted", id)
			}

			_, err := c.GetVps(ctx, id)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("error checking VPS: %w", err)
			}
		}
	}
}

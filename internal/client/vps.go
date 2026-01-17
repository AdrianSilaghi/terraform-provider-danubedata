package client

import (
	"context"
	"fmt"
	"time"
)

// VpsInstance represents a VPS instance from the API
type VpsInstance struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	StatusLabel      string   `json:"status_label"`
	ResourceProfile  string   `json:"resource_profile"`
	CPUCores         int      `json:"cpu_cores"`
	MemorySizeGB     int      `json:"memory_size_gb"`
	StorageSizeGB    int      `json:"storage_size_gb"`
	Image            string   `json:"image"`
	Datacenter       string   `json:"datacenter"`
	Node             *string  `json:"node"`
	PublicIP         *string  `json:"public_ip"`
	PrivateIP        *string  `json:"private_ip"`
	IPv6Address      *string  `json:"ipv6_address"`
	VNCAccessURL     *string  `json:"vnc_access_url"`
	MonthlyCostCents int      `json:"monthly_cost_cents"`
	MonthlyCost      float64  `json:"monthly_cost_dollars"`
	DeployedAt       *string  `json:"deployed_at"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	TeamID           int      `json:"team_id"`
	UserID           int      `json:"user_id"`
	SSHKeyID         *string  `json:"ssh_key_id"`
	CanBeStarted     bool     `json:"can_be_started"`
	CanBeStopped     bool     `json:"can_be_stopped"`
	CanBeRebooted    bool     `json:"can_be_rebooted"`
	CanBeDestroyed   bool     `json:"can_be_destroyed"`
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
}

// UpdateVpsRequest represents a request to update a VPS
type UpdateVpsRequest struct {
	ResourceProfile string  `json:"resource_profile,omitempty"`
	Password        *string `json:"password,omitempty"`
	PasswordConfirm *string `json:"password_confirmation,omitempty"`
}

// VpsImage represents an available VPS image
type VpsImage struct {
	ID          string  `json:"id"`
	Image       string  `json:"image"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Distro      string  `json:"distro"`
	Version     string  `json:"version"`
	Family      *string `json:"family"`
	DefaultUser string  `json:"default_user"`
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

// WaitForVpsStatus waits for a VPS to reach a target status
func (c *Client) WaitForVpsStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

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
				// If not found and we're waiting for deletion, that's success
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking VPS status: %w", err)
			}

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

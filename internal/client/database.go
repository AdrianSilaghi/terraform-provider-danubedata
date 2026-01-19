package client

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DatabaseEngine represents the engine object in database instance response
type DatabaseEngine struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DatabaseInstance represents a database instance from the API
type DatabaseInstance struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Status             string         `json:"status"`
	StatusLabel        string         `json:"status_label"`
	ResourceProfile    string         `json:"resource_profile"`
	CPUCores           int            `json:"cpu_cores"`
	MemorySizeMB       int            `json:"memory_size_mb"`
	StorageSizeGB      int            `json:"storage_size_gb"`
	DatabaseName       *string        `json:"database_name"`
	Version            string         `json:"version"`
	Engine             DatabaseEngine `json:"engine"`
	Datacenter         string         `json:"datacenter"`
	Endpoint           *string        `json:"endpoint"`
	Port               *int           `json:"port"`
	Username           *string        `json:"username"`
	ParameterGroupID   *string        `json:"parameter_group_id"`
	MonthlyCostCents   int            `json:"monthly_cost_cents"`
	MonthlyCostDollars float64        `json:"monthly_cost_dollars"`
	DeployedAt         *string        `json:"deployed_at"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	TeamID             int            `json:"team_id"`
	UserID             int            `json:"user_id"`
	CanBeStarted       bool           `json:"can_be_started"`
	CanBeStopped       bool           `json:"can_be_stopped"`
	CanBeDestroyed     bool           `json:"can_be_destroyed"`
}

// CreateDatabaseRequest represents a request to create a database instance
type CreateDatabaseRequest struct {
	Name             string  `json:"name"`
	Provider         string  `json:"provider"` // mysql, postgresql, mariadb
	DatabaseName     string  `json:"database_name,omitempty"`
	Version          string  `json:"version,omitempty"`
	Datacenter       string  `json:"datacenter"`
	ResourceProfile  string  `json:"resource_profile"`
	ParameterGroupID *string `json:"parameter_group_id,omitempty"`
}

// UpdateDatabaseRequest represents a request to update a database instance
type UpdateDatabaseRequest struct {
	Name             string  `json:"name,omitempty"`
	ResourceProfile  string  `json:"resource_profile,omitempty"`
	ParameterGroupID *string `json:"parameter_group_id,omitempty"`
}

type createDatabaseResponse struct {
	Message  string           `json:"message"`
	Instance DatabaseInstance `json:"instance"`
}

type showDatabaseResponse struct {
	Instance       DatabaseInstance `json:"instance"`
	ConnectionInfo string           `json:"connection_info"`
	MonthlyCost    float64          `json:"monthly_cost"`
}

type updateDatabaseResponse struct {
	Message  string           `json:"message"`
	Instance DatabaseInstance `json:"instance"`
}

// DatabaseCredentials represents database connection credentials
type DatabaseCredentials struct {
	ConnectionInfo string `json:"connection_info"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

// CreateDatabase creates a new database instance
func (c *Client) CreateDatabase(ctx context.Context, req CreateDatabaseRequest) (*DatabaseInstance, error) {
	var resp createDatabaseResponse
	if err := c.doRequest(ctx, "POST", "/database", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// GetDatabase retrieves a database instance by ID
func (c *Client) GetDatabase(ctx context.Context, id string) (*DatabaseInstance, error) {
	var resp showDatabaseResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/database/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// UpdateDatabase updates a database instance
func (c *Client) UpdateDatabase(ctx context.Context, id string, req UpdateDatabaseRequest) (*DatabaseInstance, error) {
	var resp updateDatabaseResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/database/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// DeleteDatabase deletes a database instance
func (c *Client) DeleteDatabase(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/database/%s", id), nil, nil)
}

// StopDatabase stops a database instance
func (c *Client) StopDatabase(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/database/%s/stop", id), nil, nil)
}

// StartDatabase starts a database instance
func (c *Client) StartDatabase(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/database/%s/start", id), nil, nil)
}

// GetDatabaseCredentials retrieves credentials for a database instance
func (c *Client) GetDatabaseCredentials(ctx context.Context, id string) (*DatabaseCredentials, error) {
	var resp DatabaseCredentials
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/database/%s/credentials", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WaitForDatabaseStatus waits for a database instance to reach a target status
func (c *Client) WaitForDatabaseStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately before waiting
	instance, err := c.GetDatabase(ctx, id)
	if err != nil {
		if IsNotFound(err) && targetStatus == "deleted" {
			return nil
		}
		return fmt.Errorf("error checking database status: %w", err)
	}
	status := strings.ToLower(instance.Status)
	if status == targetStatus {
		return nil
	}
	if status == "error" {
		return fmt.Errorf("database %s entered error state", id)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for database %s to reach status %s", id, targetStatus)
			}

			instance, err := c.GetDatabase(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking database status: %w", err)
			}

			status := strings.ToLower(instance.Status)
			if status == targetStatus {
				return nil
			}

			if status == "error" {
				return fmt.Errorf("database %s entered error state", id)
			}
		}
	}
}

// WaitForDatabaseDeletion waits for a database instance to be deleted
func (c *Client) WaitForDatabaseDeletion(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for database %s to be deleted", id)
			}

			_, err := c.GetDatabase(ctx, id)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("error checking database: %w", err)
			}
		}
	}
}

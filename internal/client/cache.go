package client

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CacheProvider represents the provider object in cache instance response
type CacheProvider struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CacheInstance represents a cache instance from the API
type CacheInstance struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Status             string        `json:"status"`
	StatusLabel        string        `json:"status_label"`
	ResourceProfile    string        `json:"resource_profile"`
	CPUCores           int           `json:"cpu_cores"`
	MemorySizeMB       int           `json:"memory_size_mb"`
	Version            string        `json:"version"`
	Provider           CacheProvider `json:"provider"`
	Datacenter         string        `json:"datacenter"`
	Endpoint           *string       `json:"endpoint"`
	Port               *int          `json:"port"`
	ParameterGroupID   *string       `json:"parameter_group_id"`
	MonthlyCostCents   int           `json:"monthly_cost_cents"`
	MonthlyCostDollars float64       `json:"monthly_cost_dollars"`
	DeployedAt         *string       `json:"deployed_at"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
	TeamID             int           `json:"team_id"`
	UserID             int           `json:"user_id"`
	CanBeStarted       bool          `json:"can_be_started"`
	CanBeStopped       bool          `json:"can_be_stopped"`
	CanBeDestroyed     bool          `json:"can_be_destroyed"`
}

// Provider represents a cache/database provider
type Provider struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// CreateCacheRequest represents a request to create a cache instance
type CreateCacheRequest struct {
	Name             string  `json:"name"`
	Provider         string  `json:"provider"` // redis, valkey, dragonfly
	MemorySizeMB     int     `json:"memory_size_mb"`
	CPUCores         int     `json:"cpu_cores"`
	Version          string  `json:"version,omitempty"`
	Datacenter       string  `json:"datacenter"`
	ResourceProfile  string  `json:"resource_profile"`
	ParameterGroupID *string `json:"parameter_group_id,omitempty"`
}

// UpdateCacheRequest represents a request to update a cache instance
type UpdateCacheRequest struct {
	Name             string  `json:"name,omitempty"`
	MemorySizeMB     *int    `json:"memory_size_mb,omitempty"`
	CPUCores         *int    `json:"cpu_cores,omitempty"`
	ResourceProfile  string  `json:"resource_profile,omitempty"`
	ParameterGroupID *string `json:"parameter_group_id,omitempty"`
}

type createCacheResponse struct {
	Message  string        `json:"message"`
	Instance CacheInstance `json:"instance"`
}

type showCacheResponse struct {
	Instance       CacheInstance `json:"instance"`
	ConnectionInfo string        `json:"connection_info"`
	MonthlyCost    float64       `json:"monthly_cost"`
}

type updateCacheResponse struct {
	Message  string        `json:"message"`
	Instance CacheInstance `json:"instance"`
}

// CacheConnectionInfo represents cache connection details
type CacheConnectionInfo struct {
	ConnectionInfo string `json:"connection_info"`
	Password       string `json:"password"`
}

type listCacheResponse struct {
	Data       []CacheInstance `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// CreateCache creates a new cache instance
func (c *Client) CreateCache(ctx context.Context, req CreateCacheRequest) (*CacheInstance, error) {
	var resp createCacheResponse
	if err := c.doRequest(ctx, "POST", "/cache", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// GetCache retrieves a cache instance by ID
func (c *Client) GetCache(ctx context.Context, id string) (*CacheInstance, error) {
	var resp showCacheResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/cache/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// UpdateCache updates a cache instance
func (c *Client) UpdateCache(ctx context.Context, id string, req UpdateCacheRequest) (*CacheInstance, error) {
	var resp updateCacheResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/cache/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Instance, nil
}

// DeleteCache deletes a cache instance
func (c *Client) DeleteCache(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/cache/%s", id), nil, nil)
}

// StopCache stops a cache instance
func (c *Client) StopCache(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/cache/%s/stop", id), nil, nil)
}

// StartCache starts a cache instance
func (c *Client) StartCache(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/cache/%s/start", id), nil, nil)
}

// GetCacheConnectionInfo retrieves connection information for a cache instance
func (c *Client) GetCacheConnectionInfo(ctx context.Context, id string) (*CacheConnectionInfo, error) {
	var resp CacheConnectionInfo
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/cache/%s/connection-info", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListCaches lists all cache instances (handles pagination automatically)
func (c *Client) ListCaches(ctx context.Context) ([]CacheInstance, error) {
	var allInstances []CacheInstance
	page := 1

	for {
		var resp listCacheResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/cache?page=%d", page), nil, &resp); err != nil {
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

// WaitForCacheStatus waits for a cache instance to reach a target status
func (c *Client) WaitForCacheStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately before waiting
	instance, err := c.GetCache(ctx, id)
	if err != nil {
		if IsNotFound(err) && targetStatus == "deleted" {
			return nil
		}
		return fmt.Errorf("error checking cache status: %w", err)
	}
	status := strings.ToLower(instance.Status)
	if status == targetStatus {
		return nil
	}
	if status == "error" {
		return fmt.Errorf("cache %s entered error state", id)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for cache %s to reach status %s", id, targetStatus)
			}

			instance, err := c.GetCache(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking cache status: %w", err)
			}

			status := strings.ToLower(instance.Status)
			if status == targetStatus {
				return nil
			}

			if status == "error" {
				return fmt.Errorf("cache %s entered error state", id)
			}
		}
	}
}

// WaitForCacheDeletion waits for a cache instance to be deleted
func (c *Client) WaitForCacheDeletion(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for cache %s to be deleted", id)
			}

			_, err := c.GetCache(ctx, id)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("error checking cache: %w", err)
			}
		}
	}
}

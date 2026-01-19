package client

import (
	"context"
	"fmt"
	"time"
)

// VpsSnapshot represents a VPS snapshot from the API
type VpsSnapshot struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	SizeGB        float64 `json:"size_gb"`
	VpsInstanceID string  `json:"vps_instance_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// CacheSnapshot represents a cache snapshot from the API
type CacheSnapshot struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Status          string  `json:"status"`
	SizeMB          float64 `json:"size_mb"`
	CacheInstanceID string  `json:"cache_instance_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// DatabaseSnapshot represents a database snapshot from the API
type DatabaseSnapshot struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Status             string  `json:"status"`
	SizeGB             float64 `json:"size_gb"`
	DatabaseInstanceID string  `json:"database_instance_id"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// CreateVpsSnapshotRequest represents a request to create a VPS snapshot
type CreateVpsSnapshotRequest struct {
	VpsInstanceID string `json:"vps_instance_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
}

// CreateCacheSnapshotRequest represents a request to create a cache snapshot
type CreateCacheSnapshotRequest struct {
	CacheInstanceID string `json:"cache_instance_id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
}

// CreateDatabaseSnapshotRequest represents a request to create a database snapshot
type CreateDatabaseSnapshotRequest struct {
	DatabaseInstanceID string `json:"database_instance_id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
}

type createVpsSnapshotResponse struct {
	Message  string      `json:"message"`
	Snapshot VpsSnapshot `json:"snapshot"`
}

type createCacheSnapshotResponse struct {
	Message  string        `json:"message"`
	Snapshot CacheSnapshot `json:"snapshot"`
}

type createDatabaseSnapshotResponse struct {
	Message  string           `json:"message"`
	Snapshot DatabaseSnapshot `json:"snapshot"`
}

type listVpsSnapshotsResponse struct {
	Data       []VpsSnapshot `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type listCacheSnapshotsResponse struct {
	Data       []CacheSnapshot `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type listDatabaseSnapshotsResponse struct {
	Data       []DatabaseSnapshot `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

// VPS Snapshot operations

// CreateVpsSnapshot creates a new VPS snapshot
func (c *Client) CreateVpsSnapshot(ctx context.Context, req CreateVpsSnapshotRequest) (*VpsSnapshot, error) {
	var resp createVpsSnapshotResponse
	if err := c.doRequest(ctx, "POST", "/snapshots/vps", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Snapshot, nil
}

// GetVpsSnapshot retrieves a VPS snapshot by ID
func (c *Client) GetVpsSnapshot(ctx context.Context, id string) (*VpsSnapshot, error) {
	// List all snapshots and find the one with matching ID
	snapshots, err := c.ListVpsSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range snapshots {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, &NotFoundError{Resource: "VPS snapshot", ID: id}
}

// ListVpsSnapshots lists all VPS snapshots (handles pagination automatically)
func (c *Client) ListVpsSnapshots(ctx context.Context) ([]VpsSnapshot, error) {
	var allSnapshots []VpsSnapshot
	page := 1

	for {
		var resp listVpsSnapshotsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/snapshots/vps?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allSnapshots = append(allSnapshots, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allSnapshots, nil
}

// RestoreVpsSnapshot restores a VPS snapshot
func (c *Client) RestoreVpsSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/snapshots/vps/%s/restore", id), nil, nil)
}

// DeleteVpsSnapshot deletes a VPS snapshot
func (c *Client) DeleteVpsSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/snapshots/vps/%s", id), nil, nil)
}

// WaitForVpsSnapshotStatus waits for a VPS snapshot to reach a target status
func (c *Client) WaitForVpsSnapshotStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for VPS snapshot %s to reach status %s", id, targetStatus)
			}

			snapshot, err := c.GetVpsSnapshot(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking VPS snapshot status: %w", err)
			}

			if snapshot.Status == targetStatus {
				return nil
			}

			if snapshot.Status == "error" || snapshot.Status == "failed" {
				return fmt.Errorf("VPS snapshot %s entered error state", id)
			}
		}
	}
}

// Cache Snapshot operations

// CreateCacheSnapshot creates a new cache snapshot
func (c *Client) CreateCacheSnapshot(ctx context.Context, req CreateCacheSnapshotRequest) (*CacheSnapshot, error) {
	var resp createCacheSnapshotResponse
	if err := c.doRequest(ctx, "POST", "/snapshots/cache", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Snapshot, nil
}

// GetCacheSnapshot retrieves a cache snapshot by ID
func (c *Client) GetCacheSnapshot(ctx context.Context, id string) (*CacheSnapshot, error) {
	snapshots, err := c.ListCacheSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range snapshots {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, &NotFoundError{Resource: "Cache snapshot", ID: id}
}

// ListCacheSnapshots lists all cache snapshots (handles pagination automatically)
func (c *Client) ListCacheSnapshots(ctx context.Context) ([]CacheSnapshot, error) {
	var allSnapshots []CacheSnapshot
	page := 1

	for {
		var resp listCacheSnapshotsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/snapshots/cache?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allSnapshots = append(allSnapshots, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allSnapshots, nil
}

// RestoreCacheSnapshot restores a cache snapshot
func (c *Client) RestoreCacheSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/snapshots/cache/%s/restore", id), nil, nil)
}

// DeleteCacheSnapshot deletes a cache snapshot
func (c *Client) DeleteCacheSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/snapshots/cache/%s", id), nil, nil)
}

// Database Snapshot operations

// CreateDatabaseSnapshot creates a new database snapshot
func (c *Client) CreateDatabaseSnapshot(ctx context.Context, req CreateDatabaseSnapshotRequest) (*DatabaseSnapshot, error) {
	var resp createDatabaseSnapshotResponse
	if err := c.doRequest(ctx, "POST", "/snapshots/database", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Snapshot, nil
}

// GetDatabaseSnapshot retrieves a database snapshot by ID
func (c *Client) GetDatabaseSnapshot(ctx context.Context, id string) (*DatabaseSnapshot, error) {
	snapshots, err := c.ListDatabaseSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range snapshots {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, &NotFoundError{Resource: "Database snapshot", ID: id}
}

// ListDatabaseSnapshots lists all database snapshots (handles pagination automatically)
func (c *Client) ListDatabaseSnapshots(ctx context.Context) ([]DatabaseSnapshot, error) {
	var allSnapshots []DatabaseSnapshot
	page := 1

	for {
		var resp listDatabaseSnapshotsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/snapshots/database?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allSnapshots = append(allSnapshots, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allSnapshots, nil
}

// RestoreDatabaseSnapshot restores a database snapshot
func (c *Client) RestoreDatabaseSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/snapshots/database/%s/restore", id), nil, nil)
}

// DeleteDatabaseSnapshot deletes a database snapshot
func (c *Client) DeleteDatabaseSnapshot(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/snapshots/database/%s", id), nil, nil)
}

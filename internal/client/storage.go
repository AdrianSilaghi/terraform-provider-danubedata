package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StorageBucket represents a storage bucket from the API
type StorageBucket struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	DisplayName        *string         `json:"display_name"`
	Status             string          `json:"status"`
	StatusLabel        string          `json:"status_label"`
	Region             string          `json:"region"`
	EndpointURL        string          `json:"endpoint_url"`
	PublicURL          *string         `json:"public_url"`
	MinioBucketName    string          `json:"minio_bucket_name"`
	PublicAccess       bool            `json:"public_access"`
	VersioningEnabled  bool            `json:"versioning_enabled"`
	EncryptionEnabled  bool            `json:"encryption_enabled"`
	EncryptionType     *string         `json:"encryption_type"`
	ObjectLockEnabled  bool            `json:"object_lock_enabled"`
	ObjectLockMode     *string         `json:"object_lock_default_retention_mode"`
	ObjectLockDays     *int64          `json:"object_lock_default_retention_days"`
	SizeBytes          int64           `json:"size_bytes"`
	SizeLimitBytes     *int64          `json:"size_limit_bytes"`
	ObjectCount        int             `json:"object_count"`
	Tags               StorageTags     `json:"tags"`
	CorsConfiguration  json.RawMessage `json:"cors_configuration"`
	LifecycleRules     json.RawMessage `json:"lifecycle_rules"`
	MonthlyCostCents   int             `json:"monthly_cost_cents"`
	MonthlyCostDollars float64         `json:"monthly_cost_dollars"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
	TeamID             int             `json:"team_id"`
	UserID             int             `json:"user_id"`
	CanBeModified      bool            `json:"can_be_modified"`
	CanBeDestroyed     bool            `json:"can_be_destroyed"`
}

// StorageTags accepts the empty PHP array emitted for buckets without tags.
type StorageTags map[string]string

func (tags *StorageTags) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" || string(raw) == "[]" {
		*tags = nil
		return nil
	}
	return json.Unmarshal(raw, (*map[string]string)(tags))
}

// CreateStorageBucketRequest represents a request to create a storage bucket
type CreateStorageBucketRequest struct {
	Name               string            `json:"name"`
	DisplayName        string            `json:"display_name,omitempty"`
	Region             string            `json:"region"`
	VersioningEnabled  bool              `json:"versioning_enabled,omitempty"`
	PublicAccess       bool              `json:"public_access,omitempty"`
	EncryptionEnabled  bool              `json:"encryption_enabled"`
	EncryptionType     string            `json:"encryption_type,omitempty"`
	EncryptionKMSKeyID string            `json:"encryption_kms_key_id,omitempty"`
	ObjectLockEnabled  bool              `json:"object_lock_enabled,omitempty"`
	ObjectLockMode     *string           `json:"object_lock_default_retention_mode,omitempty"`
	ObjectLockDays     *int64            `json:"object_lock_default_retention_days,omitempty"`
	CorsConfiguration  json.RawMessage   `json:"cors_configuration,omitempty"`
	LifecycleRules     json.RawMessage   `json:"lifecycle_rules,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
}

// UpdateStorageBucketRequest represents a request to update a storage bucket
type UpdateStorageBucketRequest struct {
	DisplayName        *string            `json:"display_name,omitempty"`
	VersioningEnabled  *bool              `json:"versioning_enabled,omitempty"`
	PublicAccess       *bool              `json:"public_access,omitempty"`
	EncryptionEnabled  *bool              `json:"encryption_enabled,omitempty"`
	EncryptionType     *string            `json:"encryption_type,omitempty"`
	EncryptionKMSKeyID *string            `json:"encryption_kms_key_id,omitempty"`
	ObjectLockMode     **string           `json:"object_lock_default_retention_mode,omitempty"`
	ObjectLockDays     **int64            `json:"object_lock_default_retention_days,omitempty"`
	SizeLimitBytes     **int64            `json:"size_limit_bytes,omitempty"`
	CorsConfiguration  json.RawMessage    `json:"cors_configuration,omitempty"`
	LifecycleRules     json.RawMessage    `json:"lifecycle_rules,omitempty"`
	Tags               *map[string]string `json:"tags,omitempty"`
}

type createStorageBucketResponse struct {
	Message string        `json:"message"`
	Bucket  StorageBucket `json:"bucket"`
}

type showStorageBucketResponse struct {
	Bucket   StorageBucket `json:"bucket"`
	Endpoint string        `json:"endpoint"`
}

type updateStorageBucketResponse struct {
	Message string        `json:"message"`
	Bucket  StorageBucket `json:"bucket"`
}

// StorageAccessKey represents a storage access key from the API
type StorageAccessKey struct {
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	AccessKeyID       string                       `json:"access_key_id"`
	SecretAccessKey   string                       `json:"secret_access_key,omitempty"` // Only returned on creation
	Status            string                       `json:"status"`
	StatusLabel       string                       `json:"status_label"`
	AccessType        string                       `json:"access_type"`
	ARN               *string                      `json:"arn"`
	Scope             string                       `json:"scope"`
	BucketPermissions []StorageKeyBucketPermission `json:"bucket_permissions"`
	IsPrefixScoped    bool                         `json:"is_prefix_scoped"`
	ExpiresAt         *string                      `json:"expires_at"`
	LastUsedAt        *string                      `json:"last_used_at"`
	RevokedAt         *string                      `json:"revoked_at"`
	IsExpired         bool                         `json:"is_expired"`
	CreatedAt         string                       `json:"created_at"`
	UpdatedAt         string                       `json:"updated_at"`
	TeamID            int                          `json:"team_id"`
	UserID            int                          `json:"user_id"`
}

type StorageKeyBucketPermission struct {
	BucketID string `json:"bucket_id"`
	Level    string `json:"level"`
}

// CreateStorageAccessKeyRequest represents a request to create a storage access key
type CreateStorageAccessKeyRequest struct {
	Name              string                       `json:"name"`
	ExpiresAt         *string                      `json:"expires_at,omitempty"`
	Scope             string                       `json:"scope,omitempty"`
	BucketPermissions []StorageKeyBucketPermission `json:"bucket_permissions,omitempty"`
}

// CreateStorageAccessKeyResponse represents the response from creating an access key
// Note: The create response has a different format than the show response
type CreateStorageAccessKeyResponse struct {
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	AccessKeyID       string                       `json:"access_key_id"`
	SecretAccessKey   string                       `json:"secret_access_key"`
	ExpiresAt         *string                      `json:"expires_at"`
	IsPrefixScoped    bool                         `json:"is_prefix_scoped"`
	ARN               *string                      `json:"arn"`
	Scope             string                       `json:"scope"`
	BucketPermissions []StorageKeyBucketPermission `json:"bucket_permissions"`
	Message           string                       `json:"message"`
}

type listStorageBucketsResponse struct {
	Data       []StorageBucket `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type listStorageAccessKeysResponse struct {
	Data       []StorageAccessKey `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

// CreateStorageBucket creates a new storage bucket
func (c *Client) CreateStorageBucket(ctx context.Context, req CreateStorageBucketRequest) (*StorageBucket, error) {
	var resp createStorageBucketResponse
	if err := c.doRequest(ctx, "POST", "/storage/buckets", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// GetStorageBucket retrieves a storage bucket by ID
func (c *Client) GetStorageBucket(ctx context.Context, id string) (*StorageBucket, error) {
	var resp showStorageBucketResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/storage/buckets/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// UpdateStorageBucket updates a storage bucket
func (c *Client) UpdateStorageBucket(ctx context.Context, id string, req UpdateStorageBucketRequest) (*StorageBucket, error) {
	var resp updateStorageBucketResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/storage/buckets/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// DeleteStorageBucket deletes a storage bucket
func (c *Client) DeleteStorageBucket(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/storage/buckets/%s", id), nil, nil)
}

// ListStorageBuckets lists all storage buckets (handles pagination automatically)
func (c *Client) ListStorageBuckets(ctx context.Context) ([]StorageBucket, error) {
	var allBuckets []StorageBucket
	page := 1

	for {
		var resp listStorageBucketsResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/storage/buckets?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allBuckets = append(allBuckets, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allBuckets, nil
}

// WaitForStorageBucketStatus waits for a storage bucket to reach a target status
func (c *Client) WaitForStorageBucketStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Check immediately before waiting
	bucket, err := c.GetStorageBucket(ctx, id)
	if err != nil {
		if IsNotFound(err) && targetStatus == "deleted" {
			return nil
		}
		return fmt.Errorf("error checking storage bucket status: %w", err)
	}
	status := strings.ToLower(bucket.Status)
	if status == targetStatus {
		return nil
	}
	if status == "error" {
		return fmt.Errorf("storage bucket %s entered error state", id)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			bucket, err := c.GetStorageBucket(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking storage bucket status: %w", err)
			}

			status := strings.ToLower(bucket.Status)
			if status == targetStatus {
				return nil
			}

			if status == "error" {
				return fmt.Errorf("storage bucket %s entered error state", id)
			}
		}
	}
}

// WaitForStorageBucketDeletion waits for a storage bucket to be deleted
func (c *Client) WaitForStorageBucketDeletion(ctx context.Context, id string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	tick := time.After(0)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick:
			tick = ticker.C

			_, err := c.GetStorageBucket(ctx, id)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("error checking storage bucket: %w", err)
			}
		}
	}
}

// CreateStorageAccessKey creates a new storage access key
func (c *Client) CreateStorageAccessKey(ctx context.Context, req CreateStorageAccessKeyRequest) (*CreateStorageAccessKeyResponse, error) {
	var resp CreateStorageAccessKeyResponse
	if err := c.doRequest(ctx, "POST", "/storage/access-keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStorageAccessKey retrieves a storage access key by ID
func (c *Client) GetStorageAccessKey(ctx context.Context, id string) (*StorageAccessKey, error) {
	var resp struct {
		AccessKey StorageAccessKey `json:"access_key"`
	}
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/storage/access-keys/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.AccessKey, nil
}

// DeleteStorageAccessKey revokes/deletes a storage access key
func (c *Client) DeleteStorageAccessKey(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/storage/access-keys/%s", id), nil, nil)
}

// ListStorageAccessKeys lists all storage access keys (handles pagination automatically)
func (c *Client) ListStorageAccessKeys(ctx context.Context) ([]StorageAccessKey, error) {
	var allKeys []StorageAccessKey
	page := 1

	for {
		var resp listStorageAccessKeysResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/storage/access-keys?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allKeys = append(allKeys, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allKeys, nil
}

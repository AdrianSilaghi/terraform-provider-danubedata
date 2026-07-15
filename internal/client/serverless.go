package client

import (
	"context"
	"fmt"
	"time"
)

// ServerlessContainer represents a serverless container from the API
type ServerlessContainer struct {
	ID                   string            `json:"id"`
	TeamID               int               `json:"team_id"`
	UserID               int               `json:"user_id"`
	Name                 string            `json:"name"`
	Status               string            `json:"status"`
	ResourceProfile      string            `json:"resource_profile"`
	DeploymentType       string            `json:"deployment_type"`
	SourceType           *string           `json:"source_type"`
	Image                *string           `json:"image"`
	ImageTag             string            `json:"image_tag"`
	RepositoryURL        *string           `json:"repository_url"`
	RepositoryBranch     string            `json:"repository_branch"`
	GitAuthType          string            `json:"git_auth_type"`
	Port                 int               `json:"port"`
	MinScale             int               `json:"min_scale"`
	MaxScale             int               `json:"max_scale"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	URL                  string            `json:"url"`
	CreatedAt            string            `json:"created_at"`
	UpdatedAt            string            `json:"updated_at"`

	// MonthlyCost is not a container column. It is the sibling `monthly_cost`
	// field (current_month_cost_cents / 100) the show endpoint returns
	// alongside `container`; GetServerless copies it on after decoding.
	MonthlyCost float64 `json:"-"`
}

// CreateServerlessRequest represents a request to create a serverless container
type CreateServerlessRequest struct {
	Name                 string            `json:"name"`
	DeploymentType       string            `json:"deployment_type"`
	ResourceProfile      string            `json:"resource_profile,omitempty"`
	Image                string            `json:"image,omitempty"`
	ImageTag             string            `json:"image_tag,omitempty"`
	RepositoryURL        string            `json:"repository_url,omitempty"`
	RepositoryBranch     string            `json:"repository_branch,omitempty"`
	SourceType           string            `json:"source_type,omitempty"`
	GitAuthType          string            `json:"git_auth_type,omitempty"`
	GitCredentials       string            `json:"git_credentials,omitempty"`
	Port                 int               `json:"port,omitempty"`
	MinScale             int               `json:"min_scale,omitempty"`
	MaxScale             int               `json:"max_scale,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// UpdateServerlessRequest represents a request to update a serverless container
type UpdateServerlessRequest struct {
	ResourceProfile      string            `json:"resource_profile,omitempty"`
	Image                string            `json:"image,omitempty"`
	ImageTag             string            `json:"image_tag,omitempty"`
	RepositoryURL        string            `json:"repository_url,omitempty"`
	RepositoryBranch     string            `json:"repository_branch,omitempty"`
	SourceType           string            `json:"source_type,omitempty"`
	GitAuthType          string            `json:"git_auth_type,omitempty"`
	GitCredentials       string            `json:"git_credentials,omitempty"`
	Port                 int               `json:"port,omitempty"`
	MinScale             *int              `json:"min_scale,omitempty"`
	MaxScale             *int              `json:"max_scale,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

type createServerlessResponse struct {
	Message   string              `json:"message"`
	Container ServerlessContainer `json:"container"`
}

type showServerlessResponse struct {
	Container   ServerlessContainer `json:"container"`
	URL         string              `json:"url"`
	MonthlyCost float64             `json:"monthly_cost"`
}

type listServerlessResponse struct {
	Data       []ServerlessContainer `json:"data"`
	Pagination Pagination            `json:"pagination"`
}

// CreateServerless creates a new serverless container
func (c *Client) CreateServerless(ctx context.Context, req CreateServerlessRequest) (*ServerlessContainer, error) {
	var resp createServerlessResponse
	if err := c.doRequest(ctx, "POST", "/serverless", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Container, nil
}

// GetServerless retrieves a serverless container by ID
func (c *Client) GetServerless(ctx context.Context, id string) (*ServerlessContainer, error) {
	var resp showServerlessResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/serverless/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	container := resp.Container
	container.URL = resp.URL
	container.MonthlyCost = resp.MonthlyCost
	return &container, nil
}

// ListServerless retrieves all serverless containers (handles pagination automatically)
func (c *Client) ListServerless(ctx context.Context) ([]ServerlessContainer, error) {
	var allContainers []ServerlessContainer
	page := 1

	for {
		var resp listServerlessResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/serverless?page=%d", page), nil, &resp); err != nil {
			return nil, err
		}
		allContainers = append(allContainers, resp.Data...)

		if page >= resp.Pagination.LastPage || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return allContainers, nil
}

// UpdateServerless updates a serverless container
func (c *Client) UpdateServerless(ctx context.Context, id string, req UpdateServerlessRequest) (*ServerlessContainer, error) {
	var resp showServerlessResponse
	if err := c.doRequest(ctx, "PUT", fmt.Sprintf("/serverless/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Container, nil
}

// DeleteServerless deletes a serverless container
func (c *Client) DeleteServerless(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/serverless/%s", id), nil, nil)
}

// GetServerlessStatus gets the status of a serverless container
func (c *Client) GetServerlessStatus(ctx context.Context, id string) (string, error) {
	container, err := c.GetServerless(ctx, id)
	if err != nil {
		return "", err
	}
	return container.Status, nil
}

// WaitForServerlessStatus waits for a serverless container to reach a target status
func (c *Client) WaitForServerlessStatus(ctx context.Context, id string, targetStatus string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Check immediately
	status, err := c.GetServerlessStatus(ctx, id)
	if err != nil {
		if IsNotFound(err) && targetStatus == "deleted" {
			return nil
		}
		return fmt.Errorf("error checking serverless status: %w", err)
	}
	if status == targetStatus {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for serverless container %s to reach status %s", id, targetStatus)
			}

			status, err := c.GetServerlessStatus(ctx, id)
			if err != nil {
				if IsNotFound(err) && targetStatus == "deleted" {
					return nil
				}
				return fmt.Errorf("error checking serverless status: %w", err)
			}

			if status == targetStatus {
				return nil
			}

			if status == "error" || status == "failed" {
				return fmt.Errorf("serverless container %s entered error state", id)
			}
		}
	}
}

// WaitForServerlessDeletion waits for a serverless container to be deleted
func (c *Client) WaitForServerlessDeletion(ctx context.Context, id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for serverless container %s to be deleted", id)
			}

			_, err := c.GetServerless(ctx, id)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("error checking serverless container: %w", err)
			}
		}
	}
}

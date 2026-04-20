package client

import (
	"context"
	"fmt"
)

// EnableCacheDns enables public DNS for a cache instance.
func (c *Client) EnableCacheDns(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/cache/%s/dns", id), nil, nil)
}

// DisableCacheDns disables public DNS for a cache instance.
func (c *Client) DisableCacheDns(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/cache/%s/dns", id), nil, nil)
}

// EnableDatabaseDns enables public DNS for a database instance.
func (c *Client) EnableDatabaseDns(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/database/%s/dns", id), nil, nil)
}

// DisableDatabaseDns disables public DNS for a database instance.
func (c *Client) DisableDatabaseDns(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/database/%s/dns", id), nil, nil)
}

package client

import (
	"context"
	"fmt"
)

// StaticSite represents a DanubeData static site (pages).
type StaticSite struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// StaticSiteDomainDNSInstructions is the DNS record a domain owner must add to prove
// ownership of a custom domain.
type StaticSiteDomainDNSInstructions struct {
	RecordType   string `json:"record_type"`
	RecordName   string `json:"record_name"`
	RecordValue  string `json:"record_value"`
	Instructions string `json:"instructions"`
}

// StaticSiteDomain represents a domain attached to a static site.
type StaticSiteDomain struct {
	ID                 string                          `json:"id"`
	Domain             string                          `json:"domain"`
	VerificationStatus string                          `json:"verification_status"`
	TLSStatus          string                          `json:"tls_status"`
	DeploymentStatus   string                          `json:"deployment_status"`
	IsPrimary          bool                            `json:"is_primary"`
	DNSInstructions    StaticSiteDomainDNSInstructions `json:"dns_instructions"`
	CreatedAt          string                          `json:"created_at"`
}

// CreateStaticSiteRequest is the payload for creating a static site.
type CreateStaticSiteRequest struct {
	Name string  `json:"name"`
	Plan *string `json:"plan,omitempty"`
}

// AddStaticSiteDomainRequest is the payload for adding a custom domain.
type AddStaticSiteDomainRequest struct {
	Domain string `json:"domain"`
}

type staticSiteResponse struct {
	Message string     `json:"message"`
	Data    StaticSite `json:"data"`
}

type listStaticSitesResponse struct {
	Data       []StaticSite `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

type staticSiteDomainResponse struct {
	Message string           `json:"message"`
	Data    StaticSiteDomain `json:"data"`
}

type listStaticSiteDomainsResponse struct {
	Data []StaticSiteDomain `json:"data"`
}

// CreateStaticSite creates a new static site under the caller's current team.
func (c *Client) CreateStaticSite(ctx context.Context, req CreateStaticSiteRequest) (*StaticSite, error) {
	var resp staticSiteResponse
	if err := c.doRequest(ctx, "POST", "/static-sites", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetStaticSite retrieves a static site by ID.
func (c *Client) GetStaticSite(ctx context.Context, id string) (*StaticSite, error) {
	var resp staticSiteResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/static-sites/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ListStaticSites lists all static sites for a team (handles pagination).
func (c *Client) ListStaticSites(ctx context.Context, teamID int) ([]StaticSite, error) {
	var all []StaticSite
	page := 1
	for {
		var resp listStaticSitesResponse
		if err := c.doRequest(ctx, "GET", fmt.Sprintf("/teams/%d/static-sites?page=%d", teamID, page), nil, &resp); err != nil {
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

// DeleteStaticSite deletes a static site.
func (c *Client) DeleteStaticSite(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/static-sites/%s", id), nil, nil)
}

// ListStaticSiteDomains lists all domains for a static site.
func (c *Client) ListStaticSiteDomains(ctx context.Context, siteID string) ([]StaticSiteDomain, error) {
	var resp listStaticSiteDomainsResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/static-sites/%s/domains", siteID), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// AddStaticSiteDomain adds a custom domain to a static site.
func (c *Client) AddStaticSiteDomain(ctx context.Context, siteID string, req AddStaticSiteDomainRequest) (*StaticSiteDomain, error) {
	var resp staticSiteDomainResponse
	if err := c.doRequest(ctx, "POST", fmt.Sprintf("/static-sites/%s/domains", siteID), req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// DeleteStaticSiteDomain removes a domain from a static site.
func (c *Client) DeleteStaticSiteDomain(ctx context.Context, siteID, domainID string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/static-sites/%s/domains/%s", siteID, domainID), nil, nil)
}

// VerifyStaticSiteDomain triggers verification of a custom domain.
func (c *Client) VerifyStaticSiteDomain(ctx context.Context, siteID, domainID string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/static-sites/%s/domains/%s/verify", siteID, domainID), nil, nil)
}

// FindStaticSiteDomain looks up a domain on a static site by name.
func (c *Client) FindStaticSiteDomain(ctx context.Context, siteID, domain string) (*StaticSiteDomain, error) {
	domains, err := c.ListStaticSiteDomains(ctx, siteID)
	if err != nil {
		return nil, err
	}
	for i := range domains {
		if domains[i].Domain == domain {
			return &domains[i], nil
		}
	}
	return nil, &NotFoundError{Resource: "static site domain", ID: domain}
}

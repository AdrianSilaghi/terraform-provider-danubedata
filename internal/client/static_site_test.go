package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateStaticSite(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/static-sites" {
			t.Errorf("Path = %v, want /static-sites", r.URL.Path)
		}

		var req CreateStaticSiteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Name != "my-site" {
			t.Errorf("Name = %v, want my-site", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(staticSiteResponse{
			Message: "Site created",
			Data: StaticSite{
				ID:     "site-123",
				Name:   "my-site",
				Slug:   "my-site",
				URL:    "https://my-site.pages.danubedata.ro",
				Plan:   "free",
				Status: "pending",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	plan := "free"
	site, err := c.CreateStaticSite(context.Background(), CreateStaticSiteRequest{Name: "my-site", Plan: &plan})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.ID != "site-123" {
		t.Errorf("ID = %v, want site-123", site.ID)
	}
	if site.URL != "https://my-site.pages.danubedata.ro" {
		t.Errorf("URL = %v, want https://my-site.pages.danubedata.ro", site.URL)
	}
}

func TestClient_GetStaticSite(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static-sites/site-123" {
			t.Errorf("Path = %v, want /static-sites/site-123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(staticSiteResponse{
			Data: StaticSite{ID: "site-123", Name: "my-site", URL: "https://my-site.pages.danubedata.ro"},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	site, err := c.GetStaticSite(context.Background(), "site-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.ID != "site-123" {
		t.Errorf("ID = %v, want site-123", site.ID)
	}
}

func TestClient_DeleteStaticSite(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/static-sites/site-123" {
			t.Errorf("Path = %v, want /static-sites/site-123", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.DeleteStaticSite(context.Background(), "site-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AddStaticSiteDomain(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/static-sites/site-123/domains" {
			t.Errorf("Path = %v, want /static-sites/site-123/domains", r.URL.Path)
		}

		var req AddStaticSiteDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Domain != "www.example.com" {
			t.Errorf("Domain = %v, want www.example.com", req.Domain)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(staticSiteDomainResponse{
			Message: "Domain added",
			Data: StaticSiteDomain{
				ID:                 "domain-99",
				Domain:             "www.example.com",
				VerificationStatus: "pending",
				TLSStatus:          "pending",
				DeploymentStatus:   "pending",
				IsPrimary:          false,
				DNSInstructions: StaticSiteDomainDNSInstructions{
					RecordType:   "TXT",
					RecordName:   "_danubedata-verify.www.example.com",
					RecordValue:  "abc123token",
					Instructions: "Add a TXT record with the name '_danubedata-verify.www.example.com' and value 'abc123token' to your DNS provider.",
				},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	d, err := c.AddStaticSiteDomain(context.Background(), "site-123", AddStaticSiteDomainRequest{Domain: "www.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != "domain-99" {
		t.Errorf("ID = %v, want domain-99", d.ID)
	}
	if d.Domain != "www.example.com" {
		t.Errorf("Domain = %v, want www.example.com", d.Domain)
	}
	if d.DNSInstructions.RecordType != "TXT" {
		t.Errorf("DNSInstructions.RecordType = %v, want TXT", d.DNSInstructions.RecordType)
	}
}

func TestClient_ListStaticSiteDomains(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static-sites/site-123/domains" {
			t.Errorf("Path = %v, want /static-sites/site-123/domains", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listStaticSiteDomainsResponse{
			Data: []StaticSiteDomain{
				{ID: "domain-1", Domain: "default.pages.danubedata.ro", VerificationStatus: "verified", IsPrimary: true},
				{ID: "domain-2", Domain: "www.example.com", VerificationStatus: "pending", IsPrimary: false},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	domains, err := c.ListStaticSiteDomains(context.Background(), "site-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("got %d domains, want 2", len(domains))
	}
}

func TestClient_DeleteStaticSiteDomain(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/static-sites/site-123/domains/domain-99" {
			t.Errorf("Path = %v, want /static-sites/site-123/domains/domain-99", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.DeleteStaticSiteDomain(context.Background(), "site-123", "domain-99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

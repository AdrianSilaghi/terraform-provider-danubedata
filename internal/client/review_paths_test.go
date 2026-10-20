package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestListsReadEveryPage(t *testing.T) {
	tests := []struct {
		name string
		path string
		list func(*Client) (int, error)
	}{
		{"VPS", "/vps", func(c *Client) (int, error) { v, e := c.ListVps(context.Background()); return len(v), e }},
		{"cache", "/cache", func(c *Client) (int, error) { v, e := c.ListCaches(context.Background()); return len(v), e }},
		{"database", "/database", func(c *Client) (int, error) { v, e := c.ListDatabases(context.Background()); return len(v), e }},
		{"firewall", "/firewalls", func(c *Client) (int, error) { v, e := c.ListFirewalls(context.Background()); return len(v), e }},
		{"parameter group", "/parameter-groups", func(c *Client) (int, error) {
			v, e := c.ListParameterGroups(context.Background(), ListParameterGroupsOptions{})
			return len(v), e
		}},
		{"serverless", "/serverless", func(c *Client) (int, error) { v, e := c.ListServerless(context.Background()); return len(v), e }},
		{"VPS snapshot", "/snapshots/vps", func(c *Client) (int, error) { v, e := c.ListVpsSnapshots(context.Background()); return len(v), e }},
		{"cache snapshot", "/snapshots/cache", func(c *Client) (int, error) { v, e := c.ListCacheSnapshots(context.Background()); return len(v), e }},
		{"database snapshot", "/snapshots/database", func(c *Client) (int, error) { v, e := c.ListDatabaseSnapshots(context.Background()); return len(v), e }},
		{"SSH key", "/ssh-keys", func(c *Client) (int, error) { v, e := c.ListSshKeys(context.Background()); return len(v), e }},
		{"static site", "/teams/17/static-sites", func(c *Client) (int, error) { v, e := c.ListStaticSites(context.Background(), 17); return len(v), e }},
		{"storage bucket", "/storage/buckets", func(c *Client) (int, error) { v, e := c.ListStorageBuckets(context.Background()); return len(v), e }},
		{"storage access key", "/storage/access-keys", func(c *Client) (int, error) { v, e := c.ListStorageAccessKeys(context.Background()); return len(v), e }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var pages []int
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("path = %q, want %q", r.URL.Path, tc.path)
				}
				page, err := strconv.Atoi(r.URL.Query().Get("page"))
				if err != nil {
					t.Errorf("invalid page: %v", err)
				}
				pages = append(pages, page)
				_, _ = fmt.Fprint(w, `{"data":[{}],"pagination":{"last_page":2}}`)
			})
			defer server.Close()
			count, err := tc.list(newTestClient(server))
			if err != nil || count != 2 || len(pages) != 2 || pages[0] != 1 || pages[1] != 2 {
				t.Fatalf("list returned %d items, pages %v, err %v", count, pages, err)
			}
		})
	}
}

func TestLookupsDistinguishMissingFromDenied(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		body string
		find func(*Client) error
	}{
		{"replica", "/database/db/replicas", `{"replicas":[{"replica_index":1}]}`, func(c *Client) error { _, e := c.FindDatabaseReplica(context.Background(), "db", 2); return e }},
		{"site domain", "/static-sites/site/domains", `{"data":[{"domain":"other.example"}]}`, func(c *Client) error {
			_, e := c.FindStaticSiteDomain(context.Background(), "site", "wanted.example")
			return e
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var denied atomic.Bool
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("path = %q, want %q", r.URL.Path, tc.path)
				}
				if denied.Load() {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				_, _ = fmt.Fprint(w, tc.body)
			})
			defer server.Close()
			c := newTestClient(server)
			if err := tc.find(c); !IsNotFound(err) {
				t.Fatalf("absent resource = %v, want not found", err)
			}
			denied.Store(true)
			if err := tc.find(c); err == nil || IsNotFound(err) {
				t.Fatalf("permission failure became not found: %v", err)
			}
		})
	}
}

func TestClientRejectsMalformedAndDeniedLaterPages(t *testing.T) {
	for _, response := range []struct {
		name string
		code int
		body string
	}{
		{"denied", http.StatusForbidden, `{"message":"denied"}`},
		{"malformed", http.StatusOK, `{"data":`},
	} {
		t.Run(response.name, func(t *testing.T) {
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("page") == "1" {
					_, _ = fmt.Fprint(w, `{"data":[{"id":"first"}],"pagination":{"last_page":2}}`)
					return
				}
				w.WriteHeader(response.code)
				_, _ = fmt.Fprint(w, response.body)
			})
			defer server.Close()
			items, err := newTestClient(server).ListVps(context.Background())
			if err == nil || len(items) != 0 {
				t.Fatalf("partial list escaped after bad later page: %v, %v", items, err)
			}
			var apiErr *APIError
			if response.name == "denied" && !errors.As(err, &apiErr) {
				t.Fatalf("permission failure lost: %v", err)
			}
		})
	}
}

func TestSnapshotReadsFindLaterPage(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		get  func(*Client) (int64, error)
	}{
		{"VPS", "/snapshots/vps", func(c *Client) (int64, error) {
			v, e := c.GetVpsSnapshot(context.Background(), 2)
			if e != nil {
				return 0, e
			}
			return v.ID, nil
		}},
		{"cache", "/snapshots/cache", func(c *Client) (int64, error) {
			v, e := c.GetCacheSnapshot(context.Background(), 2)
			if e != nil {
				return 0, e
			}
			return v.ID, nil
		}},
		{"database", "/snapshots/database", func(c *Client) (int64, error) {
			v, e := c.GetDatabaseSnapshot(context.Background(), 2)
			if e != nil {
				return 0, e
			}
			return v.ID, nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pages []string
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("path = %q, want %q", r.URL.Path, tc.path)
				}
				page := r.URL.Query().Get("page")
				pages = append(pages, page)
				_, _ = fmt.Fprintf(w, `{"data":[{"id":%s,"status":"ready"}],"pagination":{"last_page":2}}`, page)
			})
			defer server.Close()
			id, err := tc.get(newTestClient(server))
			if err != nil || id != 2 || len(pages) != 2 || pages[0] != "1" || pages[1] != "2" {
				t.Fatalf("snapshot = %d, pages %v, err %v", id, pages, err)
			}
		})
	}
}

func TestReplicaAndDomainLookupsReturnMatchingItem(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/database/db/replicas":
			_, _ = fmt.Fprint(w, `{"replicas":[{"replica_index":1,"name":"first"},{"replica_index":2,"name":"second"}]}`)
		case "/static-sites/site/domains":
			_, _ = fmt.Fprint(w, `{"data":[{"domain":"first.example"},{"domain":"second.example"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()
	c := newTestClient(server)
	replica, err := c.FindDatabaseReplica(context.Background(), "db", 2)
	if err != nil || replica.Name != "second" {
		t.Fatalf("matching replica = %+v, %v", replica, err)
	}
	domain, err := c.FindStaticSiteDomain(context.Background(), "site", "second.example")
	if err != nil || domain.Domain != "second.example" {
		t.Fatalf("matching domain = %+v, %v", domain, err)
	}
}

func TestRestoreAndVerifyActionsUseExpectedRoutes(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
		call func(*Client) error
	}{
		{"restore cache snapshot", "/snapshots/cache/23/restore", func(c *Client) error { return c.RestoreCacheSnapshot(context.Background(), 23) }},
		{"restore database snapshot", "/snapshots/database/23/restore", func(c *Client) error { return c.RestoreDatabaseSnapshot(context.Background(), 23) }},
		{"verify site domain", "/static-sites/site/domains/domain/verify", func(c *Client) error { return c.VerifyStaticSiteDomain(context.Background(), "site", "domain") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tc.path {
					t.Errorf("request = %s %s, want POST %s", r.Method, r.URL.Path, tc.path)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			defer server.Close()
			if err := tc.call(newTestClient(server)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

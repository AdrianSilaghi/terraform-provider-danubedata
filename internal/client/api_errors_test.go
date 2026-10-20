package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestAPIPermissionFailuresAreNeverMissingResources(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Project access refused"}`))
	})
	defer server.Close()
	c := newTestClient(server)
	tests := map[string]func(context.Context) error{
		"create VPS":       func(ctx context.Context) error { _, err := c.CreateVps(ctx, CreateVpsRequest{}); return err },
		"read VPS":         func(ctx context.Context) error { _, err := c.GetVps(ctx, "id"); return err },
		"update VPS":       func(ctx context.Context) error { _, err := c.UpdateVps(ctx, "id", UpdateVpsRequest{}); return err },
		"delete VPS":       func(ctx context.Context) error { return c.DeleteVps(ctx, "id") },
		"list VPS":         func(ctx context.Context) error { _, err := c.ListVps(ctx); return err },
		"create cache":     func(ctx context.Context) error { _, err := c.CreateCache(ctx, CreateCacheRequest{}); return err },
		"read cache":       func(ctx context.Context) error { _, err := c.GetCache(ctx, "id"); return err },
		"update cache":     func(ctx context.Context) error { _, err := c.UpdateCache(ctx, "id", UpdateCacheRequest{}); return err },
		"delete cache":     func(ctx context.Context) error { return c.DeleteCache(ctx, "id") },
		"list caches":      func(ctx context.Context) error { _, err := c.ListCaches(ctx); return err },
		"cache connection": func(ctx context.Context) error { _, err := c.GetCacheConnectionInfo(ctx, "id"); return err },
		"create database":  func(ctx context.Context) error { _, err := c.CreateDatabase(ctx, CreateDatabaseRequest{}); return err },
		"read database":    func(ctx context.Context) error { _, err := c.GetDatabase(ctx, "id"); return err },
		"update database": func(ctx context.Context) error {
			_, err := c.UpdateDatabase(ctx, "id", UpdateDatabaseRequest{})
			return err
		},
		"delete database":      func(ctx context.Context) error { return c.DeleteDatabase(ctx, "id") },
		"list databases":       func(ctx context.Context) error { _, err := c.ListDatabases(ctx); return err },
		"database credentials": func(ctx context.Context) error { _, err := c.GetDatabaseCredentials(ctx, "id"); return err },
		"create replicas": func(ctx context.Context) error {
			_, err := c.AddDatabaseReplicas(ctx, "id", AddDatabaseReplicasRequest{})
			return err
		},
		"read replicas": func(ctx context.Context) error { _, err := c.ListDatabaseReplicas(ctx, "id"); return err },
		"create parameters": func(ctx context.Context) error {
			_, err := c.CreateParameterGroup(ctx, CreateParameterGroupRequest{})
			return err
		},
		"read parameters": func(ctx context.Context) error { _, err := c.GetParameterGroup(ctx, "id"); return err },
		"update parameters": func(ctx context.Context) error {
			_, err := c.UpdateParameterGroup(ctx, "id", UpdateParameterGroupRequest{})
			return err
		},
		"list parameters": func(ctx context.Context) error {
			_, err := c.ListParameterGroups(ctx, ListParameterGroupsOptions{})
			return err
		},
		"create firewall": func(ctx context.Context) error { _, err := c.CreateFirewall(ctx, CreateFirewallRequest{}); return err },
		"read firewall":   func(ctx context.Context) error { _, err := c.GetFirewall(ctx, "id"); return err },
		"update firewall": func(ctx context.Context) error {
			_, err := c.UpdateFirewall(ctx, "id", UpdateFirewallRequest{})
			return err
		},
		"list firewalls": func(ctx context.Context) error { _, err := c.ListFirewalls(ctx); return err },
		"create serverless": func(ctx context.Context) error {
			_, err := c.CreateServerless(ctx, CreateServerlessRequest{})
			return err
		},
		"read serverless": func(ctx context.Context) error { _, err := c.GetServerless(ctx, "id"); return err },
		"update serverless": func(ctx context.Context) error {
			_, err := c.UpdateServerless(ctx, "id", UpdateServerlessRequest{})
			return err
		},
		"list serverless": func(ctx context.Context) error { _, err := c.ListServerless(ctx); return err },
		"create bucket": func(ctx context.Context) error {
			_, err := c.CreateStorageBucket(ctx, CreateStorageBucketRequest{})
			return err
		},
		"read bucket": func(ctx context.Context) error { _, err := c.GetStorageBucket(ctx, "id"); return err },
		"update bucket": func(ctx context.Context) error {
			_, err := c.UpdateStorageBucket(ctx, "id", UpdateStorageBucketRequest{})
			return err
		},
		"list buckets": func(ctx context.Context) error { _, err := c.ListStorageBuckets(ctx); return err },
		"create storage key": func(ctx context.Context) error {
			_, err := c.CreateStorageAccessKey(ctx, CreateStorageAccessKeyRequest{})
			return err
		},
		"read storage key":  func(ctx context.Context) error { _, err := c.GetStorageAccessKey(ctx, "id"); return err },
		"list storage keys": func(ctx context.Context) error { _, err := c.ListStorageAccessKeys(ctx); return err },
		"create SSH key":    func(ctx context.Context) error { _, err := c.CreateSshKey(ctx, CreateSshKeyRequest{}); return err },
		"read SSH key":      func(ctx context.Context) error { _, err := c.GetSshKey(ctx, "1"); return err },
		"list SSH keys":     func(ctx context.Context) error { _, err := c.ListSshKeys(ctx); return err },
		"create site": func(ctx context.Context) error {
			_, err := c.CreateStaticSite(ctx, CreateStaticSiteRequest{})
			return err
		},
		"read site":  func(ctx context.Context) error { _, err := c.GetStaticSite(ctx, "id"); return err },
		"list sites": func(ctx context.Context) error { _, err := c.ListStaticSites(ctx, 17); return err },
		"create domain": func(ctx context.Context) error {
			_, err := c.AddStaticSiteDomain(ctx, "id", AddStaticSiteDomainRequest{})
			return err
		},
		"list domains": func(ctx context.Context) error { _, err := c.ListStaticSiteDomains(ctx, "id"); return err },
		"create VPS snapshot": func(ctx context.Context) error {
			_, err := c.CreateVpsSnapshot(ctx, CreateVpsSnapshotRequest{})
			return err
		},
		"read VPS snapshot": func(ctx context.Context) error { _, err := c.GetVpsSnapshot(ctx, 1); return err },
		"create cache snapshot": func(ctx context.Context) error {
			_, err := c.CreateCacheSnapshot(ctx, CreateCacheSnapshotRequest{})
			return err
		},
		"read cache snapshot": func(ctx context.Context) error { _, err := c.GetCacheSnapshot(ctx, 1); return err },
		"create database snapshot": func(ctx context.Context) error {
			_, err := c.CreateDatabaseSnapshot(ctx, CreateDatabaseSnapshotRequest{})
			return err
		},
		"read database snapshot": func(ctx context.Context) error { _, err := c.GetDatabaseSnapshot(ctx, 1); return err },
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			err := run(context.Background())
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
				t.Fatalf("permission failure lost: %v", err)
			}
			if IsNotFound(err) {
				t.Fatal("permission failure reported as a missing resource")
			}
		})
	}
}

func TestCancelProvisioningWaits(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Error("cancelled operation reached the API")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()
	c := newTestClient(server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := map[string]func() error{
		"VPS":                 func() error { return c.WaitForVpsStatus(ctx, "id", "running", time.Hour) },
		"VPS deletion":        func() error { return c.WaitForVpsDeletion(ctx, "id", time.Hour) },
		"cache":               func() error { return c.WaitForCacheStatus(ctx, "id", "running", time.Hour) },
		"cache deletion":      func() error { return c.WaitForCacheDeletion(ctx, "id", time.Hour) },
		"database":            func() error { return c.WaitForDatabaseStatus(ctx, "id", "running", time.Hour) },
		"database deletion":   func() error { return c.WaitForDatabaseDeletion(ctx, "id", time.Hour) },
		"replica":             func() error { return c.WaitForDatabaseReplicaReady(ctx, "id", 1, time.Hour) },
		"bucket":              func() error { return c.WaitForStorageBucketStatus(ctx, "id", "ready", time.Hour) },
		"bucket deletion":     func() error { return c.WaitForStorageBucketDeletion(ctx, "id", time.Hour) },
		"serverless":          func() error { return c.WaitForServerlessStatus(ctx, "id", "ready", time.Hour) },
		"serverless deletion": func() error { return c.WaitForServerlessDeletion(ctx, "id", time.Hour) },
		"serverless rollout":  func() error { return c.WaitForServerlessGeneration(ctx, "id", 1) },
		"VPS snapshot":        func() error { return c.WaitForVpsSnapshotStatus(ctx, 1, "ready", time.Hour) },
		"cache snapshot":      func() error { return c.WaitForCacheSnapshotStatus(ctx, 1, "ready", time.Hour) },
		"database snapshot":   func() error { return c.WaitForDatabaseSnapshotStatus(ctx, 1, "ready", time.Hour) },
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if err := run(); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
		})
	}
}

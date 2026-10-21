package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func waitStateBody(status string) string {
	return fmt.Sprintf(`{"status":%q,"instance":{"status":%q},"container":{"status":%q},"bucket":{"status":%q}}`, status, status, status, status)
}

func snapshotStateBody(status string) string {
	return fmt.Sprintf(`{"data":[{"id":1,"status":%q}],"pagination":{"last_page":1}}`, status)
}

func TestWaitersObserveNextPoll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		first  string
		second string
		wait   func(*Client) error
	}{
		{"VPS", waitStateBody("provisioning"), waitStateBody("running"), func(c *Client) error {
			return c.WaitForVpsStatus(context.Background(), "id", "running", 20*time.Second)
		}},
		{"cache", waitStateBody("provisioning"), waitStateBody("running"), func(c *Client) error {
			return c.WaitForCacheStatus(context.Background(), "id", "running", 20*time.Second)
		}},
		{"database", waitStateBody("provisioning"), waitStateBody("running"), func(c *Client) error {
			return c.WaitForDatabaseStatus(context.Background(), "id", "running", 20*time.Second)
		}},
		{"bucket", waitStateBody("creating"), waitStateBody("ready"), func(c *Client) error {
			return c.WaitForStorageBucketStatus(context.Background(), "id", "ready", 20*time.Second)
		}},
		{"serverless", waitStateBody("creating"), waitStateBody("ready"), func(c *Client) error {
			return c.WaitForServerlessStatus(context.Background(), "id", "ready", 20*time.Second)
		}},
		{"VPS snapshot", snapshotStateBody("creating"), snapshotStateBody("ready"), func(c *Client) error {
			return c.WaitForVpsSnapshotStatus(context.Background(), 1, "ready", 20*time.Second)
		}},
		{"cache snapshot", snapshotStateBody("creating"), snapshotStateBody("ready"), func(c *Client) error {
			return c.WaitForCacheSnapshotStatus(context.Background(), 1, "ready", 20*time.Second)
		}},
		{"database snapshot", snapshotStateBody("creating"), snapshotStateBody("ready"), func(c *Client) error {
			return c.WaitForDatabaseSnapshotStatus(context.Background(), 1, "ready", 20*time.Second)
		}},
		{"replica", `{"replicas":[{"replica_index":1,"status":"creating","ready":false}]}`, `{"replicas":[{"replica_index":1,"status":"running","ready":true}]}`, func(c *Client) error {
			return c.WaitForDatabaseReplicaReady(context.Background(), "id", 1, 20*time.Second)
		}},
	}
	var wg sync.WaitGroup
	for _, tc := range tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var calls atomic.Int32
			server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
				body := tc.first
				if calls.Add(1) > 1 {
					body = tc.second
				}
				_, _ = fmt.Fprint(w, body)
			})
			defer server.Close()
			if err := tc.wait(newTestClient(server)); err != nil {
				t.Errorf("%s: %v", tc.name, err)
				return
			}
			if calls.Load() != 2 {
				t.Errorf("%s: requests = %d, want initial read and one poll", tc.name, calls.Load())
			}
		}()
	}
	wg.Wait()
}

func TestDeletionWaitersObserveNextPoll(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	for _, tc := range []struct {
		name string
		wait func(*Client) error
	}{
		{"VPS", func(c *Client) error { return c.WaitForVpsDeletion(context.Background(), "id", 20*time.Second) }},
		{"cache", func(c *Client) error { return c.WaitForCacheDeletion(context.Background(), "id", 20*time.Second) }},
		{"database", func(c *Client) error { return c.WaitForDatabaseDeletion(context.Background(), "id", 20*time.Second) }},
		{"bucket", func(c *Client) error {
			return c.WaitForStorageBucketDeletion(context.Background(), "id", 20*time.Second)
		}},
		{"serverless", func(c *Client) error { return c.WaitForServerlessDeletion(context.Background(), "id", 20*time.Second) }},
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var calls atomic.Int32
			server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
				if calls.Add(1) == 1 {
					_, _ = fmt.Fprint(w, waitStateBody("deleting"))
					return
				}
				w.WriteHeader(http.StatusNotFound)
			})
			defer server.Close()
			if err := tc.wait(newTestClient(server)); err != nil {
				t.Errorf("%s: %v", tc.name, err)
				return
			}
			if calls.Load() != 2 {
				t.Errorf("%s: requests = %d, want initial read and one poll", tc.name, calls.Load())
			}
		}()
	}
	wg.Wait()
}

func TestWaitersSurfacePollFailures(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	for _, tc := range []struct {
		name   string
		first  string
		second string
		code   int
		wait   func(*Client) error
		check  func(error) bool
	}{
		{"VPS permission", waitStateBody("provisioning"), `{"message":"forbidden"}`, http.StatusForbidden, func(c *Client) error {
			return c.WaitForVpsStatus(context.Background(), "id", "running", 20*time.Second)
		}, func(err error) bool {
			var apiErr *APIError
			return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden
		}},
		{"serverless failure", waitStateBody("creating"), waitStateBody("failed"), http.StatusOK, func(c *Client) error {
			return c.WaitForServerlessStatus(context.Background(), "id", "ready", 20*time.Second)
		}, func(err error) bool { return err != nil && strings.Contains(err.Error(), "error state") }},
		{"replica failure", `{"replicas":[{"replica_index":1,"status":"creating"}]}`, `{"replicas":[{"replica_index":1,"status":"failed"}]}`, http.StatusOK, func(c *Client) error {
			return c.WaitForDatabaseReplicaReady(context.Background(), "id", 1, 20*time.Second)
		}, func(err error) bool { return err != nil && strings.Contains(err.Error(), "error state") }},
		{"snapshot failure", snapshotStateBody("creating"), snapshotStateBody("restore_failed"), http.StatusOK, func(c *Client) error {
			return c.WaitForCacheSnapshotStatus(context.Background(), 1, "ready", 20*time.Second)
		}, func(err error) bool { return err != nil && strings.Contains(err.Error(), "restore_failed") }},
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var calls atomic.Int32
			server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
				if calls.Add(1) == 1 {
					_, _ = fmt.Fprint(w, tc.first)
					return
				}
				w.WriteHeader(tc.code)
				_, _ = fmt.Fprint(w, tc.second)
			})
			defer server.Close()
			if err := tc.wait(newTestClient(server)); !tc.check(err) {
				t.Errorf("%s: unexpected poll result: %v", tc.name, err)
				return
			}
			if calls.Load() != 2 {
				t.Errorf("%s: requests = %d, want initial read and one poll", tc.name, calls.Load())
			}
		}()
	}
	wg.Wait()
}

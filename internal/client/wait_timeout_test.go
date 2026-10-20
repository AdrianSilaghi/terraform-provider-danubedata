package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// A resource timeout must not wait for the next five- or ten-second poll tick.
func TestWaitersHonorShortTimeout(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"creating","instance":{"status":"creating"},"container":{"status":"creating"},"bucket":{"status":"creating"},"data":[{"id":1,"status":"creating"}],"replicas":[{"replica_index":1,"status":"creating"}],"pagination":{"last_page":1}}`))
	})
	defer server.Close()
	c := newTestClient(server)

	tests := map[string]func(context.Context) error{
		"VPS status":   func(ctx context.Context) error { return c.WaitForVpsStatus(ctx, "id", "running", 20*time.Millisecond) },
		"VPS deletion": func(ctx context.Context) error { return c.WaitForVpsDeletion(ctx, "id", 20*time.Millisecond) },
		"cache status": func(ctx context.Context) error {
			return c.WaitForCacheStatus(ctx, "id", "running", 20*time.Millisecond)
		},
		"cache deletion": func(ctx context.Context) error { return c.WaitForCacheDeletion(ctx, "id", 20*time.Millisecond) },
		"database status": func(ctx context.Context) error {
			return c.WaitForDatabaseStatus(ctx, "id", "running", 20*time.Millisecond)
		},
		"database deletion": func(ctx context.Context) error { return c.WaitForDatabaseDeletion(ctx, "id", 20*time.Millisecond) },
		"replica readiness": func(ctx context.Context) error {
			return c.WaitForDatabaseReplicaReady(ctx, "id", 1, 20*time.Millisecond)
		},
		"bucket status": func(ctx context.Context) error {
			return c.WaitForStorageBucketStatus(ctx, "id", "ready", 20*time.Millisecond)
		},
		"bucket deletion": func(ctx context.Context) error { return c.WaitForStorageBucketDeletion(ctx, "id", 20*time.Millisecond) },
		"serverless status": func(ctx context.Context) error {
			return c.WaitForServerlessStatus(ctx, "id", "ready", 20*time.Millisecond)
		},
		"serverless deletion": func(ctx context.Context) error { return c.WaitForServerlessDeletion(ctx, "id", 20*time.Millisecond) },
		"VPS snapshot": func(ctx context.Context) error {
			return c.WaitForVpsSnapshotStatus(ctx, 1, "ready", 20*time.Millisecond)
		},
		"cache snapshot": func(ctx context.Context) error {
			return c.WaitForCacheSnapshotStatus(ctx, 1, "ready", 20*time.Millisecond)
		},
		"database snapshot": func(ctx context.Context) error {
			return c.WaitForDatabaseSnapshotStatus(ctx, 1, "ready", 20*time.Millisecond)
		},
	}
	for name, wait := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			started := time.Now()
			err := wait(ctx)
			if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 300*time.Millisecond {
				t.Fatalf("20ms resource timeout took %s: %v", time.Since(started), err)
			}
		})
	}
}

func TestSnapshotWaitersCheckReadyImmediately(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1,"status":"ready"}],"pagination":{"last_page":1}}`))
	})
	defer server.Close()
	c := newTestClient(server)
	waits := map[string]func(context.Context) error{
		"VPS":      func(ctx context.Context) error { return c.WaitForVpsSnapshotStatus(ctx, 1, "ready", time.Second) },
		"cache":    func(ctx context.Context) error { return c.WaitForCacheSnapshotStatus(ctx, 1, "ready", time.Second) },
		"database": func(ctx context.Context) error { return c.WaitForDatabaseSnapshotStatus(ctx, 1, "ready", time.Second) },
	}
	for name, wait := range waits {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()
			if err := wait(ctx); err != nil {
				t.Fatalf("ready snapshot still waited for a poll tick: %v", err)
			}
		})
	}
}

func TestDeletionWaitersCheckMissingImmediately(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()
	c := newTestClient(server)
	waits := map[string]func(context.Context) error{
		"VPS":      func(ctx context.Context) error { return c.WaitForVpsDeletion(ctx, "id", time.Second) },
		"cache":    func(ctx context.Context) error { return c.WaitForCacheDeletion(ctx, "id", time.Second) },
		"database": func(ctx context.Context) error { return c.WaitForDatabaseDeletion(ctx, "id", time.Second) },
		"bucket":   func(ctx context.Context) error { return c.WaitForStorageBucketDeletion(ctx, "id", time.Second) },
		"serverless": func(ctx context.Context) error {
			return c.WaitForServerlessDeletion(ctx, "id", time.Second)
		},
	}
	for name, wait := range waits {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()
			if err := wait(ctx); err != nil {
				t.Fatalf("deleted resource still waited for a poll tick: %v", err)
			}
		})
	}
}

func TestReplicaWaiterChecksReadyImmediately(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"replicas":[{"replica_index":2,"status":"running","ready":true}]}`))
	})
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if err := newTestClient(server).WaitForDatabaseReplicaReady(ctx, "id", 2, time.Second); err != nil {
		t.Fatalf("ready replica still waited for a poll tick: %v", err)
	}
}

func TestStatusWaitersReportInitialFailure(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","instance":{"status":"error"},"container":{"status":"error"},"bucket":{"status":"error"}}`))
	})
	defer server.Close()
	c := newTestClient(server)
	waits := map[string]func(context.Context) error{
		"VPS":        func(ctx context.Context) error { return c.WaitForVpsStatus(ctx, "id", "running", time.Second) },
		"cache":      func(ctx context.Context) error { return c.WaitForCacheStatus(ctx, "id", "running", time.Second) },
		"database":   func(ctx context.Context) error { return c.WaitForDatabaseStatus(ctx, "id", "running", time.Second) },
		"bucket":     func(ctx context.Context) error { return c.WaitForStorageBucketStatus(ctx, "id", "ready", time.Second) },
		"serverless": func(ctx context.Context) error { return c.WaitForServerlessStatus(ctx, "id", "ready", time.Second) },
	}
	for name, wait := range waits {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()
			err := wait(ctx)
			if err == nil || errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("error state was not reported immediately: %v", err)
			}
		})
	}
}

func TestWaiterTimeoutCancelsInflightRequest(t *testing.T) {
	server := newTestServer(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	defer server.Close()
	started := time.Now()
	err := newTestClient(server).WaitForVpsStatus(context.Background(), "id", "running", 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 300*time.Millisecond {
		t.Fatalf("in-flight request was not canceled by resource timeout after %s: %v", time.Since(started), err)
	}
}

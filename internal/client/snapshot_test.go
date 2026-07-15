package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestClient_CreateVpsSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/snapshots/vps" {
			t.Errorf("Path = %v, want /snapshots/vps", r.URL.Path)
		}

		var req CreateVpsSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-snapshot" {
			t.Errorf("Name = %v, want my-snapshot", req.Name)
		}
		if req.VpsInstanceID != "vps-123" {
			t.Errorf("VpsInstanceID = %v, want vps-123", req.VpsInstanceID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createVpsSnapshotResponse{
			Message: "Snapshot created",
			Snapshot: VpsSnapshot{
				ID:            123,
				Name:          "my-snapshot",
				Status:        "creating",
				VpsInstanceID: "vps-123",
				SizeGB:        50.0,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snap, err := c.CreateVpsSnapshot(context.Background(), CreateVpsSnapshotRequest{
		VpsInstanceID: "vps-123",
		Name:          "my-snapshot",
		Description:   "Test snapshot",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID != 123 {
		t.Errorf("ID = %v, want 123", snap.ID)
	}
	if snap.VpsInstanceID != "vps-123" {
		t.Errorf("VpsInstanceID = %v, want vps-123", snap.VpsInstanceID)
	}
}

func TestClient_ListVpsSnapshots(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/snapshots/vps" {
			t.Errorf("Path = %v, want /snapshots/vps", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listVpsSnapshotsResponse{
			Data: []VpsSnapshot{
				{
					ID:     1,
					Name:   "snapshot-one",
					Status: "ready",
				},
				{
					ID:     2,
					Name:   "snapshot-two",
					Status: "ready",
				},
			},
			Pagination: Pagination{
				CurrentPage: 1,
				LastPage:    1,
				Total:       2,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snapshots, err := c.ListVpsSnapshots(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("got %d snapshots, want 2", len(snapshots))
	}
}

func TestClient_GetVpsSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listVpsSnapshotsResponse{
			Data: []VpsSnapshot{
				{
					ID:     123,
					Name:   "my-snapshot",
					Status: "ready",
					SizeGB: 50.0,
				},
			},
			Pagination: Pagination{
				CurrentPage: 1,
				LastPage:    1,
				Total:       1,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snap, err := c.GetVpsSnapshot(context.Background(), 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID != 123 {
		t.Errorf("ID = %v, want 123", snap.ID)
	}
	if snap.Status != "ready" {
		t.Errorf("Status = %v, want ready", snap.Status)
	}
}

func TestClient_GetVpsSnapshot_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listVpsSnapshotsResponse{
			Data: []VpsSnapshot{},
			Pagination: Pagination{
				CurrentPage: 1,
				LastPage:    1,
				Total:       0,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetVpsSnapshot(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_RestoreVpsSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/snapshots/vps/123/restore" {
			t.Errorf("Path = %v, want /snapshots/vps/123/restore", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Restore initiated"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.RestoreVpsSnapshot(context.Background(), 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteVpsSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/snapshots/vps/123" {
			t.Errorf("Path = %v, want /snapshots/vps/123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteVpsSnapshot(context.Background(), 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateCacheSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/snapshots/cache" {
			t.Errorf("Path = %v, want /snapshots/cache", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createCacheSnapshotResponse{
			Message: "Snapshot created",
			Snapshot: CacheSnapshot{
				ID:              123,
				Name:            "cache-snapshot",
				Status:          "creating",
				CacheInstanceID: "cache-123",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snap, err := c.CreateCacheSnapshot(context.Background(), CreateCacheSnapshotRequest{
		CacheInstanceID: "cache-123",
		Name:            "cache-snapshot",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID != 123 {
		t.Errorf("ID = %v, want 123", snap.ID)
	}
}

func TestClient_ListCacheSnapshots(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/snapshots/cache" {
			t.Errorf("Path = %v, want /snapshots/cache", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listCacheSnapshotsResponse{
			Data: []CacheSnapshot{
				{ID: 1, Name: "snapshot-one"},
				{ID: 2, Name: "snapshot-two"},
			},
			Pagination: Pagination{Total: 2},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snapshots, err := c.ListCacheSnapshots(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("got %d snapshots, want 2", len(snapshots))
	}
}

func TestClient_DeleteCacheSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/snapshots/cache/123" {
			t.Errorf("Path = %v, want /snapshots/cache/123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteCacheSnapshot(context.Background(), 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateDatabaseSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/snapshots/database" {
			t.Errorf("Path = %v, want /snapshots/database", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createDatabaseSnapshotResponse{
			Message: "Snapshot created",
			Snapshot: DatabaseSnapshot{
				ID:                 123,
				Name:               "database-snapshot",
				Status:             "creating",
				DatabaseInstanceID: "db-123",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snap, err := c.CreateDatabaseSnapshot(context.Background(), CreateDatabaseSnapshotRequest{
		DatabaseInstanceID: "db-123",
		Name:               "database-snapshot",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID != 123 {
		t.Errorf("ID = %v, want 123", snap.ID)
	}
}

func TestClient_ListDatabaseSnapshots(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/snapshots/database" {
			t.Errorf("Path = %v, want /snapshots/database", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listDatabaseSnapshotsResponse{
			Data: []DatabaseSnapshot{
				{ID: 1, Name: "snapshot-one"},
			},
			Pagination: Pagination{Total: 1},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	snapshots, err := c.ListDatabaseSnapshots(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("got %d snapshots, want 1", len(snapshots))
	}
}

func TestClient_DeleteDatabaseSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/snapshots/database/123" {
			t.Errorf("Path = %v, want /snapshots/database/123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteDatabaseSnapshot(context.Background(), 123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_WaitForVpsSnapshotStatus_Failed(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listVpsSnapshotsResponse{
			Data: []VpsSnapshot{
				{ID: 123, Status: "create_failed"},
			},
			Pagination: Pagination{CurrentPage: 1, LastPage: 1, Total: 1},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.WaitForVpsSnapshotStatus(context.Background(), 123, "ready", 7*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "create_failed") {
		t.Errorf("error = %v, want it to mention create_failed", err)
	}
	if strings.Contains(err.Error(), "timeout") {
		t.Errorf("error = %v, should fail fast on create_failed rather than timing out", err)
	}
}

func TestClient_WaitForCacheSnapshotStatus_Failed(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listCacheSnapshotsResponse{
			Data: []CacheSnapshot{
				{ID: 123, Status: "restore_failed"},
			},
			Pagination: Pagination{CurrentPage: 1, LastPage: 1, Total: 1},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.WaitForCacheSnapshotStatus(context.Background(), 123, "ready", 12*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "restore_failed") {
		t.Errorf("error = %v, want it to mention restore_failed", err)
	}
	if strings.Contains(err.Error(), "timeout") {
		t.Errorf("error = %v, should fail fast on restore_failed rather than timing out", err)
	}
}

func TestClient_WaitForDatabaseSnapshotStatus_Failed(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listDatabaseSnapshotsResponse{
			Data: []DatabaseSnapshot{
				{ID: 123, Status: "restore_failed"},
			},
			Pagination: Pagination{CurrentPage: 1, LastPage: 1, Total: 1},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.WaitForDatabaseSnapshotStatus(context.Background(), 123, "ready", 12*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "restore_failed") {
		t.Errorf("error = %v, want it to mention restore_failed", err)
	}
	if strings.Contains(err.Error(), "timeout") {
		t.Errorf("error = %v, should fail fast on restore_failed rather than timing out", err)
	}
}

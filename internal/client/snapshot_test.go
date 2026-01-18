package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
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
				ID:            "snap-123",
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
	if snap.ID != "snap-123" {
		t.Errorf("ID = %v, want snap-123", snap.ID)
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
					ID:     "snap-1",
					Name:   "snapshot-one",
					Status: "ready",
				},
				{
					ID:     "snap-2",
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
					ID:     "snap-123",
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
	snap, err := c.GetVpsSnapshot(context.Background(), "snap-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.ID != "snap-123" {
		t.Errorf("ID = %v, want snap-123", snap.ID)
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
	_, err := c.GetVpsSnapshot(context.Background(), "nonexistent")

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
		if r.URL.Path != "/snapshots/vps/snap-123/restore" {
			t.Errorf("Path = %v, want /snapshots/vps/snap-123/restore", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Restore initiated"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.RestoreVpsSnapshot(context.Background(), "snap-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeleteVpsSnapshot(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/snapshots/vps/snap-123" {
			t.Errorf("Path = %v, want /snapshots/vps/snap-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteVpsSnapshot(context.Background(), "snap-123")

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
				ID:              "csnap-123",
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
	if snap.ID != "csnap-123" {
		t.Errorf("ID = %v, want csnap-123", snap.ID)
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
				{ID: "csnap-1", Name: "snapshot-one"},
				{ID: "csnap-2", Name: "snapshot-two"},
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
		if r.URL.Path != "/snapshots/cache/csnap-123" {
			t.Errorf("Path = %v, want /snapshots/cache/csnap-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteCacheSnapshot(context.Background(), "csnap-123")

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
				ID:                 "dbsnap-123",
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
	if snap.ID != "dbsnap-123" {
		t.Errorf("ID = %v, want dbsnap-123", snap.ID)
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
				{ID: "dbsnap-1", Name: "snapshot-one"},
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
		if r.URL.Path != "/snapshots/database/dbsnap-123" {
			t.Errorf("Path = %v, want /snapshots/database/dbsnap-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteDatabaseSnapshot(context.Background(), "dbsnap-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

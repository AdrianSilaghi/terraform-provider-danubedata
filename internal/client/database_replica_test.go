package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_ListDatabaseReplicas(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/database/db-123/replicas" {
			t.Errorf("Path = %v, want /database/db-123/replicas", r.URL.Path)
		}

		endpoint := "db-123-reader.tenant-1.svc.cluster.local"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(DatabaseReplicaList{
			Master: DatabaseReplicaMaster{
				Name:     "db-123",
				NodeID:   "node-0",
				Endpoint: &endpoint,
				Status:   "running",
				Ready:    true,
			},
			Replicas: []DatabaseReplica{
				{Name: "db-123-r1", NodeID: "node-r1", ReplicaIndex: 1, Status: "running", Ready: true, IsReplicationHealthy: true},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	list, err := c.ListDatabaseReplicas(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Replicas) != 1 {
		t.Fatalf("got %d replicas, want 1", len(list.Replicas))
	}
	if list.Replicas[0].ReplicaIndex != 1 {
		t.Errorf("ReplicaIndex = %v, want 1", list.Replicas[0].ReplicaIndex)
	}
	if list.Master.Name != "db-123" {
		t.Errorf("Master.Name = %v, want db-123", list.Master.Name)
	}
}

func TestClient_AddDatabaseReplicas(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/database/db-123/replicas" {
			t.Errorf("Path = %v, want /database/db-123/replicas", r.URL.Path)
		}

		var req AddDatabaseReplicasRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.ReplicaCount != 2 {
			t.Errorf("ReplicaCount = %v, want 2", req.ReplicaCount)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(addDatabaseReplicasResponse{
			Message: "Replicas added",
			Replicas: []DatabaseReplica{
				{Name: "db-123-r1", ReplicaIndex: 1, Status: "provisioning"},
				{Name: "db-123-r2", ReplicaIndex: 2, Status: "provisioning"},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	added, err := c.AddDatabaseReplicas(context.Background(), "db-123", AddDatabaseReplicasRequest{ReplicaCount: 2})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != 2 {
		t.Fatalf("got %d replicas, want 2", len(added))
	}
	if added[0].ReplicaIndex != 1 {
		t.Errorf("added[0].ReplicaIndex = %v, want 1", added[0].ReplicaIndex)
	}
}

func TestClient_DeleteDatabaseReplica(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/database/db-123/replicas/2" {
			t.Errorf("Path = %v, want /database/db-123/replicas/2", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "deleted", "status": "deleting"})
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteDatabaseReplica(context.Background(), "db-123", 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

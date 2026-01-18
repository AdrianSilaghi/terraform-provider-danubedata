package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/database" {
			t.Errorf("Path = %v, want /database", r.URL.Path)
		}

		var req CreateDatabaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-database" {
			t.Errorf("Name = %v, want my-database", req.Name)
		}
		if req.DatabaseProviderID != 1 {
			t.Errorf("DatabaseProviderID = %v, want 1", req.DatabaseProviderID)
		}
		if req.StorageSizeGB != 20 {
			t.Errorf("StorageSizeGB = %v, want 20", req.StorageSizeGB)
		}

		endpoint := "my-database.mysql.cluster.local"
		port := 3306
		dbName := "mydb"
		username := "admin"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createDatabaseResponse{
			Message: "Database instance created",
			Instance: DatabaseInstance{
				ID:              "db-123",
				Name:            "my-database",
				Status:          "creating",
				ResourceProfile: "standard",
				CPUCores:        2,
				MemorySizeMB:    2048,
				StorageSizeGB:   20,
				DatabaseName:    &dbName,
				Version:         "8.0",
				Endpoint:        &endpoint,
				Port:            &port,
				Username:        &username,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	db, err := c.CreateDatabase(context.Background(), CreateDatabaseRequest{
		Name:               "my-database",
		DatabaseProviderID: 1,
		DatabaseName:       "mydb",
		StorageSizeGB:      20,
		MemorySizeMB:       2048,
		CPUCores:           2,
		Version:            "8.0",
		HetznerDatacenter:  "fsn1",
		ResourceProfile:    "standard",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-123" {
		t.Errorf("ID = %v, want db-123", db.ID)
	}
	if db.Name != "my-database" {
		t.Errorf("Name = %v, want my-database", db.Name)
	}
	if db.StorageSizeGB != 20 {
		t.Errorf("StorageSizeGB = %v, want 20", db.StorageSizeGB)
	}
}

func TestClient_GetDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/database/db-123" {
			t.Errorf("Path = %v, want /database/db-123", r.URL.Path)
		}

		endpoint := "my-database.mysql.cluster.local"
		port := 3306
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(showDatabaseResponse{
			Instance: DatabaseInstance{
				ID:            "db-123",
				Name:          "my-database",
				Status:        "running",
				CPUCores:      2,
				MemorySizeMB:  2048,
				StorageSizeGB: 20,
				Version:       "8.0",
				Endpoint:      &endpoint,
				Port:          &port,
			},
			ConnectionInfo: "mysql://admin@my-database.mysql.cluster.local:3306/mydb",
			MonthlyCost:    29.99,
		})
	})
	defer server.Close()

	c := newTestClient(server)
	db, err := c.GetDatabase(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-123" {
		t.Errorf("ID = %v, want db-123", db.ID)
	}
	if db.Status != "running" {
		t.Errorf("Status = %v, want running", db.Status)
	}
	if db.StorageSizeGB != 20 {
		t.Errorf("StorageSizeGB = %v, want 20", db.StorageSizeGB)
	}
}

func TestClient_GetDatabase_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Database instance not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetDatabase(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_UpdateDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/database/db-123" {
			t.Errorf("Path = %v, want /database/db-123", r.URL.Path)
		}

		var req UpdateDatabaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.StorageSizeGB == nil || *req.StorageSizeGB != 50 {
			t.Errorf("StorageSizeGB = %v, want 50", req.StorageSizeGB)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updateDatabaseResponse{
			Message: "Database instance updated",
			Instance: DatabaseInstance{
				ID:            "db-123",
				Name:          "my-database",
				StorageSizeGB: 50,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	storageSize := 50
	db, err := c.UpdateDatabase(context.Background(), "db-123", UpdateDatabaseRequest{
		StorageSizeGB: &storageSize,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.StorageSizeGB != 50 {
		t.Errorf("StorageSizeGB = %v, want 50", db.StorageSizeGB)
	}
}

func TestClient_DeleteDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/database/db-123" {
			t.Errorf("Path = %v, want /database/db-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteDatabase(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StartDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/database/db-123/start" {
			t.Errorf("Path = %v, want /database/db-123/start", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Database starting"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StartDatabase(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StopDatabase(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/database/db-123/stop" {
			t.Errorf("Path = %v, want /database/db-123/stop", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StopDatabase(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetDatabaseCredentials(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/database/db-123/credentials" {
			t.Errorf("Path = %v, want /database/db-123/credentials", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DatabaseCredentials{
			ConnectionInfo: "mysql://admin@my-database.mysql.cluster.local:3306/mydb",
			Username:       "admin",
			Password:       "secret-password",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	creds, err := c.GetDatabaseCredentials(context.Background(), "db-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Username != "admin" {
		t.Errorf("Username = %v, want admin", creds.Username)
	}
	if creds.Password != "secret-password" {
		t.Errorf("Password = %v, want secret-password", creds.Password)
	}
}

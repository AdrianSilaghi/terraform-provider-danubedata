package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateStorageBucket(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/storage/buckets" {
			t.Errorf("Path = %v, want /storage/buckets", r.URL.Path)
		}

		var req CreateStorageBucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-bucket" {
			t.Errorf("Name = %v, want my-bucket", req.Name)
		}
		if req.Region != "fsn1" {
			t.Errorf("Region = %v, want fsn1", req.Region)
		}
		if !req.VersioningEnabled {
			t.Error("VersioningEnabled = false, want true")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createStorageBucketResponse{
			Message: "Bucket created",
			Bucket: StorageBucket{
				ID:                "bucket-123",
				Name:              "my-bucket",
				Status:            "active",
				Region:            "fsn1",
				EndpointURL:       "https://s3.fsn1.danubedata.com",
				MinioBucketName:   "dd-1-my-bucket",
				VersioningEnabled: true,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	bucket, err := c.CreateStorageBucket(context.Background(), CreateStorageBucketRequest{
		Name:              "my-bucket",
		Region:            "fsn1",
		VersioningEnabled: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bucket.ID != "bucket-123" {
		t.Errorf("ID = %v, want bucket-123", bucket.ID)
	}
	if bucket.Name != "my-bucket" {
		t.Errorf("Name = %v, want my-bucket", bucket.Name)
	}
	if !bucket.VersioningEnabled {
		t.Error("VersioningEnabled = false, want true")
	}
}

func TestClient_GetStorageBucket(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/storage/buckets/bucket-123" {
			t.Errorf("Path = %v, want /storage/buckets/bucket-123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showStorageBucketResponse{
			Bucket: StorageBucket{
				ID:                "bucket-123",
				Name:              "my-bucket",
				Status:            "active",
				Region:            "fsn1",
				EndpointURL:       "https://s3.fsn1.danubedata.com",
				SizeBytes:         1024 * 1024 * 100, // 100 MB
				ObjectCount:       50,
				VersioningEnabled: true,
			},
			Endpoint: "https://s3.fsn1.danubedata.com",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	bucket, err := c.GetStorageBucket(context.Background(), "bucket-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bucket.ID != "bucket-123" {
		t.Errorf("ID = %v, want bucket-123", bucket.ID)
	}
	if bucket.Status != "active" {
		t.Errorf("Status = %v, want active", bucket.Status)
	}
	if bucket.ObjectCount != 50 {
		t.Errorf("ObjectCount = %v, want 50", bucket.ObjectCount)
	}
}

func TestClient_GetStorageBucket_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Bucket not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetStorageBucket(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_UpdateStorageBucket(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/storage/buckets/bucket-123" {
			t.Errorf("Path = %v, want /storage/buckets/bucket-123", r.URL.Path)
		}

		var req UpdateStorageBucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.PublicAccess == nil || !*req.PublicAccess {
			t.Error("PublicAccess = false, want true")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updateStorageBucketResponse{
			Message: "Bucket updated",
			Bucket: StorageBucket{
				ID:           "bucket-123",
				Name:         "my-bucket",
				PublicAccess: true,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	publicAccess := true
	bucket, err := c.UpdateStorageBucket(context.Background(), "bucket-123", UpdateStorageBucketRequest{
		PublicAccess: &publicAccess,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bucket.PublicAccess {
		t.Error("PublicAccess = false, want true")
	}
}

func TestClient_DeleteStorageBucket(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/storage/buckets/bucket-123" {
			t.Errorf("Path = %v, want /storage/buckets/bucket-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteStorageBucket(context.Background(), "bucket-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateStorageAccessKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/storage/access-keys" {
			t.Errorf("Path = %v, want /storage/access-keys", r.URL.Path)
		}

		var req CreateStorageAccessKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-key" {
			t.Errorf("Name = %v, want my-key", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(CreateStorageAccessKeyResponse{
			ID:              "key-123",
			Name:            "my-key",
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			Message:         "Access key created",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	key, err := c.CreateStorageAccessKey(context.Background(), CreateStorageAccessKeyRequest{
		Name: "my-key",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != "key-123" {
		t.Errorf("ID = %v, want key-123", key.ID)
	}
	if key.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKeyID = %v, want AKIAIOSFODNN7EXAMPLE", key.AccessKeyID)
	}
	if key.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("SecretAccessKey = %v, want wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", key.SecretAccessKey)
	}
}

func TestClient_GetStorageAccessKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/storage/access-keys/key-123" {
			t.Errorf("Path = %v, want /storage/access-keys/key-123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			AccessKey StorageAccessKey `json:"access_key"`
		}{
			AccessKey: StorageAccessKey{
				ID:          "key-123",
				Name:        "my-key",
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Status:      "active",
				AccessType:  "full",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	key, err := c.GetStorageAccessKey(context.Background(), "key-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != "key-123" {
		t.Errorf("ID = %v, want key-123", key.ID)
	}
	if key.Status != "active" {
		t.Errorf("Status = %v, want active", key.Status)
	}
}

func TestClient_GetStorageAccessKey_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Access key not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetStorageAccessKey(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_DeleteStorageAccessKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/storage/access-keys/key-123" {
			t.Errorf("Path = %v, want /storage/access-keys/key-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteStorageAccessKey(context.Background(), "key-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestClient_CreateCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/cache" {
			t.Errorf("Path = %v, want /cache", r.URL.Path)
		}

		var req CreateCacheRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-cache" {
			t.Errorf("Name = %v, want my-cache", req.Name)
		}
		if req.Provider != "redis" {
			t.Errorf("Provider = %v, want redis", req.Provider)
		}

		endpoint := "my-cache.redis.cluster.local"
		port := 6379
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createCacheResponse{
			Message: "Cache instance created",
			Instance: CacheInstance{
				ID:              "cache-123",
				Name:            "my-cache",
				Status:          "creating",
				ResourceProfile: "standard",
				CPUCores:        1,
				MemorySizeMB:    512,
				Provider:        CacheProvider{ID: 1, Name: "redis"},
				Endpoint:        &endpoint,
				Port:            &port,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	cache, err := c.CreateCache(context.Background(), CreateCacheRequest{
		Name:            "my-cache",
		Provider:        "redis",
		Datacenter:      "fsn1",
		ResourceProfile: "standard",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cache.ID != "cache-123" {
		t.Errorf("ID = %v, want cache-123", cache.ID)
	}
	if cache.Name != "my-cache" {
		t.Errorf("Name = %v, want my-cache", cache.Name)
	}
}

func TestClient_GetCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/cache/cache-123" {
			t.Errorf("Path = %v, want /cache/cache-123", r.URL.Path)
		}

		endpoint := "my-cache.redis.cluster.local"
		port := 6379
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showCacheResponse{
			Instance: CacheInstance{
				ID:           "cache-123",
				Name:         "my-cache",
				Status:       "running",
				CPUCores:     2,
				MemorySizeMB: 1024,
				Version:      "7.2",
				Provider:     CacheProvider{ID: 1, Name: "Redis", Type: "redis"},
				Endpoint:     &endpoint,
				Port:         &port,
			},
			ConnectionInfo: "redis://my-cache.redis.cluster.local:6379",
			MonthlyCost:    9.99,
		})
	})
	defer server.Close()

	c := newTestClient(server)
	cache, err := c.GetCache(context.Background(), "cache-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cache.ID != "cache-123" {
		t.Errorf("ID = %v, want cache-123", cache.ID)
	}
	if cache.Status != "running" {
		t.Errorf("Status = %v, want running", cache.Status)
	}
	if cache.MemorySizeMB != 1024 {
		t.Errorf("MemorySizeMB = %v, want 1024", cache.MemorySizeMB)
	}
	if cache.Provider.Type != "redis" {
		t.Errorf("Provider.Type = %v, want redis", cache.Provider.Type)
	}
}

func TestClient_GetCache_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Cache instance not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetCache(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_UpdateCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/cache/cache-123" {
			t.Errorf("Path = %v, want /cache/cache-123", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if strings.Contains(string(body), "memory_size_mb") || strings.Contains(string(body), "cpu_cores") {
			t.Errorf("request body must not contain memory_size_mb/cpu_cores, got: %s", body)
		}

		var req UpdateCacheRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.ResourceProfile != "small" {
			t.Errorf("ResourceProfile = %v, want small", req.ResourceProfile)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updateCacheResponse{
			Message: "Cache instance updated",
			Instance: CacheInstance{
				ID:              "cache-123",
				Name:            "my-cache",
				ResourceProfile: "small",
				MemorySizeMB:    2048,
				CPUCores:        2,
				Provider:        CacheProvider{ID: 1, Name: "Redis", Type: "redis"},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	cache, err := c.UpdateCache(context.Background(), "cache-123", UpdateCacheRequest{
		ResourceProfile: "small",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cache.ResourceProfile != "small" {
		t.Errorf("ResourceProfile = %v, want small", cache.ResourceProfile)
	}
	if cache.MemorySizeMB != 2048 {
		t.Errorf("MemorySizeMB = %v, want 2048", cache.MemorySizeMB)
	}
}

func TestClient_DeleteCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/cache/cache-123" {
			t.Errorf("Path = %v, want /cache/cache-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteCache(context.Background(), "cache-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StartCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/cache/cache-123/start" {
			t.Errorf("Path = %v, want /cache/cache-123/start", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Cache starting"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StartCache(context.Background(), "cache-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StopCache(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/cache/cache-123/stop" {
			t.Errorf("Path = %v, want /cache/cache-123/stop", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StopCache(context.Background(), "cache-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetCacheConnectionInfo(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/cache/cache-123/connection-info" {
			t.Errorf("Path = %v, want /cache/cache-123/connection-info", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CacheConnectionInfo{
			ConnectionInfo: "redis://my-cache.redis.cluster.local:6379",
			Password:       "secret-password",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	info, err := c.GetCacheConnectionInfo(context.Background(), "cache-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Password != "secret-password" {
		t.Errorf("Password = %v, want secret-password", info.Password)
	}
	if info.ConnectionInfo != "redis://my-cache.redis.cluster.local:6379" {
		t.Errorf("ConnectionInfo = %v, want redis://my-cache.redis.cluster.local:6379", info.ConnectionInfo)
	}
}

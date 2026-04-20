package client

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_EnableCacheDns(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/cache/cache-123/dns" {
			t.Errorf("Path = %v, want /cache/cache-123/dns", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "DNS enabled"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.EnableCacheDns(context.Background(), "cache-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DisableCacheDns(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/cache/cache-123/dns" {
			t.Errorf("Path = %v, want /cache/cache-123/dns", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.DisableCacheDns(context.Background(), "cache-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_EnableDatabaseDns(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/database/db-123/dns" {
			t.Errorf("Path = %v, want /database/db-123/dns", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "DNS enabled"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.EnableDatabaseDns(context.Background(), "db-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DisableDatabaseDns(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/database/db-123/dns" {
			t.Errorf("Path = %v, want /database/db-123/dns", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	if err := c.DisableDatabaseDns(context.Background(), "db-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

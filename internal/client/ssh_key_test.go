package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateSshKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/ssh-keys" {
			t.Errorf("Path = %v, want /ssh-keys", r.URL.Path)
		}

		var req CreateSshKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-key" {
			t.Errorf("Name = %v, want my-key", req.Name)
		}
		if req.PublicKey != "ssh-ed25519 AAAAC3NzaC1..." {
			t.Errorf("PublicKey = %v, want ssh-ed25519 AAAAC3NzaC1...", req.PublicKey)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createSshKeyResponse{
			Message: "SSH key created",
			Key: SshKey{
				ID:          123,
				Name:        "my-key",
				Fingerprint: "SHA256:abc123",
				PublicKey:   "ssh-ed25519 AAAAC3NzaC1...",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	key, err := c.CreateSshKey(context.Background(), CreateSshKeyRequest{
		Name:      "my-key",
		PublicKey: "ssh-ed25519 AAAAC3NzaC1...",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != 123 {
		t.Errorf("ID = %v, want 123", key.ID)
	}
	if key.Name != "my-key" {
		t.Errorf("Name = %v, want my-key", key.Name)
	}
	if key.Fingerprint != "SHA256:abc123" {
		t.Errorf("Fingerprint = %v, want SHA256:abc123", key.Fingerprint)
	}
}

func TestClient_GetSshKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/ssh-keys/123" {
			t.Errorf("Path = %v, want /ssh-keys/123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showSshKeyResponse{
			Key: SshKey{
				ID:          123,
				Name:        "my-key",
				Fingerprint: "SHA256:abc123",
				PublicKey:   "ssh-ed25519 AAAAC3NzaC1...",
				UserID:      1,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	key, err := c.GetSshKey(context.Background(), "123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != 123 {
		t.Errorf("ID = %v, want 123", key.ID)
	}
	if key.Fingerprint != "SHA256:abc123" {
		t.Errorf("Fingerprint = %v, want SHA256:abc123", key.Fingerprint)
	}
}

func TestClient_GetSshKey_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "SSH key not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetSshKey(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ListSshKeys(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/ssh-keys" {
			t.Errorf("Path = %v, want /ssh-keys", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listSshKeysResponse{
			Data: []SshKey{
				{
					ID:          1,
					Name:        "key-one",
					Fingerprint: "SHA256:aaa",
				},
				{
					ID:          2,
					Name:        "key-two",
					Fingerprint: "SHA256:bbb",
				},
			},
			Pagination: Pagination{
				CurrentPage: 1,
				LastPage:    1,
				PerPage:     10,
				Total:       2,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	keys, err := c.ListSshKeys(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(keys))
	}
	if keys[0].ID != 1 {
		t.Errorf("keys[0].ID = %v, want 1", keys[0].ID)
	}
	if keys[1].Name != "key-two" {
		t.Errorf("keys[1].Name = %v, want key-two", keys[1].Name)
	}
}

func TestClient_DeleteSshKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/ssh-keys/123" {
			t.Errorf("Path = %v, want /ssh-keys/123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteSshKey(context.Background(), "123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

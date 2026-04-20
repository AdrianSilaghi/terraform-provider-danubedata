package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetVpsPassword(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vps/vps-123/password" {
			t.Errorf("Path = %v, want /vps/vps-123/password", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(VpsPassword{
			Password: "s3cret",
			Username: "root",
			PublicIP: strPtr("1.2.3.4"),
		})
	})
	defer server.Close()

	c := newTestClient(server)
	creds, err := c.GetVpsPassword(context.Background(), "vps-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Password != "s3cret" {
		t.Errorf("Password = %v, want s3cret", creds.Password)
	}
	if creds.Username != "root" {
		t.Errorf("Username = %v, want root", creds.Username)
	}
}

func strPtr(s string) *string { return &s }

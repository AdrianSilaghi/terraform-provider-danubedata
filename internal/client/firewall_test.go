package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/firewalls" {
			t.Errorf("Path = %v, want /firewalls", r.URL.Path)
		}

		var req CreateFirewallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-firewall" {
			t.Errorf("Name = %v, want my-firewall", req.Name)
		}
		if len(req.Rules) != 1 {
			t.Errorf("Rules count = %v, want 1", len(req.Rules))
		}
		if req.Rules[0].Action != "allow" {
			t.Errorf("Rules[0].Action = %v, want allow", req.Rules[0].Action)
		}

		portStart := 22
		portEnd := 22
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createFirewallResponse{
			Message: "Firewall created",
			Firewall: Firewall{
				ID:            "fw-123",
				Name:          "my-firewall",
				Status:        "active",
				DefaultAction: "deny",
				Rules: []FirewallRule{
					{
						ID:             "rule-1",
						Action:         "allow",
						Direction:      "inbound",
						Protocol:       "tcp",
						PortRangeStart: &portStart,
						PortRangeEnd:   &portEnd,
						SourceIPs:      []string{"0.0.0.0/0"},
						Priority:       100,
					},
				},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	portStart := 22
	portEnd := 22
	fw, err := c.CreateFirewall(context.Background(), CreateFirewallRequest{
		Name:          "my-firewall",
		DefaultAction: "deny",
		Rules: []CreateFirewallRuleRequest{
			{
				Action:         "allow",
				Direction:      "inbound",
				Protocol:       "tcp",
				PortRangeStart: &portStart,
				PortRangeEnd:   &portEnd,
				SourceIPs:      []string{"0.0.0.0/0"},
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.ID != "fw-123" {
		t.Errorf("ID = %v, want fw-123", fw.ID)
	}
	if len(fw.Rules) != 1 {
		t.Fatalf("Rules count = %v, want 1", len(fw.Rules))
	}
	if fw.Rules[0].Action != "allow" {
		t.Errorf("Rules[0].Action = %v, want allow", fw.Rules[0].Action)
	}
}

func TestClient_GetFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123" {
			t.Errorf("Path = %v, want /firewalls/fw-123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showFirewallResponse{
			Firewall: Firewall{
				ID:            "fw-123",
				Name:          "my-firewall",
				Status:        "active",
				DefaultAction: "deny",
				Rules:         []FirewallRule{},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	fw, err := c.GetFirewall(context.Background(), "fw-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.ID != "fw-123" {
		t.Errorf("ID = %v, want fw-123", fw.ID)
	}
	if fw.DefaultAction != "deny" {
		t.Errorf("DefaultAction = %v, want deny", fw.DefaultAction)
	}
}

func TestClient_GetFirewall_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Firewall not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetFirewall(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ListFirewalls(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/firewalls" {
			t.Errorf("Path = %v, want /firewalls", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listFirewallsResponse{
			Data: []Firewall{
				{
					ID:   "fw-1",
					Name: "firewall-one",
				},
				{
					ID:   "fw-2",
					Name: "firewall-two",
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
	firewalls, err := c.ListFirewalls(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(firewalls) != 2 {
		t.Fatalf("got %d firewalls, want 2", len(firewalls))
	}
	if firewalls[0].ID != "fw-1" {
		t.Errorf("firewalls[0].ID = %v, want fw-1", firewalls[0].ID)
	}
}

func TestClient_UpdateFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123" {
			t.Errorf("Path = %v, want /firewalls/fw-123", r.URL.Path)
		}

		var req UpdateFirewallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Name != "updated-firewall" {
			t.Errorf("Name = %v, want updated-firewall", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showFirewallResponse{
			Firewall: Firewall{
				ID:   "fw-123",
				Name: "updated-firewall",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	fw, err := c.UpdateFirewall(context.Background(), "fw-123", UpdateFirewallRequest{
		Name: "updated-firewall",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.Name != "updated-firewall" {
		t.Errorf("Name = %v, want updated-firewall", fw.Name)
	}
}

func TestClient_DeleteFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123" {
			t.Errorf("Path = %v, want /firewalls/fw-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteFirewall(context.Background(), "fw-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_AttachFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123/attach" {
			t.Errorf("Path = %v, want /firewalls/fw-123/attach", r.URL.Path)
		}

		var req AttachFirewallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.InstanceType != "vps" {
			t.Errorf("InstanceType = %v, want vps", req.InstanceType)
		}
		if req.InstanceID != "vps-456" {
			t.Errorf("InstanceID = %v, want vps-456", req.InstanceID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Firewall attached"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.AttachFirewall(context.Background(), "fw-123", AttachFirewallRequest{
		InstanceType: "vps",
		InstanceID:   "vps-456",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DetachFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123/detach" {
			t.Errorf("Path = %v, want /firewalls/fw-123/detach", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DetachFirewall(context.Background(), "fw-123", AttachFirewallRequest{
		InstanceType: "vps",
		InstanceID:   "vps-456",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DeployFirewall(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/firewalls/fw-123/deploy" {
			t.Errorf("Path = %v, want /firewalls/fw-123/deploy", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Firewall deployment initiated"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeployFirewall(context.Background(), "fw-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

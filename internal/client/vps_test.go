package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestClient_CreateVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/vps" {
			t.Errorf("Path = %v, want /vps", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var req CreateVpsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "test-vps" {
			t.Errorf("Name = %v, want test-vps", req.Name)
		}
		if req.Image != "ubuntu-22.04" {
			t.Errorf("Image = %v, want ubuntu-22.04", req.Image)
		}
		if req.SSHKeyID == nil || *req.SSHKeyID != 123 {
			t.Errorf("SSHKeyID = %v, want 123", req.SSHKeyID)
		}

		// ssh_key_id must be sent as a JSON number, not a string: the API's
		// StoreVpsInstanceRequest declares it `ssh_key_id?: number | null`.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("failed to decode raw request: %v", err)
		}
		if string(raw["ssh_key_id"]) != "123" {
			t.Errorf("raw ssh_key_id = %s, want bare number 123", raw["ssh_key_id"])
		}

		// cpu_cores/memory_size_gb/storage_size_gb are not in
		// StoreVpsInstanceRequest's accepted fields; CreateVpsRequest must
		// not be able to send them at all.
		for _, key := range []string{"cpu_cores", "memory_size_gb", "storage_size_gb"} {
			if _, present := raw[key]; present {
				t.Errorf("request body unexpectedly contains %q", key)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createVpsResponse{
			Message: "VPS created",
			Instance: VpsInstance{
				ID:     "vps-123",
				Name:   "test-vps",
				Status: "creating",
				Image:  "ubuntu-22.04",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	var sshKeyID int64 = 123
	vps, err := c.CreateVps(context.Background(), CreateVpsRequest{
		Name:       "test-vps",
		Image:      "ubuntu-22.04",
		Datacenter: "fsn1",
		AuthMethod: "ssh_key",
		SSHKeyID:   &sshKeyID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vps.ID != "vps-123" {
		t.Errorf("ID = %v, want vps-123", vps.ID)
	}
	if vps.Name != "test-vps" {
		t.Errorf("Name = %v, want test-vps", vps.Name)
	}
}

func TestClient_GetVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/vps/vps-123" {
			t.Errorf("Path = %v, want /vps/vps-123", r.URL.Path)
		}

		publicIP := "192.168.1.1"
		privateIP := "10.0.0.5"
		var sshKeyID int64 = 456
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showVpsResponse{
			Instance: VpsInstance{
				ID:                "vps-123",
				Name:              "test-vps",
				Status:            "running",
				CPUAllocationType: "shared",
				CPUCores:          2,
				MemorySizeGB:      4,
				StorageSizeGB:     50,
				PublicIP:          &publicIP,
				SSHKeyID:          &sshKeyID,
			},
			ConnectionInfo: &vpsConnectionInfo{
				PrivateIP: &privateIP,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	vps, err := c.GetVps(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vps.ID != "vps-123" {
		t.Errorf("ID = %v, want vps-123", vps.ID)
	}
	if vps.Status != "running" {
		t.Errorf("Status = %v, want running", vps.Status)
	}
	if vps.CPUCores != 2 {
		t.Errorf("CPUCores = %v, want 2", vps.CPUCores)
	}
	if *vps.PublicIP != "192.168.1.1" {
		t.Errorf("PublicIP = %v, want 192.168.1.1", *vps.PublicIP)
	}
	if vps.PrivateIP == nil || *vps.PrivateIP != "10.0.0.5" {
		t.Errorf("PrivateIP = %v, want 10.0.0.5", vps.PrivateIP)
	}
	if vps.SSHKeyID == nil || *vps.SSHKeyID != 456 {
		t.Errorf("SSHKeyID = %v, want 456", vps.SSHKeyID)
	}
}

func TestClient_GetVps_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "VPS not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetVps(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_UpdateVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/vps/vps-123" {
			t.Errorf("Path = %v, want /vps/vps-123", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var req UpdateVpsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.ResourceProfile != "micro_shared" {
			t.Errorf("ResourceProfile = %v, want micro_shared", req.ResourceProfile)
		}

		// UpdateVpsInstanceRequest only accepts resource_profile and
		// cpu_allocation_type; UpdateVpsRequest must not be able to send
		// anything else.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("failed to decode raw request: %v", err)
		}
		for _, key := range []string{"cpu_cores", "memory_size_gb", "storage_size_gb", "password", "password_confirmation"} {
			if _, present := raw[key]; present {
				t.Errorf("request body unexpectedly contains %q", key)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showVpsResponse{
			Instance: VpsInstance{
				ID:              "vps-123",
				Name:            "test-vps",
				Status:          "running",
				ResourceProfile: "micro_shared",
				CPUCores:        3,
				MemorySizeGB:    4,
				StorageSizeGB:   60,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	vps, err := c.UpdateVps(context.Background(), "vps-123", UpdateVpsRequest{
		ResourceProfile: "micro_shared",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vps.ResourceProfile != "micro_shared" {
		t.Errorf("ResourceProfile = %v, want micro_shared", vps.ResourceProfile)
	}
	if vps.CPUCores != 3 {
		t.Errorf("CPUCores = %v, want 3", vps.CPUCores)
	}
}

func TestClient_DeleteVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/vps/vps-123" {
			t.Errorf("Path = %v, want /vps/vps-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteVps(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StartVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/vps/vps-123/start" {
			t.Errorf("Path = %v, want /vps/vps-123/start", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "VPS starting"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StartVps(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StopVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/vps/vps-123/stop" {
			t.Errorf("Path = %v, want /vps/vps-123/stop", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.StopVps(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_RebootVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/vps/vps-123/reboot" {
			t.Errorf("Path = %v, want /vps/vps-123/reboot", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.RebootVps(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ReinstallVps(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/vps/vps-123/reinstall" {
			t.Errorf("Path = %v, want /vps/vps-123/reinstall", r.URL.Path)
		}

		var req ReinstallVpsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Image != "debian-12" {
			t.Errorf("Image = %v, want debian-12", req.Image)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.ReinstallVps(context.Background(), "vps-123", ReinstallVpsRequest{
		Image: "debian-12",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetVpsStatus(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vps/vps-123/status" {
			t.Errorf("Path = %v, want /vps/vps-123/status", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(statusVpsResponse{
			Status:      "running",
			StatusLabel: "Running",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	status, err := c.GetVpsStatus(context.Background(), "vps-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "running" {
		t.Errorf("Status = %v, want running", status)
	}
}

func TestClient_ListVpsImages(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vps/images" {
			t.Errorf("Path = %v, want /vps/images", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listImagesResponse{
			Images: []VpsImage{
				{
					ID:          "ubuntu-22.04",
					Image:       "ubuntu-22.04",
					Label:       "Ubuntu 22.04 LTS",
					Distro:      "ubuntu",
					Version:     "22.04",
					DefaultUser: "ubuntu",
				},
				{
					ID:          "debian-12",
					Image:       "debian-12",
					Label:       "Debian 12",
					Distro:      "debian",
					Version:     float64(12),
					DefaultUser: "debian",
				},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	images, err := c.ListVpsImages(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("got %d images, want 2", len(images))
	}
	if images[0].ID != "ubuntu-22.04" {
		t.Errorf("images[0].ID = %v, want ubuntu-22.04", images[0].ID)
	}
}

func TestVpsImage_GetVersion(t *testing.T) {
	tests := []struct {
		name    string
		version interface{}
		want    string
	}{
		{
			name:    "string version",
			version: "22.04",
			want:    "22.04",
		},
		{
			name:    "integer version as float64",
			version: float64(12),
			want:    "12",
		},
		{
			name:    "decimal version",
			version: float64(10.5),
			want:    "10.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &VpsImage{Version: tt.version}
			got := img.GetVersion()
			if got != tt.want {
				t.Errorf("GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

package client

import (
	"context"
	"encoding/json"
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

		var req CreateVpsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "test-vps" {
			t.Errorf("Name = %v, want test-vps", req.Name)
		}
		if req.Image != "ubuntu-22.04" {
			t.Errorf("Image = %v, want ubuntu-22.04", req.Image)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createVpsResponse{
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
	sshKeyID := "key-123"
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(showVpsResponse{
			Instance: VpsInstance{
				ID:                "vps-123",
				Name:              "test-vps",
				Status:            "running",
				CPUAllocationType: "shared",
				CPUCores:          2,
				MemorySizeGB:      4,
				StorageSizeGB:     50,
				PublicIP:          &publicIP,
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
}

func TestClient_GetVps_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "VPS not found"}`))
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

		var req UpdateVpsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.CPUCores == nil || *req.CPUCores != 4 {
			t.Errorf("CPUCores = %v, want 4", req.CPUCores)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(showVpsResponse{
			Instance: VpsInstance{
				ID:       "vps-123",
				Name:     "test-vps",
				Status:   "running",
				CPUCores: 4,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	cpuCores := 4
	vps, err := c.UpdateVps(context.Background(), "vps-123", UpdateVpsRequest{
		CPUCores: &cpuCores,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vps.CPUCores != 4 {
		t.Errorf("CPUCores = %v, want 4", vps.CPUCores)
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
		w.Write([]byte(`{"message": "VPS starting"}`))
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
		json.NewEncoder(w).Encode(statusVpsResponse{
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
		json.NewEncoder(w).Encode(listImagesResponse{
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

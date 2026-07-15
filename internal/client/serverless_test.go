package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateServerless(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/serverless" {
			t.Errorf("Path = %v, want /serverless", r.URL.Path)
		}

		var req CreateServerlessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Name != "my-app" {
			t.Errorf("Name = %v, want my-app", req.Name)
		}
		if req.DeploymentType != "docker_image" {
			t.Errorf("DeploymentType = %v, want docker_image", req.DeploymentType)
		}
		if req.Image != "nginx" {
			t.Errorf("Image = %v, want nginx", req.Image)
		}
		if req.ImageTag != "latest" {
			t.Errorf("ImageTag = %v, want latest", req.ImageTag)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createServerlessResponse{
			Message: "Container created",
			Container: ServerlessContainer{
				ID:             "srv-123",
				Name:           "my-app",
				Status:         "pending",
				DeploymentType: "docker_image",
				Image:          strPtr("nginx"),
				ImageTag:       "latest",
				Port:           80,
				MinScale:       0,
				MaxScale:       10,
				URL:            "https://my-app.serverless.danubedata.ro",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.CreateServerless(context.Background(), CreateServerlessRequest{
		Name:            "my-app",
		DeploymentType:  "docker_image",
		ResourceProfile: "small",
		Image:           "nginx",
		ImageTag:        "latest",
		Port:            80,
		MinScale:        0,
		MaxScale:        10,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.ID != "srv-123" {
		t.Errorf("ID = %v, want srv-123", container.ID)
	}
	if container.Name != "my-app" {
		t.Errorf("Name = %v, want my-app", container.Name)
	}
	if container.DeploymentType != "docker_image" {
		t.Errorf("DeploymentType = %v, want docker_image", container.DeploymentType)
	}
}

func TestClient_GetServerless(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/serverless/srv-123" {
			t.Errorf("Path = %v, want /serverless/srv-123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showServerlessResponse{
			Container: ServerlessContainer{
				ID:             "srv-123",
				Name:           "my-app",
				Status:         "running",
				DeploymentType: "docker_image",
				Image:          strPtr("nginx"),
				ImageTag:       "latest",
				Port:           80,
				MinScale:       0,
				MaxScale:       10,
				EnvironmentVariables: map[string]string{
					"ENV": "production",
				},
			},
			URL:         "https://my-app.serverless.danubedata.ro",
			MonthlyCost: 4.99,
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.GetServerless(context.Background(), "srv-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.ID != "srv-123" {
		t.Errorf("ID = %v, want srv-123", container.ID)
	}
	if container.Status != "running" {
		t.Errorf("Status = %v, want running", container.Status)
	}
	if container.URL != "https://my-app.serverless.danubedata.ro" {
		t.Errorf("URL = %v, want https://my-app.serverless.danubedata.ro", container.URL)
	}
	if container.MonthlyCost != 4.99 {
		t.Errorf("MonthlyCost = %v, want 4.99", container.MonthlyCost)
	}
	if container.EnvironmentVariables["ENV"] != "production" {
		t.Errorf("EnvironmentVariables[ENV] = %v, want production", container.EnvironmentVariables["ENV"])
	}
}

// TestClient_GetServerless_IgnoresSensitiveFields verifies that even if the
// API were to include webhook_secret / git_credentials_encrypted / api_key
// in a response body (they should not - the model hides them), the client
// has nowhere to decode them into: ServerlessContainer declares no such
// fields, so they can never reach Terraform state.
func TestClient_GetServerless_IgnoresSensitiveFields(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"container": {
				"id": "srv-123",
				"name": "my-app",
				"status": "running",
				"deployment_type": "git_repository",
				"webhook_secret": "should-never-surface",
				"git_credentials_encrypted": "should-never-surface",
				"api_key": "should-never-surface"
			},
			"url": "https://my-app.serverless.danubedata.ro",
			"monthly_cost": 0
		}`))
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.GetServerless(context.Background(), "srv-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.ID != "srv-123" {
		t.Errorf("ID = %v, want srv-123", container.ID)
	}
}

func TestClient_GetServerless_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Serverless container not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetServerless(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ListServerless(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/serverless" {
			t.Errorf("Path = %v, want /serverless", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listServerlessResponse{
			Data: []ServerlessContainer{
				{
					ID:     "srv-1",
					Name:   "app-one",
					Status: "running",
				},
				{
					ID:     "srv-2",
					Name:   "app-two",
					Status: "running",
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
	containers, err := c.ListServerless(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("got %d containers, want 2", len(containers))
	}
	if containers[0].ID != "srv-1" {
		t.Errorf("containers[0].ID = %v, want srv-1", containers[0].ID)
	}
}

func TestClient_UpdateServerless(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/serverless/srv-123" {
			t.Errorf("Path = %v, want /serverless/srv-123", r.URL.Path)
		}

		var req UpdateServerlessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.MaxScale == nil || *req.MaxScale != 20 {
			t.Errorf("MaxScale = %v, want 20", req.MaxScale)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showServerlessResponse{
			Container: ServerlessContainer{
				ID:       "srv-123",
				Name:     "my-app",
				MaxScale: 20,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	maxScale := 20
	container, err := c.UpdateServerless(context.Background(), "srv-123", UpdateServerlessRequest{
		MaxScale: &maxScale,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.MaxScale != 20 {
		t.Errorf("MaxScale = %v, want 20", container.MaxScale)
	}
}

func TestClient_DeleteServerless(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/serverless/srv-123" {
			t.Errorf("Path = %v, want /serverless/srv-123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteServerless(context.Background(), "srv-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetServerlessStatus(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showServerlessResponse{
			Container: ServerlessContainer{
				ID:     "srv-123",
				Name:   "my-app",
				Status: "running",
			},
			URL: "https://my-app.serverless.danubedata.ro",
		})
	})
	defer server.Close()

	c := newTestClient(server)
	status, err := c.GetServerlessStatus(context.Background(), "srv-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != "running" {
		t.Errorf("Status = %v, want running", status)
	}
}

func TestClient_CreateServerless_WithGit(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateServerlessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.DeploymentType != "git_repository" {
			t.Errorf("DeploymentType = %v, want git_repository", req.DeploymentType)
		}
		if req.RepositoryURL != "https://github.com/user/repo" {
			t.Errorf("RepositoryURL = %v, want https://github.com/user/repo", req.RepositoryURL)
		}
		if req.RepositoryBranch != "main" {
			t.Errorf("RepositoryBranch = %v, want main", req.RepositoryBranch)
		}
		if req.SourceType != "dockerfile" {
			t.Errorf("SourceType = %v, want dockerfile", req.SourceType)
		}
		if req.GitAuthType != "none" {
			t.Errorf("GitAuthType = %v, want none", req.GitAuthType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createServerlessResponse{
			Message: "Container created",
			Container: ServerlessContainer{
				ID:               "srv-124",
				Name:             "git-app",
				Status:           "building",
				DeploymentType:   "git_repository",
				SourceType:       strPtr("dockerfile"),
				RepositoryURL:    strPtr("https://github.com/user/repo"),
				RepositoryBranch: "main",
				GitAuthType:      "none",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.CreateServerless(context.Background(), CreateServerlessRequest{
		Name:             "git-app",
		DeploymentType:   "git_repository",
		RepositoryURL:    "https://github.com/user/repo",
		RepositoryBranch: "main",
		SourceType:       "dockerfile",
		GitAuthType:      "none",
		Port:             8080,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.DeploymentType != "git_repository" {
		t.Errorf("DeploymentType = %v, want git_repository", container.DeploymentType)
	}
	if container.RepositoryURL == nil || *container.RepositoryURL != "https://github.com/user/repo" {
		t.Errorf("RepositoryURL = %v, want https://github.com/user/repo", container.RepositoryURL)
	}
}

func TestClient_CreateServerless_WithEnvVars(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateServerlessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.EnvironmentVariables) != 2 {
			t.Errorf("EnvironmentVariables count = %v, want 2", len(req.EnvironmentVariables))
		}
		if req.EnvironmentVariables["DB_HOST"] != "localhost" {
			t.Errorf("EnvironmentVariables[DB_HOST] = %v, want localhost", req.EnvironmentVariables["DB_HOST"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createServerlessResponse{
			Message: "Container created",
			Container: ServerlessContainer{
				ID:     "srv-125",
				Name:   "env-app",
				Status: "pending",
				EnvironmentVariables: map[string]string{
					"DB_HOST": "localhost",
					"DB_PORT": "5432",
				},
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.CreateServerless(context.Background(), CreateServerlessRequest{
		Name:           "env-app",
		DeploymentType: "docker_image",
		Image:          "myapp",
		ImageTag:       "latest",
		EnvironmentVariables: map[string]string{
			"DB_HOST": "localhost",
			"DB_PORT": "5432",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(container.EnvironmentVariables) != 2 {
		t.Errorf("EnvironmentVariables count = %v, want 2", len(container.EnvironmentVariables))
	}
}

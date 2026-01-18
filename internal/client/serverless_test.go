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
		if req.DeploymentType != "docker" {
			t.Errorf("DeploymentType = %v, want docker", req.DeploymentType)
		}
		if req.ImageURL != "nginx:latest" {
			t.Errorf("ImageURL = %v, want nginx:latest", req.ImageURL)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createServerlessResponse{
			Message: "Container created",
			Container: ServerlessContainer{
				ID:             "srv-123",
				Name:           "my-app",
				Status:         "creating",
				DeploymentType: "docker",
				ImageURL:       "nginx:latest",
				Port:           80,
				MinInstances:   0,
				MaxInstances:   10,
				URL:            "https://my-app.serverless.danubedata.com",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.CreateServerless(context.Background(), CreateServerlessRequest{
		Name:            "my-app",
		ResourceProfile: "small",
		DeploymentType:  "docker",
		ImageURL:        "nginx:latest",
		Port:            80,
		MinInstances:    0,
		MaxInstances:    10,
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
	if container.DeploymentType != "docker" {
		t.Errorf("DeploymentType = %v, want docker", container.DeploymentType)
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
				DeploymentType: "docker",
				ImageURL:       "nginx:latest",
				Port:           80,
				MinInstances:   0,
				MaxInstances:   10,
				EnvironmentVariables: map[string]string{
					"ENV": "production",
				},
			},
			URL:         "https://my-app.serverless.danubedata.com",
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
	if container.URL != "https://my-app.serverless.danubedata.com" {
		t.Errorf("URL = %v, want https://my-app.serverless.danubedata.com", container.URL)
	}
	if container.EnvironmentVariables["ENV"] != "production" {
		t.Errorf("EnvironmentVariables[ENV] = %v, want production", container.EnvironmentVariables["ENV"])
	}
}

func TestClient_GetServerless_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Serverless container not found"}`))
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
		if req.MaxInstances == nil || *req.MaxInstances != 20 {
			t.Errorf("MaxInstances = %v, want 20", req.MaxInstances)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(showServerlessResponse{
			Container: ServerlessContainer{
				ID:           "srv-123",
				Name:         "my-app",
				MaxInstances: 20,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	maxInstances := 20
	container, err := c.UpdateServerless(context.Background(), "srv-123", UpdateServerlessRequest{
		MaxInstances: &maxInstances,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.MaxInstances != 20 {
		t.Errorf("MaxInstances = %v, want 20", container.MaxInstances)
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
			URL: "https://my-app.serverless.danubedata.com",
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

		if req.DeploymentType != "git" {
			t.Errorf("DeploymentType = %v, want git", req.DeploymentType)
		}
		if req.GitRepository != "https://github.com/user/repo" {
			t.Errorf("GitRepository = %v, want https://github.com/user/repo", req.GitRepository)
		}
		if req.GitBranch != "main" {
			t.Errorf("GitBranch = %v, want main", req.GitBranch)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createServerlessResponse{
			Message: "Container created",
			Container: ServerlessContainer{
				ID:             "srv-124",
				Name:           "git-app",
				Status:         "building",
				DeploymentType: "git",
				GitRepository:  "https://github.com/user/repo",
				GitBranch:      "main",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	container, err := c.CreateServerless(context.Background(), CreateServerlessRequest{
		Name:           "git-app",
		DeploymentType: "git",
		GitRepository:  "https://github.com/user/repo",
		GitBranch:      "main",
		Port:           8080,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if container.DeploymentType != "git" {
		t.Errorf("DeploymentType = %v, want git", container.DeploymentType)
	}
	if container.GitRepository != "https://github.com/user/repo" {
		t.Errorf("GitRepository = %v, want https://github.com/user/repo", container.GitRepository)
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
				Status: "creating",
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
		DeploymentType: "docker",
		ImageURL:       "myapp:latest",
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

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_CreateParameterGroup(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.URL.Path != "/parameter-groups" {
			t.Errorf("Path = %v, want /parameter-groups", r.URL.Path)
		}

		var req CreateParameterGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Name != "redis-prod" {
			t.Errorf("Name = %v, want redis-prod", req.Name)
		}
		if req.Type != "cache" {
			t.Errorf("Type = %v, want cache", req.Type)
		}
		if req.ProviderType != "redis" {
			t.Errorf("ProviderType = %v, want redis", req.ProviderType)
		}
		if req.Parameters["maxmemory-policy"] != "allkeys-lru" {
			t.Errorf("Parameters[maxmemory-policy] = %v, want allkeys-lru", req.Parameters["maxmemory-policy"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(parameterGroupResponse{
			Message: "Parameter group created",
			ParameterGroup: ParameterGroup{
				ID:           42,
				Name:         "redis-prod",
				Type:         "cache",
				ProviderType: "redis",
				Parameters: map[string]interface{}{
					"maxmemory-policy": "allkeys-lru",
				},
				LockedParameters: []string{},
				IsDefault:        false,
				IsActive:         true,
				IsSystem:         false,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	pg, err := c.CreateParameterGroup(context.Background(), CreateParameterGroupRequest{
		Name:         "redis-prod",
		Type:         "cache",
		ProviderType: "redis",
		Parameters: map[string]interface{}{
			"maxmemory-policy": "allkeys-lru",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pg.ID != 42 {
		t.Errorf("ID = %v, want 42", pg.ID)
	}
	if pg.Name != "redis-prod" {
		t.Errorf("Name = %v, want redis-prod", pg.Name)
	}
	if pg.ProviderType != "redis" {
		t.Errorf("ProviderType = %v, want redis", pg.ProviderType)
	}
}

func TestClient_GetParameterGroup(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/parameter-groups/42" {
			t.Errorf("Path = %v, want /parameter-groups/42", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(parameterGroupResponse{
			ParameterGroup: ParameterGroup{
				ID:           42,
				Name:         "redis-prod",
				Type:         "cache",
				ProviderType: "redis",
				Parameters: map[string]interface{}{
					"maxmemory-policy": "allkeys-lru",
				},
				LockedParameters: []string{"maxmemory"},
				IsSystem:         false,
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	pg, err := c.GetParameterGroup(context.Background(), "42")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pg.ID != 42 {
		t.Errorf("ID = %v, want 42", pg.ID)
	}
	if len(pg.LockedParameters) != 1 || pg.LockedParameters[0] != "maxmemory" {
		t.Errorf("LockedParameters = %v, want [maxmemory]", pg.LockedParameters)
	}
}

func TestClient_GetParameterGroup_NotFound(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Parameter group not found"}`))
	})
	defer server.Close()

	c := newTestClient(server)
	_, err := c.GetParameterGroup(context.Background(), "999")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestClient_ListParameterGroups(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %v, want GET", r.Method)
		}
		if r.URL.Path != "/parameter-groups" {
			t.Errorf("Path = %v, want /parameter-groups", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listParameterGroupsResponse{
			Data: []ParameterGroup{
				{ID: 1, Name: "redis-default", Type: "cache", ProviderType: "redis", IsSystem: true},
				{ID: 2, Name: "mysql-default", Type: "database", ProviderType: "mysql", IsSystem: true},
			},
			Pagination: Pagination{CurrentPage: 1, LastPage: 1, PerPage: 50, Total: 2},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	groups, err := c.ListParameterGroups(context.Background(), ListParameterGroupsOptions{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[0].Name != "redis-default" {
		t.Errorf("groups[0].Name = %v, want redis-default", groups[0].Name)
	}
}

func TestClient_ListParameterGroups_WithFilters(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("type") != "cache" {
			t.Errorf("type query = %v, want cache", q.Get("type"))
		}
		if q.Get("provider_type") != "redis" {
			t.Errorf("provider_type query = %v, want redis", q.Get("provider_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listParameterGroupsResponse{
			Data:       []ParameterGroup{{ID: 1, Name: "redis-default", Type: "cache", ProviderType: "redis"}},
			Pagination: Pagination{CurrentPage: 1, LastPage: 1, PerPage: 50, Total: 1},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	groups, err := c.ListParameterGroups(context.Background(), ListParameterGroupsOptions{
		Type:         "cache",
		ProviderType: "redis",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
}

func TestClient_UpdateParameterGroup(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Method = %v, want PUT", r.Method)
		}
		if r.URL.Path != "/parameter-groups/42" {
			t.Errorf("Path = %v, want /parameter-groups/42", r.URL.Path)
		}

		var req UpdateParameterGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Name == nil || *req.Name != "redis-prod-v2" {
			t.Errorf("Name = %v, want redis-prod-v2", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(parameterGroupResponse{
			ParameterGroup: ParameterGroup{
				ID:           42,
				Name:         "redis-prod-v2",
				Type:         "cache",
				ProviderType: "redis",
			},
		})
	})
	defer server.Close()

	c := newTestClient(server)
	name := "redis-prod-v2"
	pg, err := c.UpdateParameterGroup(context.Background(), "42", UpdateParameterGroupRequest{
		Name: &name,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pg.Name != "redis-prod-v2" {
		t.Errorf("Name = %v, want redis-prod-v2", pg.Name)
	}
}

func TestClient_DeleteParameterGroup(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Method = %v, want DELETE", r.Method)
		}
		if r.URL.Path != "/parameter-groups/42" {
			t.Errorf("Path = %v, want /parameter-groups/42", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	c := newTestClient(server)
	err := c.DeleteParameterGroup(context.Background(), "42")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitDatabaseResource_GrowsStorageAfterCreate(t *testing.T) {
	created, dns, backups := false, false, false
	status, size, grows := "running", 20, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		instance := func() map[string]interface{} {
			return map[string]interface{}{
				"id": "db-1", "name": "unit-database", "status": status,
				"resource_profile": "small", "storage_size_gb": size,
				"version": "8.0", "datacenter": "fsn1", "network_tier": "standard",
				"provider":   map[string]interface{}{"name": "MySQL", "type": "mysql"},
				"created_at": "2026-01-01", "updated_at": "2026-01-01",
			}
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /database":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if _, ok := body["storage_size_gb"]; ok {
				t.Errorf("create sent storage_size_gb: %#v", body)
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
		case "GET /database/db-1":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
		case "GET /database/db-1/credentials":
			_ = json.NewEncoder(w).Encode(map[string]string{"username": "root", "password": "secret", "connection_info": "mysql://unit"})
		case "PUT /database/db-1":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if raw, ok := body["storage_size_gb"]; ok {
				if string(raw) != "40" {
					t.Errorf("storage grow target = %s, want 40", raw)
				}
				size = 40
				grows++
			}
			if raw, ok := body["automated_snapshots_enabled"]; ok {
				backups = string(raw) == "true"
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
		case "POST /database/db-1/dns":
			dns = true
		case "POST /database/db-1/stop":
			status = "stopped"
		case "DELETE /database/db-1":
			created = false
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_database" "test" {
  name = "unit-database"
  engine = "mysql"
  resource_profile = "small"
  datacenter = "fsn1"
  storage_size_gb = 40
  dns_enabled = true
  automated_snapshots_enabled = true
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps:                    []resource.TestStep{{Config: config, Check: resource.TestCheckResourceAttr("danubedata_database.test", "storage_size_gb", "40")}},
	})
	if grows != 1 || !dns || !backups || created {
		t.Fatalf("grows=%d dns=%v backups=%v created=%v, want one grow, both settings, and cleanup", grows, dns, backups, created)
	}
}

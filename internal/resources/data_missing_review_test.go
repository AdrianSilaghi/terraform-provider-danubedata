package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestUnitDataResource_RecreatesExternallyDeletedInstance(t *testing.T) {
	for _, kind := range []string{"cache", "database"} {
		t.Run(kind, func(t *testing.T) {
			created, creates := false, 0
			status := "running"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				id := kind + "-" + strconv.Itoa(creates)
				instance := map[string]interface{}{
					"id": id, "name": "unit-instance", "status": status,
					"resource_profile": "small", "version": "8.0", "datacenter": "fsn1",
					"network_tier": "standard", "provider": map[string]interface{}{"name": "Redis", "type": "redis"},
					"storage_size_gb": 20, "created_at": "2026-01-01", "updated_at": "2026-01-01",
				}
				if kind == "database" {
					instance["provider"] = map[string]interface{}{"name": "MySQL", "type": "mysql"}
				}
				switch r.Method + " " + r.URL.Path {
				case "POST /" + kind:
					creates++
					created, status = true, "running"
					instance["id"] = kind + "-" + strconv.Itoa(creates)
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance})
				case "GET /" + kind + "/" + kind + "-" + strconv.Itoa(creates-1):
					w.WriteHeader(http.StatusNotFound)
				case "GET /" + kind + "/" + id:
					if !created {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance})
				case "GET /" + kind + "/" + id + "/credentials", "GET /" + kind + "/" + id + "/connection-info":
					w.WriteHeader(http.StatusForbidden)
				case "POST /" + kind + "/" + id + "/stop":
					status = "stopped"
				case "DELETE /" + kind + "/" + id:
					created = false
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			engine := `cache_provider = "redis"`
			if kind == "database" {
				engine = `engine = "mysql"`
			}
			config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_%s" "test" {
  name = "unit-instance"
  %s
  resource_profile = "small"
  datacenter = "fsn1"
}`, server.URL, kind, engine)
			resourceName := "danubedata_" + kind + ".test"
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config, Check: func(*terraform.State) error { created = false; return nil }, ExpectNonEmptyPlan: true},
					{Config: config, Check: resource.TestCheckResourceAttr(resourceName, "id", kind+"-2")},
				},
			})
			if creates != 2 || created {
				t.Fatalf("creates=%d created=%v, want recreation and cleanup", creates, created)
			}
		})
	}
}

package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestUnitDataResource_UpdatesStoppedInstanceAndDisablesDns(t *testing.T) {
	for _, kind := range []string{"cache", "database"} {
		t.Run(kind, func(t *testing.T) {
			created, dns := false, false
			status, profile, size := "running", "small", 20
			updates, stopCalls, deniedCredentials := 0, 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				id := kind + "-1"
				instance := func() map[string]interface{} {
					result := map[string]interface{}{
						"id": id, "name": "unit-instance", "status": status,
						"resource_profile": profile, "version": "8.0",
						"datacenter": "fsn1", "network_tier": "standard",
						"provider":   map[string]interface{}{"name": "Redis", "type": "redis"},
						"created_at": "2026-01-01", "updated_at": "2026-01-01",
					}
					if kind == "database" {
						result["provider"] = map[string]interface{}{"name": "MySQL", "type": "mysql"}
						result["storage_size_gb"] = size
					}
					return result
				}
				switch r.Method + " " + r.URL.Path {
				case "POST /" + kind:
					created = true
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
				case "GET /" + kind + "/" + id:
					if !created {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
				case "GET /" + kind + "/" + id + "/credentials", "GET /" + kind + "/" + id + "/connection-info":
					deniedCredentials++
					w.WriteHeader(http.StatusForbidden)
				case "PUT /" + kind + "/" + id:
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if string(body["resource_profile"]) != `"medium"` {
						t.Errorf("profile update = %#v", body)
					}
					if kind == "database" && string(body["storage_size_gb"]) != "40" {
						t.Errorf("storage update = %#v", body)
					}
					profile, size = "medium", 40
					updates++
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance()})
				case "POST /" + kind + "/" + id + "/dns":
					dns = true
				case "DELETE /" + kind + "/" + id + "/dns":
					dns = false
				case "POST /" + kind + "/" + id + "/stop":
					stopCalls++
					status = "stopped"
				case "DELETE /" + kind + "/" + id:
					created = false
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			config := func(updated bool) string {
				profile, dnsEnabled, storage := "small", "true", ""
				if updated {
					profile, dnsEnabled = "medium", "false"
				}
				engine := `cache_provider = "redis"`
				if kind == "database" {
					engine = `engine = "mysql"`
					if updated {
						storage = "storage_size_gb = 40"
					}
				}
				return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_%s" "test" {
  name = "unit-instance"
  %s
  resource_profile = %q
  datacenter = "fsn1"
  dns_enabled = %s
  %s
}`, server.URL, kind, engine, profile, dnsEnabled, storage)
			}
			resourceName := "danubedata_" + kind + ".test"
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config(false), Check: func(*terraform.State) error { status = "stopped"; return nil }},
					{Config: config(true), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "resource_profile", "medium"),
						resource.TestCheckResourceAttr(resourceName, "dns_enabled", "false"),
						resource.TestCheckResourceAttr(resourceName, "status", "stopped"),
					)},
				},
			})
			if updates != 1 || dns || stopCalls != 0 || deniedCredentials == 0 || created {
				t.Fatalf("updates=%d dns=%v stops=%d denied_credentials=%d created=%v", updates, dns, stopCalls, deniedCredentials, created)
			}
		})
	}
}

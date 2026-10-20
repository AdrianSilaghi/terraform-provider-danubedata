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

func TestUnitDataSnapshotResource_Lifecycle(t *testing.T) {
	for _, kind := range []string{"cache", "database"} {
		t.Run(kind, func(t *testing.T) {
			created, creates := false, 0
			instanceField := kind + "_instance_id"
			sizeField := "size_gb"
			if kind == "cache" {
				sizeField = "size_mb"
			}
			snapshot := map[string]interface{}{
				"id": 17, "name": "unit-snapshot", "description": "backup",
				instanceField: "instance-1", "status": "ready", sizeField: 1.25,
				"size_bytes": 1250, "size_basis": "stored",
				"created_at": "2026-01-01", "updated_at": "2026-01-01",
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + r.URL.Path {
				case "POST /snapshots/" + kind:
					var body map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body[instanceField] != "instance-1" || body["name"] != "unit-snapshot" || body["description"] != "backup" {
						t.Errorf("create request = %#v", body)
					}
					created = true
					creates++
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"snapshot": snapshot})
				case "GET /snapshots/" + kind:
					items := []interface{}{}
					if created {
						items = append(items, snapshot)
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": items, "pagination": map[string]int{"last_page": 1}})
				case "DELETE /snapshots/" + kind + "/17":
					created = false
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			config := func(deleteTimeout string) string {
				return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_%s_snapshot" "test" {
  name = "unit-snapshot"
  description = "backup"
  %s = "instance-1"
  timeouts { delete = %q }
}`, server.URL, kind, instanceField, deleteTimeout)
			}
			resourceName := "danubedata_" + kind + "_snapshot.test"
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config("10m"), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", "17"),
						resource.TestCheckResourceAttr(resourceName, sizeField, "1.25"),
						resource.TestCheckResourceAttr(resourceName, "size_bytes", "1250"),
						resource.TestCheckResourceAttr(resourceName, "size_basis", "stored"),
					)},
					{Config: config("5m"), Check: resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "5m")},
					{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
				},
			})
			if creates != 1 || created {
				t.Fatalf("creates=%d created=%v, want one create and cleanup", creates, created)
			}
		})
	}
}

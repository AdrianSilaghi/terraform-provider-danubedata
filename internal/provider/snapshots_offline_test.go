package provider_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOfflineSnapshotLifecycles(t *testing.T) {
	tests := []struct {
		kind, parentField, sizeField, size, parentID string
		id, sizeBytes                                int
	}{
		{"vps", "vps_instance_id", "size_gb", "1.25", "vps-uuid", 101, 1342177280},
		{"cache", "cache_instance_id", "size_mb", "32.5", "cache-uuid", 201, 34078720},
		{"database", "database_instance_id", "size_gb", "0.5", "database-uuid", 301, 536870912},
	}
	for _, tc := range tests {
		t.Run(tc.kind, func(t *testing.T) {
			var mu sync.Mutex
			created, nullSize, lists, deletes := false, false, 0, 0
			path := "/snapshots/" + tc.kind
			name := tc.kind + "-snapshot"
			ready := map[string]any{
				"id": tc.id, "name": name, "description": "offline snapshot", "status": "ready",
				"size_bytes": tc.sizeBytes, "size_basis": "stored", tc.sizeField: tc.size,
				tc.parentField: tc.parentID, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
				tc.kind + "_instance": map[string]any{"id": tc.parentID, "name": "parent"},
			}
			if tc.kind == "vps" {
				ready["description"] = nil
				ready["expires_at"] = nil
				ready["vps_instance"] = map[string]any{"id": tc.parentID, "name": "parent", "deleted_at": nil}
			}
			if tc.kind == "database" {
				ready["database_instance"] = map[string]any{"id": tc.parentID, "name": "parent", "deleted": false}
				ready["kept_status"] = nil
				ready["kept_size_bytes"] = nil
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.Header.Get("Authorization") != "Bearer offline-token" || r.Header.Get("X-Team-Id") != "17" {
					t.Errorf("unexpected snapshot API headers: authorization=%q team=%q", r.Header.Get("Authorization"), r.Header.Get("X-Team-Id"))
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == path:
					if created {
						t.Errorf("snapshot %s was created twice", tc.kind)
						w.WriteHeader(http.StatusConflict)
						return
					}
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["name"] != name || body[tc.parentField] != tc.parentID || (tc.kind == "vps" && body["description"] != nil) || (tc.kind != "vps" && body["description"] != "offline snapshot") {
						t.Errorf("unexpected %s snapshot request: %+v, %v", tc.kind, body, err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					created = true
					pending := maps.Clone(ready)
					pending["status"] = "creating"
					pending[tc.sizeField] = nil
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]any{"message": "Snapshot creation initiated", "snapshot": pending})
				case r.Method == http.MethodGet && r.URL.Path == path && r.URL.Query().Get("page") == "1":
					lists++
					data := []any{}
					if created {
						item := maps.Clone(ready)
						if nullSize {
							item[tc.sizeField] = nil
						}
						data = append(data, item)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data":       data,
						"pagination": map[string]int{"current_page": 1, "last_page": 1, "per_page": 50, "total": len(data)},
					})
				case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf("%s/%d", path, tc.id):
					deletes++
					created = false
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected snapshot API request %s %s", r.Method, r.URL.String())
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			description := `description = "offline snapshot"`
			if tc.kind == "vps" {
				description = ""
			}
			config := fmt.Sprintf(`
provider "danubedata" {
  base_url  = %q
  api_token = "offline-token"
  team_id   = 17
}
resource "danubedata_%s_snapshot" "test" {
  name        = %q
  %s
  %s = %q
  timeouts {
    create = "30s"
    delete = "30s"
  }
}
`, server.URL, tc.kind, name, description, tc.parentField, tc.parentID)
			updatedConfig := strings.Replace(config, `create = "30s"`, `create = "45s"`, 1)
			resourceName := "danubedata_" + tc.kind + "_snapshot.test"
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
				},
				Steps: []resource.TestStep{
					{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprint(tc.id)),
						resource.TestCheckResourceAttr(resourceName, "status", "ready"),
						resource.TestCheckResourceAttr(resourceName, tc.sizeField, tc.size),
					)},
					{Config: updatedConfig, Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprint(tc.id)),
						resource.TestCheckResourceAttr(resourceName, "timeouts.create", "45s"),
					)},
					{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"timeouts"}},
					{RefreshState: true, PreConfig: func() {
						mu.Lock()
						nullSize = true
						mu.Unlock()
					}, Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprint(tc.id)),
						resource.TestCheckNoResourceAttr(resourceName, tc.sizeField),
					)},
				},
			})
			mu.Lock()
			defer mu.Unlock()
			if created || lists < 2 || deletes != 1 {
				t.Errorf("incomplete %s snapshot lifecycle: created=%v lists=%d deletes=%d", tc.kind, created, lists, deletes)
			}
		})
	}
}

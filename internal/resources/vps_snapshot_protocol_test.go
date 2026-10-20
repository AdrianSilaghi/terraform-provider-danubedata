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

func TestUnitVpsSnapshotResourceLifecycle(t *testing.T) {
	exists := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snapshot := map[string]any{"id": 42, "name": "unit-snapshot", "description": nil, "status": "ready", "vps_instance_id": "vps-1", "size_gb": "1.25", "size_bytes": 1342177280, "size_basis": "stored", "expires_at": nil, "created_at": "2026-01-01", "updated_at": "2026-01-01"}
		switch r.Method + " " + r.URL.Path {
		case "POST /snapshots/vps":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["vps_instance_id"] != "vps-1" {
				t.Errorf("snapshot create: %#v", body)
			}
			exists = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"snapshot": snapshot})
		case "GET /snapshots/vps":
			data := []map[string]any{}
			if exists {
				data = append(data, snapshot)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": data, "pagination": map[string]int{"current_page": 1, "last_page": 1, "per_page": 20, "total": len(data)}})
		case "DELETE /snapshots/vps/42":
			exists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
 base_url = %q
 api_token = "test-token"
}
resource "danubedata_vps_snapshot" "test" {
 name = "unit-snapshot"
 vps_instance_id = "vps-1"
}`, server.URL)
	resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr("danubedata_vps_snapshot.test", "size_bytes", "1342177280")},
		{ResourceName: "danubedata_vps_snapshot.test", ImportState: true, ImportStateVerify: true},
	}})
}

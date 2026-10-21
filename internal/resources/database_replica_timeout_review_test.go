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

func TestUnitDatabaseReplicaResource_TimeoutChangePersists(t *testing.T) {
	created, creates := false, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		replica := map[string]interface{}{
			"name": "replica-1", "node_id": "node-1", "replica_index": 1,
			"status": "running", "ready": true, "is_replication_healthy": true,
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /database/db-1/replicas":
			creates++
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"replicas": []interface{}{replica}})
		case "GET /database/db-1/replicas":
			items := []interface{}{}
			if created {
				items = append(items, replica)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"replicas": items})
		case "DELETE /database/db-1/replicas/1":
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
resource "danubedata_database_replica" "test" {
  database_instance_id = "db-1"
  timeouts { delete = %q }
}`, server.URL, deleteTimeout)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("10m")},
			{Config: config("5m"), Check: resource.TestCheckResourceAttr("danubedata_database_replica.test", "timeouts.delete", "5m")},
		},
	})
	if creates != 1 {
		t.Fatalf("replica created %d times, want once", creates)
	}
}

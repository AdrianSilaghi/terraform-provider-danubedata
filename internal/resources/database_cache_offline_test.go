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

func TestUnitDatabaseCacheResource_ConfigurationLifecycle(t *testing.T) {
	for _, kind := range []string{"database", "cache"} {
		t.Run(kind, func(t *testing.T) {
			created, status, dns := false, "running", false
			updateCount := 0
			policy := map[string]interface{}{"enabled": false, "threshold_percent": 80, "increment_gib": 5}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				id := kind + "-1"
				instance := map[string]interface{}{
					"id": id, "name": "unit-test", "status": status, "resource_profile": "small",
					"version": "8.0", "datacenter": "fsn1", "network_tier": "standard",
					"provider":        map[string]interface{}{"name": "Redis", "type": "redis"},
					"engine":          map[string]interface{}{"name": "mysql"},
					"storage_size_gb": 20, "created_at": "2026-01-01", "updated_at": "2026-01-01",
				}
				if kind == "database" {
					instance["provider"] = map[string]interface{}{"name": "MySQL", "type": "mysql"}
					instance["storage_autoscale_policy"] = policy
				}
				switch r.Method + " " + r.URL.Path {
				case "POST /" + kind:
					var body map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["parameter_group_id"] != float64(42) || body["network_tier"] != "standard" {
						t.Errorf("create request: %#v", body)
					}
					if _, ok := body["configuration"].(map[string]interface{}); !ok {
						t.Errorf("create configuration = %#v", body["configuration"])
					}
					if kind == "database" {
						if body["storage_autoscaling_enabled"] != true || body["storage_autoscale_threshold_percent"] != float64(80) || body["storage_autoscale_increment_gib"] != float64(5) {
							t.Errorf("create autoscale request: %#v", body)
						}
						policy["enabled"] = true
					}
					created = true
					status = "running"
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance})
				case "GET /" + kind + "/" + id:
					if !created {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance})
				case "GET /" + kind + "/" + id + "/credentials", "GET /" + kind + "/" + id + "/connection-info":
					_ = json.NewEncoder(w).Encode(map[string]string{"username": "admin", "password": "secret", "connection_info": "private://unit"})
				case "PUT /" + kind + "/" + id:
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					updateCount++
					if updateCount == 2 {
						if string(body["parameter_group_id"]) != "null" || string(body["configuration"]) != "{}" || string(body["automated_snapshots_enabled"]) != "false" {
							t.Errorf("update request: %#v", body)
						}
						if kind == "database" && string(body["storage_autoscaling_enabled"]) != "false" {
							t.Errorf("update autoscale request: %#v", body)
						}
					}
					if kind == "database" {
						if raw, ok := body["storage_autoscaling_enabled"]; ok {
							policy["enabled"] = string(raw) == "true"
						}
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"instance": instance})
				case "POST /" + kind + "/" + id + "/dns":
					dns = true
				case "POST /" + kind + "/" + id + "/stop":
					status = "stopped"
				case "DELETE /" + kind + "/" + id:
					created = false
				case "DELETE /" + kind + "/" + id + "/dns":
					dns = false
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			config := func(clear bool) string {
				fields := "parameter_group_id = \"42\"\n  configuration = { sample = \"value\" }\n  automated_snapshots_enabled = true"
				if clear {
					fields = "configuration = {}\n  automated_snapshots_enabled = false\n  dns_enabled = true"
				}
				engine := "engine = \"mysql\""
				if kind == "cache" {
					engine = "cache_provider = \"redis\""
				} else {
					fields += "\n  storage_autoscaling_enabled = true\n  storage_autoscale_threshold_percent = 80\n  storage_autoscale_increment_gib = 5"
					if clear {
						fields = "configuration = {}\n  automated_snapshots_enabled = false\n  dns_enabled = true\n  storage_autoscaling_enabled = false\n  storage_autoscale_threshold_percent = 80\n  storage_autoscale_increment_gib = 5"
					}
				}
				return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}

resource "danubedata_%s" "test" {
  name = "unit-test"
  %s
  resource_profile = "small"
  datacenter = "fsn1"
  network_tier = "standard"
  %s
}`, server.URL, kind, engine, fields)
			}
			name := "danubedata_" + kind + ".test"
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: config(false), Check: resource.TestCheckResourceAttr(name, "parameter_group_id", "42")},
					{Config: config(true), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckNoResourceAttr(name, "parameter_group_id"),
						resource.TestCheckResourceAttr(name, "automated_snapshots_enabled", "false"),
						resource.TestCheckResourceAttr(name, "dns_enabled", "true"),
						resource.TestCheckResourceAttr(name, "network_tier", "standard"),
					)},
					{ResourceName: name, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"configuration", "automated_snapshots_enabled", "dns_enabled"}},
				},
			})
			if !dns || updateCount < 2 {
				t.Errorf("dns=%v updates=%d, want DNS and both updates", dns, updateCount)
			}
		})
	}
}

func TestUnitDatabaseReplicaResource_Lifecycle(t *testing.T) {
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		replica := map[string]interface{}{
			"name": "replica-1", "node_id": "node-1", "replica_index": 1,
			"endpoint": "replica.internal", "status": "running", "ready": true,
			"replication_status": "healthy", "seconds_behind_master": 0,
			"is_replication_healthy": true,
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /database/db-1/replicas":
			var body map[string]int
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["replica_count"] != 1 {
				t.Errorf("replica_count = %d", body["replica_count"])
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"replicas": []interface{}{replica}})
		case "GET /database/db-1/replicas":
			items := []interface{}{}
			if created {
				items = append(items, replica)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"replicas": items, "master": map[string]interface{}{}})
		case "DELETE /database/db-1/replicas/1":
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
resource "danubedata_database_replica" "test" {
  database_instance_id = "db-1"
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_database_replica.test", "id", "db-1:1"),
				resource.TestCheckResourceAttr("danubedata_database_replica.test", "ready", "true"),
			)},
			{ResourceName: "danubedata_database_replica.test", ImportState: true, ImportStateVerify: true},
		},
	})
	if created {
		t.Fatal("replica not deleted on test cleanup")
	}
}

func TestUnitParameterGroupResource_ClearsSettings(t *testing.T) {
	created := false
	parameters := map[string]interface{}{"maxmemory-policy": "allkeys-lru"}
	locked := []string{"maxmemory-policy"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		group := map[string]interface{}{
			"id": 1, "name": "unit-group", "type": "cache", "provider_type": "redis",
			"parameters": parameters, "locked_parameters": locked,
			"is_default": false, "is_active": true, "is_system": false,
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /parameter-groups":
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"parameter_group": group})
		case "GET /parameter-groups/1":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"parameter_group": group})
		case "PUT /parameter-groups/1":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if string(body["parameters"]) != "{}" || string(body["locked_parameters"]) != "[]" {
				t.Errorf("clear request: %#v", body)
			}
			parameters, locked = map[string]interface{}{}, []string{}
			group["parameters"], group["locked_parameters"] = []string{}, locked
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"parameter_group": group})
		case "DELETE /parameter-groups/1":
			created = false
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := func(clear bool) string {
		settings := `parameters = { "maxmemory-policy" = "allkeys-lru" }
  locked_parameters = ["maxmemory-policy"]`
		if clear {
			settings = "parameters = {}\n  locked_parameters = []"
		}
		return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_parameter_group" "test" {
  name = "unit-group"
  type = "cache"
  provider_type = "redis"
  %s
}`, server.URL, settings)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(false)},
			{Config: config(true), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.%", "0"),
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "locked_parameters.#", "0"),
			)},
			{ResourceName: "danubedata_parameter_group.test", ImportState: true, ImportStateVerify: true},
		},
	})
}
